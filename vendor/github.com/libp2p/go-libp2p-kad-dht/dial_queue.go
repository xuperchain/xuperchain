package dht

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	queue "github.com/libp2p/go-libp2p-peerstore/queue"
)

const (
	// DefaultDialQueueMinParallelism is the default value for the minimum number of worker dial goroutines that will
	// be alive at any time.
	DefaultDialQueueMinParallelism = 6
	// DefaultDialQueueMaxParallelism is the default value for the maximum number of worker dial goroutines that can
	// be alive at any time.
	DefaultDialQueueMaxParallelism = 20
	// DefaultDialQueueMaxIdle is the default value for the period that a worker dial goroutine waits before signalling
	// a worker pool downscaling.
	DefaultDialQueueMaxIdle = 5 * time.Second
	// DefaultDialQueueScalingMutePeriod is the default value for the amount of time to ignore further worker pool
	// scaling events, after one is processed. Its role is to reduce jitter.
	DefaultDialQueueScalingMutePeriod = 1 * time.Second
	// DefaultDialQueueScalingFactor is the default factor by which the current number of workers will be multiplied
	// or divided when upscaling and downscaling events occur, respectively.
	DefaultDialQueueScalingFactor = 1.5
)

type dialQueue struct {
	*dqParams

	nWorkers  uint
	out       *queue.ChanQueue
	startOnce sync.Once

	waitingCh chan waitingCh
	dieCh     chan struct{}
	growCh    chan struct{}
	shrinkCh  chan struct{}
}

type dqParams struct {
	ctx    context.Context
	target string
	dialFn func(context.Context, peer.ID) error
	in     *queue.ChanQueue
	config dqConfig
}

type dqConfig struct {
	// minParallelism is the minimum number of worker dial goroutines that will be alive at any time.
	minParallelism uint
	// maxParallelism is the maximum number of worker dial goroutines that can be alive at any time.
	maxParallelism uint
	// scalingFactor is the factor by which the current number of workers will be multiplied or divided when upscaling
	// and downscaling events occur, respectively.
	scalingFactor float64
	// mutePeriod is the amount of time to ignore further worker pool scaling events, after one is processed.
	// Its role is to reduce jitter.
	mutePeriod time.Duration
	// maxIdle is the period that a worker dial goroutine waits before signalling a worker pool downscaling.
	maxIdle time.Duration
}

// dqDefaultConfig returns the default configuration for dial queues. See const documentation to learn the default values.
func dqDefaultConfig() dqConfig {
	return dqConfig{
		minParallelism: DefaultDialQueueMinParallelism,
		maxParallelism: DefaultDialQueueMaxParallelism,
		scalingFactor:  DefaultDialQueueScalingFactor,
		maxIdle:        DefaultDialQueueMaxIdle,
		mutePeriod:     DefaultDialQueueScalingMutePeriod,
	}
}

type waitingCh struct {
	ch chan<- peer.ID
	ts time.Time
}

// newDialQueue returns an _unstarted_ adaptive dial queue that spawns a dynamically sized set of goroutines to
// preemptively stage dials for later handoff to the DHT protocol for RPC. It identifies backpressure on both
// ends (dial consumers and dial producers), and takes compensating action by adjusting the worker pool. To
// activate the dial queue, call Start().
//
// Why? Dialing is expensive. It's orders of magnitude slower than running an RPC on an already-established
// connection, as it requires establishing a TCP connection, multistream handshake, crypto handshake, mux handshake,
// and protocol negotiation.
//
// We start with config.minParallelism number of workers, and scale up and down based on demand and supply of
// dialled peers.
//
// The following events trigger scaling:
// - we scale up when we can't immediately return a successful dial to a new consumer.
// - we scale down when we've been idle for a while waiting for new dial attempts.
// - we scale down when we complete a dial and realise nobody was waiting for it.
//
// Dialler throttling (e.g. FD limit exceeded) is a concern, as we can easily spin up more workers to compensate, and
// end up adding fuel to the fire. Since we have no deterministic way to detect this for now, we hard-limit concurrency
// to config.maxParallelism.
func newDialQueue(params *dqParams) (*dialQueue, error) {
	dq := &dialQueue{
		dqParams:  params,
		out:       queue.NewChanQueue(params.ctx, queue.NewXORDistancePQ(params.target)),
		growCh:    make(chan struct{}, 1),
		shrinkCh:  make(chan struct{}, 1),
		waitingCh: make(chan waitingCh),
		dieCh:     make(chan struct{}, params.config.maxParallelism),
	}

	return dq, nil
}

// Start initiates action on this dial queue. It should only be called once; subsequent calls are ignored.
func (dq *dialQueue) Start() {
	dq.startOnce.Do(func() {
		go dq.control()
	})
}

