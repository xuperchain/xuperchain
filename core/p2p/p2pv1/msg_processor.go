package p2pv1

import p2pPb "github.com/xuperchain/xuperchain/core/p2p/pb"

func (p *P2PServerV1) registerSubscribe() error {
	if _, err := p.Register(p.NewSubscriber(p.msgChan, p2pPb.XuperMessage_NEW_NODE, nil, "", p.log)); err != nil {
		p.log.Error("registerSubscribe error", "error", err)
		return err
	}
	return nil
}

func (p *P2PServerV1) handleMsg() {
	for {
		select {
		case msg := <-p.msgChan:
			// handle received msg
			p.log.Info("handleMsg", "logid", msg.GetHeader().GetLogid(), "msgType", msg.GetHeader().GetType())
			go p.handleReceivedMsg(msg)
		}
	}
}

func (p *P2PServerV1) handleReceivedMsg(msg *p2pPb.XuperMessage) {
	switch msg.GetHeader().GetType() {
	case p2pPb.XuperMessage_NEW_NODE:
		p.log.Info("handleReceivedMsg", "logid", msg.GetHeader().GetLogid(), "msgType", msg.GetHeader().GetType())
		if msg.GetHeader().GetFrom() == "" {
			return
		}
		for _, peer := range p.dynamicNodes {
			if peer == msg.GetHeader().GetFrom() {
				p.log.Warn("P2PServerV1 handleReceivedMsg this dynamicNodes have been added, omit")
				return
			}
		}
		p.dynamicNodes = append(p.dynamicNodes, msg.GetHeader().GetFrom())
	default:
		p.log.Info("P2PServerV1 handleReceivedMsg receive unknow msg type", "type", msg.GetHeader().GetType())
	}
}
