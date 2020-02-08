package xchaincore

import (
	"github.com/xuperchain/xuperchain/core/common/events"
	cons_base "github.com/xuperchain/xuperchain/core/consensus/base"
	p2p_base "github.com/xuperchain/xuperchain/core/p2p/base"
)

// register events with handler
func (xc *XChainCore) initEvents() {
	eb := events.GetEventBus()
	subsTypes := []events.EventType{}
	// register events here
	if xc.coreConnection {
		subsTypes = append(subsTypes, events.ProposerReady)
		subsTypes = append(subsTypes, events.ProposerChanged)
	}

	// subscribe all events with handler
	failedTypes, err := eb.SubscribeMulti(subsTypes, xc.handleEvents)
	if err != nil {
		xc.log.Warn("XChainCore Subscribe events failed", "error", err, "failedTypes", failedTypes)
	}
}

// process all xchaincore events here
func (xc *XChainCore) handleEvents(em *events.EventMessage) {
	if em == nil || em.BcName != xc.bcname {
		return
	}
	xc.log.Debug("XChainCore handleEvents received events", "event", em)

	// handler different events here
	switch em.Type {
	case events.ProposerReady, events.ProposerChanged:
		xc.handleProposerChanged(em)
	}
}

// handle proposers changed events
func (xc *XChainCore) handleProposerChanged(em *events.EventMessage) {
	msg := em.Message
	var mcevent *cons_base.MinersChangedEvent
	switch msg.(type) {
	case *cons_base.MinersChangedEvent:
		mcevent = msg.(*cons_base.MinersChangedEvent)
	default:
		xc.log.Warn("handleProposerChanged received unknown event message", "msg", msg)
		return
	}

	cpi := &p2p_base.CorePeersInfo{
		Name:           mcevent.BcName,
		CurrentPeerIDs: make([]string, 0, len(mcevent.CurrentMiners)),
		NextPeerIDs:    make([]string, 0, len(mcevent.NextMiners)),
	}

	isCorePeer := false
	address := string(xc.address)
	for _, mi := range mcevent.CurrentMiners {
		if mi.Address == address {
			isCorePeer = true
			continue
		}
		cpi.CurrentPeerIDs = append(cpi.CurrentPeerIDs, mi.PeerInfo)
	}

	for _, mi := range mcevent.NextMiners {
		if mi.Address == address {
			isCorePeer = true
			continue
		}
		cpi.NextPeerIDs = append(cpi.NextPeerIDs, mi.PeerInfo)
	}

	// no action will be performed if current node is not one of core nodes
	if !isCorePeer {
		return
	}

	// number of next round peer id could not be 0
	if len(cpi.CurrentPeerIDs) == 0 || len(cpi.NextPeerIDs) == 0 {
		xc.log.Warn("handleProposerChanged received event with no current or next miners",
			"event", cpi.NextPeerIDs)
		return
	}

	err := xc.P2pSvr.SetCorePeers(cpi)
	if err != nil {
		xc.log.Warn("handleProposerChanged set core peers failed", "err", err)
	}
}