func (dq *dialQueue) control() {
	var (
		dialled        <-chan peer.ID
		waiting        []waitingCh
		lastScalingEvt = time.Now()
	)

	defer func() {
		for _, w := range waiting {
			close(w.ch)
		}
		waiting = nil
	}()

	// start workers

	tgt := int(dq.dqParams.config.minParallelism)
	for i := 0; i < tgt; i++ {
		go dq.worker()
	}
	dq.nWorkers = uint(tgt)

	// control workers

	for {
		// First process any backlog of dial jobs and waiters -- making progress is the priority.
		// This block is copied below; couldn't find a more concise way of doing this.
		select {
		case <-dq.ctx.Done():
			return
		case w := <-dq.waitingCh:
			waiting = append(waiting, w)
			dialled = dq.out.DeqChan
			continue // onto the top.
		case p, ok := <-dialled:
			if !ok {
				return // we're done if the ChanQueue is closed, which happens when the context is closed.
			}
			w := waiting[0]
			logger.Debugf("delivering dialled peer to DHT; took %dms.", time.Since(w.ts)/time.Millisecond)
			w.ch <- p
			close(w.ch)
			waiting = waiting[1:]
			if len(waiting) == 0 {
				// no more waiters, so stop consuming dialled jobs.
				dialled = nil
			}
			continue // onto the top.
		default:
			// there's nothing to process, so proceed onto the main select block.
		}

		select {
		case <-dq.ctx.Done():
			return
		case w := <-dq.waitingCh:
			waiting = append(waiting, w)
			dialled = dq.out.DeqChan
		case p, ok := <-dialled:
			if !ok {
				return // we're done if the ChanQueue is closed, which happens when the context is closed.
			}
			w := waiting[0]
			logger.Debugf("delivering dialled peer to DHT; took %dms.", time.Since(w.ts)/time.Millisecond)
			w.ch <- p
			close(w.ch)
			waiting = waiting[1:]
			if len(waiting) == 0 {
				// no more waiters, so stop consuming dialled jobs.
				dialled = nil
			}
		case <-dq.growCh:
			if time.Since(lastScalingEvt) < dq.config.mutePeriod {
				continue
			}
			dq.grow()
			lastScalingEvt = time.Now()
		case <-dq.shrinkCh:
			if time.Since(lastScalingEvt) < dq.config.mutePeriod {
				continue
			}
			dq.shrink()
			lastScalingEvt = time.Now()
		}
	}
}

func (dq *dialQueue) Consume() <-chan peer.ID {
	ch := make(chan peer.ID, 1)

	select {
	case p, ok := <-dq.out.DeqChan:
		// short circuit and return a dialled peer if it's immediately available, or abort if DeqChan is closed.
		if ok {
			ch <- p
		}
		close(ch)
		return ch
	case <-dq.ctx.Done():
		// return a closed channel with no value if we're done.
		close(ch)
		return ch
	default:
	}

	// we have no finished dials to return, trigger a scale up.
	select {
	case dq.growCh <- struct{}{}:
	default:
	}

	// park the channel until a dialled peer becomes available.
	select {
	case dq.waitingCh <- waitingCh{ch, time.Now()}:
		// all good
	case <-dq.ctx.Done():
		// return a closed channel with no value if we're done.
		close(ch)
	}
	return ch
}

func (dq *dialQueue) grow() {
	// no mutex needed as this is only called from the (single-threaded) control loop.
	defer func(prev uint) {
		if prev == dq.nWorkers {
			return
		}
		logger.Debugf("grew dial worker pool: %d => %d", prev, dq.nWorkers)
	}(dq.nWorkers)

	if dq.nWorkers == dq.config.maxParallelism {
		return
	}
	// choosing not to worry about uint wrapping beyond max value.
	target := uint(math.Floor(float64(dq.nWorkers) * dq.config.scalingFactor))
	if target > dq.config.maxParallelism {
		target = dq.config.maxParallelism
	}
	for ; dq.nWorkers < target; dq.nWorkers++ {
		go dq.worker()
	}
}

func (dq *dialQueue) shrink() {
	// no mutex needed as this is only called from the (single-threaded) control loop.
	defer func(prev uint) {
		if prev == dq.nWorkers {
			return
		}
		logger.Debugf("shrunk dial worker pool: %d => %d", prev, dq.nWorkers)
	}(dq.nWorkers)

	if dq.nWorkers == dq.config.minParallelism {
		return
	}
	target := uint(math.Floor(float64(dq.nWorkers) / dq.config.scalingFactor))
	if target < dq.config.minParallelism {
		target = dq.config.minParallelism
	}
	// send as many die signals as workers we have to prune.
	for ; dq.nWorkers > target; dq.nWorkers-- {
		select {
		case dq.dieCh <- struct{}{}:
		default:
			logger.Debugf("too many die signals queued up.")
		}
	}
}

func (dq *dialQueue) worker() {
	// This idle timer tracks if the environment is slow. If we're waiting to long to acquire a peer to dial,
	// it means that the DHT query is progressing slow and we should shrink the worker pool.
	idleTimer := time.NewTimer(24 * time.Hour) // placeholder init value which will be overridden immediately.
	for {
		// trap exit signals first.
		select {
		case <-dq.ctx.Done():
			return
		case <-dq.dieCh:
			return
		default:
		}

		idleTimer.Stop()
		select {
		case <-idleTimer.C:
		default:
		}
		idleTimer.Reset(dq.config.maxIdle)

		select {
		case <-dq.dieCh:
			return
		case <-dq.ctx.Done():
			return
		case <-idleTimer.C:
			// no new dial requests during our idle period; time to scale down.
		case p, ok := <-dq.in.DeqChan:
			if !ok {
				return
			}

			t := time.Now()
			if err := dq.dialFn(dq.ctx, p); err != nil {
				logger.Debugf("discarding dialled peer because of error: %v", err)
				continue
			}
			logger.Debugf("dialling %v took %dms (as observed by the dht subsystem).", p, time.Since(t)/time.Millisecond)
			waiting := len(dq.waitingCh)

			// by the time we're done dialling, it's possible that the context is closed, in which case there will
			// be nobody listening on dq.out.EnqChan and we could block forever.
			select {
			case dq.out.EnqChan <- p:
			case <-dq.ctx.Done():
				return
			}
			if waiting > 0 {
				// we have somebody to deliver this value to, so no need to shrink.
				continue
			}
		}

		// scaling down; control only arrives here if the idle timer fires, or if there are no goroutines
		// waiting for the value we just produced.
		select {
		case dq.shrinkCh <- struct{}{}:
		default:
		}
	}
}
