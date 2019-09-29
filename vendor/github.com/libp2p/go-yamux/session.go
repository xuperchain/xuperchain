package yamux

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/libp2p/go-buffer-pool"
)

// Session is used to wrap a reliable ordered connection and to
// multiplex it into multiple streams.
type Session struct {
	// remoteGoAway indicates the remote side does
	// not want futher connections. Must be first for alignment.
	remoteGoAway int32

	// localGoAway indicates that we should stop
	// accepting futher connections. Must be first for alignment.
	localGoAway int32

	// nextStreamID is the next stream we should
	// send. This depends if we are a client/server.
	nextStreamID uint32

	// config holds our configuration
	config *Config

	// logger is used for our logs
	logger *log.Logger

	// conn is the underlying connection
	conn net.Conn

	// reader is a buffered reader
	reader io.Reader

	// pings is used to track inflight pings
	pings    map[uint32]chan struct{}
	pingID   uint32
	pingLock sync.Mutex

	// streams maps a stream id to a stream, and inflight has an entry
	// for any outgoing stream that has not yet been established. Both are
	// protected by streamLock.
	streams    map[uint32]*Stream
	inflight   map[uint32]struct{}
	streamLock sync.Mutex

	// synCh acts like a semaphore. It is sized to the AcceptBacklog which
	// is assumed to be symmetric between the client and server. This allows
	// the client to avoid exceeding the backlog and instead blocks the open.
	synCh chan struct{}

	// acceptCh is used to pass ready streams to the client
	acceptCh chan *Stream

	// sendCh is used to send messages
	sendCh chan []byte

	// recvDoneCh is closed when recv() exits to avoid a race
	// between stream registration and stream shutdown
	recvDoneCh chan struct{}

	// sendDoneCh is closed when send() exits to avoid a race
	// between returning from a Stream.Write and exiting from the send loop
	// (which may be reading a buffer on-load-from Stream.Write).
	sendDoneCh chan struct{}

	// client is true if we're the client and our stream IDs should be odd.
	client bool

	// shutdown is used to safely close a session
	shutdown     bool
	shutdownErr  error
	shutdownCh   chan struct{}
	shutdownLock sync.Mutex

	// keepaliveTimer is a periodic timer for keepalive messages. It's nil
	// when keepalives are disabled.
	keepaliveLock  sync.Mutex
	keepaliveTimer *time.Timer
}

const (
	stageInitial uint32 = iota
	stageFinal
)

// newSession is used to construct a new session
func newSession(config *Config, conn net.Conn, client bool, readBuf int) *Session {
	var reader io.Reader = conn
	if readBuf > 0 {
		reader = bufio.NewReaderSize(reader, readBuf)
	}
	s := &Session{
		config:     config,
		client:     client,
		logger:     log.New(config.LogOutput, "", log.LstdFlags),
		conn:       conn,
		reader:     reader,
		pings:      make(map[uint32]chan struct{}),
		streams:    make(map[uint32]*Stream),
		inflight:   make(map[uint32]struct{}),
		synCh:      make(chan struct{}, config.AcceptBacklog),
		acceptCh:   make(chan *Stream, config.AcceptBacklog),
		sendCh:     make(chan []byte, 64),
		recvDoneCh: make(chan struct{}),
		sendDoneCh: make(chan struct{}),
		shutdownCh: make(chan struct{}),
	}
	if client {
		s.nextStreamID = 1
	} else {
		s.nextStreamID = 2
	}
	if config.EnableKeepAlive {
		s.startKeepalive()
	}
	go s.recv()
	go s.send()
	return s
}

// IsClosed does a safe check to see if we have shutdown
func (s *Session) IsClosed() bool {
	select {
	case <-s.shutdownCh:
		return true
	default:
		return false
	}
}

// CloseChan returns a read-only channel which is closed as
// soon as the session is closed.
func (s *Session) CloseChan() <-chan struct{} {
	return s.shutdownCh
}

// NumStreams returns the number of currently open streams
func (s *Session) NumStreams() int {
	s.streamLock.Lock()
	num := len(s.streams)
	s.streamLock.Unlock()
	return num
}

