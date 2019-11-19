package server

import (
	"strings"
	"time"

	"github.com/moby/moby/pkg/pubsub"
	"golang.org/x/net/context"

	"github.com/xuperchain/xuperunion/pb"
)

type PubsubService struct {
	pub    *pubsub.Publisher
	txChan chan string
}

func NewPubsubService() *PubsubService {
	return &PubsubService{
		pub:    pubsub.NewPublisher(100*time.Millisecond, 10),
		txChan: make(chan string, 1000000),
	}
}

func (p *PubsubService) Start() {
	tick := time.Tick(time.Millisecond * 5000)
	for {
		select {
		case txid := <-p.txChan:
			p.pub.Publish(txid)
		case <-tick:
			p.pub.Publish("golang: hello Go")
		}
	}
}

func (p *PubsubService) Publish(ctx context.Context, arg *pb.String) (*pb.String, error) {
	p.pub.Publish(arg.GetValue())
	return &pb.String{}, nil
}

func (p *PubsubService) Subscribe(arg *pb.String, stream pb.PubsubService_SubscribeServer) error {
	ch := p.pub.SubscribeTopic(func(v interface{}) bool {
		if key, ok := v.(string); ok {
			if strings.HasPrefix(key, arg.GetValue()) {
				return true
			}
		}
		return false
	})

	for v := range ch {
		if err := stream.Send(&pb.String{Value: v.(string)}); err != nil {
			return err
		}
	}

	return nil
}
