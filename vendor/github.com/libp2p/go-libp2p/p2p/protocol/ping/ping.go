package ping

import (
	"bytes"
	"context"
	"errors"
	"io"
	"time"

	u "github.com/ipfs/go-ipfs-util"
	logging "github.com/ipfs/go-log"
	host "github.com/libp2p/go-libp2p-host"
	inet "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
)

var log = logging.Logger("ping")

const PingSize = 32

const ID = "/ipfs/ping/1.0.0"

const pingTimeout = time.Second * 60

type PingService struct {
	Host host.Host
}

func NewPingService(h host.Host) *PingService {
	ps := &PingService{h}
	h.SetStreamHandler(ID, ps.PingHandler)
	return ps
}

func (p *PingService) PingHandler(s inet.Stream) {
	buf := make([]byte, PingSize)

	errCh := make(chan error, 1)
	defer close(errCh)
	timer := time.NewTimer(pingTimeout)
	defer timer.Stop()

	go func() {
		select {
		case <-timer.C:
			log.Debug("ping timeout")
		case err, ok := <-errCh:
			if ok {
				log.Debug(err)
			} else {
				log.Error("ping loop failed without error")
			}
		}
		s.Reset()
	}()

	for {
		_, err := io.ReadFull(s, buf)
		if err != nil {
			errCh <- err
			return
		}

		_, err = s.Write(buf)
		if err != nil {
			errCh <- err
			return
		}

		timer.Reset(pingTimeout)
	}
}

func (ps *PingService) Ping(ctx context.Context, p peer.ID) (<-chan time.Duration, error) {
	return Ping(ctx, ps.Host, p)
}

func Ping(ctx context.Context, h host.Host, p peer.ID) (<-chan time.Duration, error) {
	s, err := h.NewStream(ctx, p, ID)
	if err != nil {
		return nil, err
	}

	out := make(chan time.Duration)
	go func() {
		defer close(out)
		defer s.Reset()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				t, err := ping(s)
				if err != nil {
					log.Debugf("ping error: %s", err)
					return
				}

				h.Peerstore().RecordLatency(p, t)
				select {
				case out <- t:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return out, nil
}

func ping(s inet.Stream) (time.Duration, error) {
	buf := make([]byte, PingSize)
	u.NewTimeSeededRand().Read(buf)

	before := time.Now()
	_, err := s.Write(buf)
	if err != nil {
		return 0, err
	}

	rbuf := make([]byte, PingSize)
	_, err = io.ReadFull(s, rbuf)
	if err != nil {
		return 0, err
	}

	if !bytes.Equal(buf, rbuf) {
		return 0, errors.New("ping packet was incorrect!")
	}

	return time.Since(before), nil
}