// Open is used to create a new stream as a net.Conn
func (s *Session) Open() (net.Conn, error) {
	conn, err := s.OpenStream()
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// OpenStream is used to create a new stream
func (s *Session) OpenStream() (*Stream, error) {
	if s.IsClosed() {
		return nil, s.shutdownErr
	}
	if atomic.LoadInt32(&s.remoteGoAway) == 1 {
		return nil, ErrRemoteGoAway
	}

	// Block if we have too many inflight SYNs
	select {
	case s.synCh <- struct{}{}:
	case <-s.shutdownCh:
		return nil, s.shutdownErr
	}

GET_ID:
	// Get an ID, and check for stream exhaustion
	id := atomic.LoadUint32(&s.nextStreamID)
	if id >= math.MaxUint32-1 {
		return nil, ErrStreamsExhausted
	}
	if !atomic.CompareAndSwapUint32(&s.nextStreamID, id, id+2) {
		goto GET_ID
	}

	// Register the stream
	stream := newStream(s, id, streamInit)
	s.streamLock.Lock()
	s.streams[id] = stream
	s.inflight[id] = struct{}{}
	s.streamLock.Unlock()

	// Send the window update to create
	if err := stream.sendWindowUpdate(); err != nil {
		select {
		case <-s.synCh:
		default:
			s.logger.Printf("[ERR] yamux: aborted stream open without inflight syn semaphore")
		}
		return nil, err
	}
	return stream, nil
}

// Accept is used to block until the next available stream
// is ready to be accepted.
func (s *Session) Accept() (net.Conn, error) {
	conn, err := s.AcceptStream()
	if err != nil {
		return nil, err
	}
	return conn, err
}

// AcceptStream is used to block until the next available stream
// is ready to be accepted.
func (s *Session) AcceptStream() (*Stream, error) {
	select {
	case stream := <-s.acceptCh:
		if err := stream.sendWindowUpdate(); err != nil {
			return nil, err
		}
		return stream, nil
	case <-s.shutdownCh:
		return nil, s.shutdownErr
	}
}

// Close is used to close the session and all streams.
// Attempts to send a GoAway before closing the connection.
func (s *Session) Close() error {
	s.shutdownLock.Lock()
	defer s.shutdownLock.Unlock()

	if s.shutdown {
		return nil
	}
	s.shutdown = true
	if s.shutdownErr == nil {
		s.shutdownErr = ErrSessionShutdown
	}
	close(s.shutdownCh)
	s.conn.Close()
	s.stopKeepalive()
	<-s.recvDoneCh
	<-s.sendDoneCh

	s.streamLock.Lock()
	defer s.streamLock.Unlock()
	for _, stream := range s.streams {
		stream.forceClose()
	}
	return nil
}

// exitErr is used to handle an error that is causing the
// session to terminate.
func (s *Session) exitErr(err error) {
	s.shutdownLock.Lock()
	if s.shutdownErr == nil {
		s.shutdownErr = err
	}
	s.shutdownLock.Unlock()
	s.Close()
}

// GoAway can be used to prevent accepting further
// connections. It does not close the underlying conn.
func (s *Session) GoAway() error {
	return s.sendMsg(s.goAway(goAwayNormal), nil, nil)
}

// goAway is used to send a goAway message
func (s *Session) goAway(reason uint32) header {
	atomic.SwapInt32(&s.localGoAway, 1)
	hdr := encode(typeGoAway, 0, 0, reason)
	return hdr
}

// Ping is used to measure the RTT response time
func (s *Session) Ping() (time.Duration, error) {
	// Get a channel for the ping
	ch := make(chan struct{})

	// Get a new ping id, mark as pending
	s.pingLock.Lock()
	id := s.pingID
	s.pingID++
	s.pings[id] = ch
	s.pingLock.Unlock()

	// Send the ping request
	hdr := encode(typePing, flagSYN, 0, id)
	if err := s.sendMsg(hdr, nil, nil); err != nil {
		return 0, err
	}

	// Wait for a response
	start := time.Now()
	select {
	case <-ch:
	case <-time.After(s.config.ConnectionWriteTimeout):
		s.pingLock.Lock()
		delete(s.pings, id) // Ignore it if a response comes later.
		s.pingLock.Unlock()
		return 0, ErrTimeout
	case <-s.shutdownCh:
		return 0, s.shutdownErr
	}

	// Compute the RTT
	return time.Now().Sub(start), nil
}

// startKeepalive starts the keepalive process.
func (s *Session) startKeepalive() {
	s.keepaliveLock.Lock()
	defer s.keepaliveLock.Unlock()
	s.keepaliveTimer = time.AfterFunc(s.config.KeepAliveInterval, func() {
		s.keepaliveLock.Lock()

		if s.keepaliveTimer == nil {
			s.keepaliveLock.Unlock()
			// keepalives have been stopped.
			return
		}
		_, err := s.Ping()
		if err != nil {
			// Make sure to unlock before exiting so we don't
			// deadlock trying to shutdown keepalives.
			s.keepaliveLock.Unlock()
			s.logger.Printf("[ERR] yamux: keepalive failed: %v", err)
			s.exitErr(ErrKeepAliveTimeout)
			return
		}
		s.keepaliveTimer.Reset(s.config.KeepAliveInterval)
		s.keepaliveLock.Unlock()
	})
}

// stopKeepalive stops the keepalive process.
func (s *Session) stopKeepalive() {
	s.keepaliveLock.Lock()
	defer s.keepaliveLock.Unlock()
	if s.keepaliveTimer != nil {
		s.keepaliveTimer.Stop()
	}
}

// send sends the header and body.
func (s *Session) sendMsg(hdr header, body []byte, deadline <-chan struct{}) error {
	select {
	case <-s.shutdownCh:
		return s.shutdownErr
	default:
	}

	// duplicate as we're sending this async.
	buf := pool.Get(headerSize + len(body))
	copy(buf[:headerSize], hdr[:])
	copy(buf[headerSize:], body)

	select {
	case <-s.shutdownCh:
		pool.Put(buf)
		return s.shutdownErr
	case s.sendCh <- buf:
		return nil
	case <-deadline:
		pool.Put(buf)
		return ErrTimeout
	}
}

// send is a long running goroutine that sends data
func (s *Session) send() {
	if err := s.sendLoop(); err != nil {
		s.exitErr(err)
	}
}

func (s *Session) sendLoop() error {
	defer close(s.sendDoneCh)

	// Extend the write deadline if we've passed the halfway point. This can
	// be expensive so this ensures we only have to do this once every
	// ConnectionWriteTimeout/2 (usually 5s).
	var lastWriteDeadline time.Time
	extendWriteDeadline := func() error {
		now := time.Now()
		// If over half of the deadline has elapsed, extend it.
		if now.Add(s.config.ConnectionWriteTimeout / 2).After(lastWriteDeadline) {
			lastWriteDeadline = now.Add(s.config.ConnectionWriteTimeout)
			return s.conn.SetWriteDeadline(lastWriteDeadline)
		}
		return nil
	}

	writer := s.conn

	// FIXME: https://github.com/libp2p/go-libp2p/issues/644
	// Write coalescing is disabled for now.

	//writer := pool.Writer{W: s.conn}

	//var writeTimeout *time.Timer
	//var writeTimeoutCh <-chan time.Time
	//if s.config.WriteCoalesceDelay > 0 {
	//	writeTimeout = time.NewTimer(s.config.WriteCoalesceDelay)
	//	defer writeTimeout.Stop()

	//	writeTimeoutCh = writeTimeout.C
	//} else {
	//	ch := make(chan time.Time)
	//	close(ch)
	//	writeTimeoutCh = ch
	//}

	for {
		// yield after processing the last message, if we've shutdown.
		// s.sendCh is a buffered channel and Go doesn't guarantee select order.
		select {
		case <-s.shutdownCh:
			return nil
		default:
		}

		// Flushes at least once every 100 microseconds unless we're
		// constantly writing.
		var buf []byte
		select {
		case buf = <-s.sendCh:
		case <-s.shutdownCh:
			return nil
			//default:
			//	select {
			//	case buf = <-s.sendCh:
			//	case <-s.shutdownCh:
			//		return nil
			//	case <-writeTimeoutCh:
			//		if err := writer.Flush(); err != nil {
			//			if os.IsTimeout(err) {
			//				err = ErrConnectionWriteTimeout
			//			}
			//			return err
			//		}

			//		select {
			//		case buf = <-s.sendCh:
			//		case <-s.shutdownCh:
			//			return nil
			//		}

			//		if writeTimeout != nil {
			//			writeTimeout.Reset(s.config.WriteCoalesceDelay)
			//		}
			//	}
		}

		if err := extendWriteDeadline(); err != nil {
			pool.Put(buf)
			return err
		}

		_, err := writer.Write(buf)
		pool.Put(buf)

		if err != nil {
			if os.IsTimeout(err) {
				err = ErrConnectionWriteTimeout
			}
			return err
		}
	}
}

// recv is a long running goroutine that accepts new data
func (s *Session) recv() {
	if err := s.recvLoop(); err != nil {
		s.exitErr(err)
	}
}

// Ensure that the index of the handler (typeData/typeWindowUpdate/etc) matches the message type
var (
	handlers = []func(*Session, header) error{
		typeData:         (*Session).handleStreamMessage,
		typeWindowUpdate: (*Session).handleStreamMessage,
		typePing:         (*Session).handlePing,
		typeGoAway:       (*Session).handleGoAway,
	}
)

// recvLoop continues to receive data until a fatal error is encountered
func (s *Session) recvLoop() error {
	defer close(s.recvDoneCh)
	var hdr header
	for {
		// Read the header
		if _, err := io.ReadFull(s.reader, hdr[:]); err != nil {
			if err != io.EOF && !strings.Contains(err.Error(), "closed") && !strings.Contains(err.Error(), "reset by peer") {
				s.logger.Printf("[ERR] yamux: Failed to read header: %v", err)
			}
			return err
		}

		// Verify the version
		if hdr.Version() != protoVersion {
			s.logger.Printf("[ERR] yamux: Invalid protocol version: %d", hdr.Version())
			return ErrInvalidVersion
		}

		mt := hdr.MsgType()
		if mt < typeData || mt > typeGoAway {
			return ErrInvalidMsgType
		}

		if err := handlers[mt](s, hdr); err != nil {
			return err
		}
	}
}

// handleStreamMessage handles either a data or window update frame
func (s *Session) handleStreamMessage(hdr header) error {
	// Check for a new stream creation
	id := hdr.StreamID()
	flags := hdr.Flags()
	if flags&flagSYN == flagSYN {
		if err := s.incomingStream(id); err != nil {
			return err
		}
	}

	// Get the stream
	s.streamLock.Lock()
	stream := s.streams[id]
	s.streamLock.Unlock()

	// If we do not have a stream, likely we sent a RST
	if stream == nil {
		// Drain any data on the wire
		if hdr.MsgType() == typeData && hdr.Length() > 0 {
			s.logger.Printf("[WARN] yamux: Discarding data for stream: %d", id)
			if _, err := io.CopyN(ioutil.Discard, s.reader, int64(hdr.Length())); err != nil {
				s.logger.Printf("[ERR] yamux: Failed to discard data: %v", err)
				return nil
			}
		} else {
			s.logger.Printf("[WARN] yamux: frame for missing stream: %v", hdr)
		}
		return nil
	}

	// Check if this is a window update
	if hdr.MsgType() == typeWindowUpdate {
		if err := stream.incrSendWindow(hdr, flags); err != nil {
			if sendErr := s.sendMsg(s.goAway(goAwayProtoErr), nil, nil); sendErr != nil {
				s.logger.Printf("[WARN] yamux: failed to send go away: %v", sendErr)
			}
			return err
		}
		return nil
	}

	// Read the new data
	if err := stream.readData(hdr, flags, s.reader); err != nil {
		if sendErr := s.sendMsg(s.goAway(goAwayProtoErr), nil, nil); sendErr != nil {
			s.logger.Printf("[WARN] yamux: failed to send go away: %v", sendErr)
		}
		return err
	}
	return nil
}

// handlePing is invokde for a typePing frame
func (s *Session) handlePing(hdr header) error {
	flags := hdr.Flags()
	pingID := hdr.Length()

	// Check if this is a query, respond back in a separate context so we
	// don't interfere with the receiving thread blocking for the write.
	if flags&flagSYN == flagSYN {
		go func() {
			hdr := encode(typePing, flagACK, 0, pingID)
			if err := s.sendMsg(hdr, nil, nil); err != nil {
				s.logger.Printf("[WARN] yamux: failed to send ping reply: %v", err)
			}
		}()
		return nil
	}

	// Handle a response
	s.pingLock.Lock()
	ch := s.pings[pingID]
	if ch != nil {
		delete(s.pings, pingID)
		close(ch)
	}
	s.pingLock.Unlock()
	return nil
}

// handleGoAway is invokde for a typeGoAway frame
func (s *Session) handleGoAway(hdr header) error {
	code := hdr.Length()
	switch code {
	case goAwayNormal:
		atomic.SwapInt32(&s.remoteGoAway, 1)
	case goAwayProtoErr:
		s.logger.Printf("[ERR] yamux: received protocol error go away")
		return fmt.Errorf("yamux protocol error")
	case goAwayInternalErr:
		s.logger.Printf("[ERR] yamux: received internal error go away")
		return fmt.Errorf("remote yamux internal error")
	default:
		s.logger.Printf("[ERR] yamux: received unexpected go away")
		return fmt.Errorf("unexpected go away received")
	}
	return nil
}

// incomingStream is used to create a new incoming stream
func (s *Session) incomingStream(id uint32) error {
	if s.client != (id%2 == 0) {
		s.logger.Printf("[ERR] yamux: both endpoints are clients")
		return fmt.Errorf("both yamux endpoints are clients")
	}
	// Reject immediately if we are doing a go away
	if atomic.LoadInt32(&s.localGoAway) == 1 {
		hdr := encode(typeWindowUpdate, flagRST, id, 0)
		return s.sendMsg(hdr, nil, nil)
	}

	// Allocate a new stream
	stream := newStream(s, id, streamSYNReceived)

	s.streamLock.Lock()
	defer s.streamLock.Unlock()

	// Check if stream already exists
	if _, ok := s.streams[id]; ok {
		s.logger.Printf("[ERR] yamux: duplicate stream declared")
		if sendErr := s.sendMsg(s.goAway(goAwayProtoErr), nil, nil); sendErr != nil {
			s.logger.Printf("[WARN] yamux: failed to send go away: %v", sendErr)
		}
		return ErrDuplicateStream
	}

	// Register the stream
	s.streams[id] = stream

	// Check if we've exceeded the backlog
	select {
	case s.acceptCh <- stream:
		return nil
	default:
		// Backlog exceeded! RST the stream
		s.logger.Printf("[WARN] yamux: backlog exceeded, forcing connection reset")
		delete(s.streams, id)
		hdr := encode(typeWindowUpdate, flagRST, id, 0)
		return s.sendMsg(hdr, nil, nil)
	}
}

// closeStream is used to close a stream once both sides have
// issued a close. If there was an in-flight SYN and the stream
// was not yet established, then this will give the credit back.
func (s *Session) closeStream(id uint32) {
	s.streamLock.Lock()
	if _, ok := s.inflight[id]; ok {
		select {
		case <-s.synCh:
		default:
			s.logger.Printf("[ERR] yamux: SYN tracking out of sync")
		}
	}
	delete(s.streams, id)
	s.streamLock.Unlock()
}

// establishStream is used to mark a stream that was in the
// SYN Sent state as established.
func (s *Session) establishStream(id uint32) {
	s.streamLock.Lock()
	if _, ok := s.inflight[id]; ok {
		delete(s.inflight, id)
	} else {
		s.logger.Printf("[ERR] yamux: established stream without inflight SYN (no tracking entry)")
	}
	select {
	case <-s.synCh:
	default:
		s.logger.Printf("[ERR] yamux: established stream without inflight SYN (didn't have semaphore)")
	}
	s.streamLock.Unlock()
}
