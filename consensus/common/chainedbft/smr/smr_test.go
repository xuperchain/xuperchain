package smr

import (
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	cons_base "github.com/xuperchain/xuperunion/consensus/base"
	"github.com/xuperchain/xuperunion/consensus/common/chainedbft/config"
	"github.com/xuperchain/xuperunion/consensus/common/chainedbft/external"
	"github.com/xuperchain/xuperunion/consensus/common/chainedbft/utils"
	crypto_client "github.com/xuperchain/xuperunion/crypto/client"
	"github.com/xuperchain/xuperunion/p2pv2"
	p2p_pb "github.com/xuperchain/xuperunion/p2pv2/pb"
	"github.com/xuperchain/xuperunion/pb"
)

type user struct {
	address    string
	publicKey  string
	privateKey string
}

func MakeSmr(t *testing.T) (*Smr, error) {
	user := &user{
		address:    `dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN`,
		publicKey:  `{"Curvname":"P-256","X":74695617477160058757747208220371236837474210247114418775262229497812962582435,"Y":51348715319124770392993866417088542497927816017012182211244120852620959209571}`,
		privateKey: `{"Curvname":"P-256","X":74695617477160058757747208220371236837474210247114418775262229497812962582435,"Y":51348715319124770392993866417088542497927816017012182211244120852620959209571,"D":29079635126530934056640915735344231956621504557963207107451663058887647996601}`,
	}
	cryptoClient, err := crypto_client.CreateCryptoClient(crypto_client.CryptoTypeDefault)
	if err != nil {
		t.Error("MakeSmr CreateCryptoClient error ", err)
		return nil, err
	}
	externalCons := &external.MockExternalConsensus{}
	mockP2p := &p2pv2.MockP2pServer{}
	privateKey, _ := cryptoClient.GetEcdsaPrivateKeyFromJSON([]byte(user.privateKey))

	proposalQC := &pb.QuorumCert{
		ProposalId:  []byte("proposalQC ProposalId"),
		ProposalMsg: []byte("proposalQC ProposalMsg"),
		Type:        pb.QCState_PREPARE,
		ViewNumber:  1005,
	}

	generateQC := &pb.QuorumCert{
		ProposalId:  []byte("generateQC ProposalId"),
		ProposalMsg: []byte("generateQC ProposalMsg"),
		Type:        pb.QCState_PREPARE,
		ViewNumber:  1004,
	}

	lockedQC := &pb.QuorumCert{
		ProposalId:  []byte("lockedQC ProposalId"),
		ProposalMsg: []byte("lockedQC ProposalMsg"),
		Type:        pb.QCState_PREPARE,
		ViewNumber:  1003,
	}

	validates := []*cons_base.CandidateInfo{
		&cons_base.CandidateInfo{
			Address:  "dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN",
			PeerAddr: "",
		},
	}

	smr, err := NewSmr(
		&config.Config{},
		"xuper",
		user.address,
		user.publicKey,
		privateKey,
		validates,
		externalCons,
		cryptoClient,
		mockP2p,
		proposalQC, generateQC, lockedQC,
	)
	return smr, nil
}

func MakeProposalMsg(t *testing.T) (*p2p_pb.XuperMessage, error) {
	smr, err := MakeSmr(t)
	if err != nil {
		t.Error("MakeProposalMsg MakeSmr error", err)
		return nil, err
	}

	qc := &pb.QuorumCert{
		ProposalId:  []byte("test proposalID"),
		ProposalMsg: []byte("test proposalMsg"),
		ViewNumber:  1005,
		Type:        pb.QCState_PREPARE,
		SignInfos:   &pb.QCSignInfos{},
	}

	propMsg := &pb.ChainedBftPhaseMessage{
		Type:       pb.QCState_PREPARE,
		ViewNumber: 1005,
		ProposalQC: qc,
		Signature: &pb.SignInfo{
			Address:   `dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN`,
			PublicKey: `{"Curvname":"P-256","X":74695617477160058757747208220371236837474210247114418775262229497812962582435,"Y":51348715319124770392993866417088542497927816017012182211244120852620959209571}`,
		},
	}

	propMsg, err = utils.MakePhaseMsgSign(smr.cryptoClient, smr.privateKey, propMsg)
	if err != nil {
		t.Error("MakeProposalMsg MakePhaseMsgSign error", err)
		return nil, err
	}

	msgBuf, err := proto.Marshal(propMsg)
	if err != nil {
		t.Error("MakeProposalMsg marshal msg error", err)
		return nil, err
	}
	netMsg, _ := p2p_pb.NewXuperMessage(p2p_pb.XuperMsgVersion3, smr.bcname, "",
		p2p_pb.XuperMessage_CHAINED_BFT_NEW_PROPOSAL_MSG, msgBuf, p2p_pb.XuperMessage_NONE)
	return netMsg, nil
}

func MakeNewViewMsg(t *testing.T) (*p2p_pb.XuperMessage, error) {
	smr, err := MakeSmr(t)
	if err != nil {
		t.Error("MakeNewViewMsg MakeSmr error", err)
		return nil, err
	}
	newViewMsg := &pb.ChainedBftPhaseMessage{
		Type:       pb.QCState_NEW_VIEW,
		ViewNumber: 1007,
		Signature: &pb.SignInfo{
			Address:   `dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN`,
			PublicKey: `{"Curvname":"P-256","X":74695617477160058757747208220371236837474210247114418775262229497812962582435,"Y":51348715319124770392993866417088542497927816017012182211244120852620959209571}`,
		},
	}
	newViewMsg, err = utils.MakePhaseMsgSign(smr.cryptoClient, smr.privateKey, newViewMsg)
	if err != nil {
		t.Error("MakeNewViewMsg MakePhaseMsgSign error", err)
		return nil, err
	}
	msgBuf, err := proto.Marshal(newViewMsg)
	if err != nil {
		t.Error("MakeNewViewMsg marshal msg error", err)
		return nil, err
	}
	netMsg, _ := p2p_pb.NewXuperMessage(p2p_pb.XuperMsgVersion3, smr.bcname, "",
		p2p_pb.XuperMessage_CHAINED_BFT_NEW_VIEW_MSG, msgBuf, p2p_pb.XuperMessage_NONE)
	return netMsg, nil
}

func MakeVoteMsg(t *testing.T) (*p2p_pb.XuperMessage, error) {
	smr, err := MakeSmr(t)
	if err != nil {
		t.Error("MakeVoteMsg MakeSmr error", err)
		return nil, err
	}
	voteMsg := &pb.ChainedBftVoteMessage{
		ProposalId: []byte("proposalQC ProposalId"),
		Signature: &pb.SignInfo{
			Address:   `dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN`,
			PublicKey: `{"Curvname":"P-256","X":74695617477160058757747208220371236837474210247114418775262229497812962582435,"Y":51348715319124770392993866417088542497927816017012182211244120852620959209571}`,
		},
	}
	_, err = utils.MakeVoteMsgSign(smr.cryptoClient, smr.privateKey, voteMsg.GetSignature(), voteMsg.GetProposalId())
	if err != nil {
		t.Error("MakeVoteMsg MakeVoteMsgSign error", err)
		return nil, err
	}

	msgBuf, err := proto.Marshal(voteMsg)
	if err != nil {
		t.Error("MakeVoteMsg marshal msg error", err)
		return nil, err
	}

	netMsg, _ := p2p_pb.NewXuperMessage(p2p_pb.XuperMsgVersion3, smr.bcname, "",
		p2p_pb.XuperMessage_CHAINED_BFT_VOTE_MSG, msgBuf, p2p_pb.XuperMessage_NONE)
	return netMsg, nil
}

func TestNewSmr(t *testing.T) {
	smr, err := MakeSmr(t)
	if err != nil {
		t.Error("TestNewSmr error", err)
	}
	go func() {
		time.Sleep(1 * time.Second)
		smr.QuitCh <- true
	}()
	smr.Start()
	return
}

func TestProcessNewView(t *testing.T) {
	smr, err := MakeSmr(t)
	if err != nil {
		t.Error("TestProcessNewView MakeSmr error", err)
		return
	}
	err = smr.ProcessNewView(1005, "dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN", "dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN")
	if err != nil {
		t.Error("TestProcessNewView error", err)
	}
}

func TestProcessProposal(t *testing.T) {
	smr, err := MakeSmr(t)
	if err != nil {
		t.Error("TestProcessProposal MakeSmr error", err)
		return
	}
	_, err = smr.ProcessProposal(1005, []byte("test proposalID"), []byte("test proposalMsg"))
	if err != nil {
		t.Error("TestProcessProposal error", err)
	}
}

func TestHandleReceivedMsg(t *testing.T) {
	smr, err := MakeSmr(t)
	if err != nil {
		t.Error("TestHandleReceivedMsg MakeSmr error", err)
		return
	}
	netMsg, err := MakeProposalMsg(t)
	err = smr.handleReceivedMsg(netMsg)
	if err != nil {
		t.Error("TestHandleReceivedMsg handleReceivedMsg error", err)
		return
	}
}

func TestHandleReceivedVoteMsg(t *testing.T) {
	smr, err := MakeSmr(t)
	if err != nil {
		t.Error("TestHandleReceivedVoteMsg MakeSmr error", err)
		return
	}
	netMsg, err := MakeVoteMsg(t)
	err = smr.handleReceivedVoteMsg(netMsg)
	if err != nil {
		t.Error("TestHandleReceivedVoteMsg handleReceivedVoteMsg error", err)
	}
	if smr.votedView != 1005 {
		t.Error("TestHandleReceivedVoteMsg handleReceivedVoteMsg error", smr)
	}
}

func TestHandleReceivedNewView(t *testing.T) {
	smr, err := MakeSmr(t)
	if err != nil {
		t.Error("TestHandleReceivedNewView MakeSmr error", err)
		return
	}
	netMsg, err := MakeNewViewMsg(t)
	if err != nil {
		t.Error("TestHandleReceivedNewView MakeNewViewMsg error", err)
		return
	}
	err = smr.handleReceivedNewView(netMsg)
	if err != nil {
		t.Error("TestHandleReceivedNewView handleReceivedNewView error", err)
	}
}

func TestHandleReceivedProposal(t *testing.T) {
	smr, err := MakeSmr(t)
	if err != nil {
		t.Error("TestHandleReceivedProposal MakeSmr error", err)
		return
	}
	netMsg, err := MakeProposalMsg(t)
	if err != nil {
		t.Error("TestHandleReceivedProposal MakeProposalMsg error")
	}
	err = smr.handleReceivedProposal(netMsg)
	if err != nil {
		t.Error("TestHandleReceivedProposal handleReceivedProposal error", err)
	}
}

func TestAddVoteMsg(t *testing.T) {
	smr, err := MakeSmr(t)
	if err != nil {
		t.Error("TestAddVoteMsg MakeSmr error", err)
		return
	}

	msg := &pb.ChainedBftVoteMessage{
		ProposalId: []byte("test case1"),
		Signature: &pb.SignInfo{
			Address:   "testcase",
			PublicKey: `{"Curvname":"P-256","X":74695617477160058757747208220371236837474210247114418775262229497812962582435,"Y":51348715319124770392993866417088542497927816017012182211244120852620959209571}`,
		},
	}
	privateKey := `{"Curvname":"P-256","X":74695617477160058757747208220371236837474210247114418775262229497812962582435,"Y":51348715319124770392993866417088542497927816017012182211244120852620959209571,"D":29079635126530934056640915735344231956621504557963207107451663058887647996601}`
	priKey, _ := smr.cryptoClient.GetEcdsaPrivateKeyFromJSON([]byte(privateKey))
	sig, err := utils.MakeVoteMsgSign(smr.cryptoClient, priKey, msg.GetSignature(), msg.GetProposalId())
	msg.Signature = sig
	err = smr.addVoteMsg(msg)
	if err != ErrInValidateSets {
		t.Error("TestAddVoteMsg addVoteMsg error", "error", err)
		return
	}
	msg.Signature.Address = "dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN"
	err = smr.addVoteMsg(msg)
	if err != nil {
		t.Error("TestAddVoteMsg addVoteMsg error", "error", err)
		return
	}
	if _, ok := smr.qcVoteMsgs.Load(string(msg.GetProposalId())); !ok {
		t.Error("TestAddVoteMsg load qcVoteMsgs error")
		return
	}
}

func TestCheckVoteNum(t *testing.T) {
	smr, err := MakeSmr(t)
	if err != nil {
		t.Error("TestCheckVoteNum MakeSmr error", "error", err)
		return
	}
	ok := smr.checkVoteNum([]byte("test"))
	if ok {
		t.Error("TestCheckVoteNum checkVoteNum error")
	}
}

func TestUpdateValidateSets(t *testing.T) {
	smr, err := MakeSmr(t)
	if err != nil {
		t.Error("TestUpdateValidateSets MakeSmr error", "error", err)
		return
	}
	validates := []*cons_base.CandidateInfo{
		&cons_base.CandidateInfo{
			Address: "dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN",
		},
		&cons_base.CandidateInfo{
			Address: "dpzuVdosQrF2kmzumhVeFQZa1aYcdgFp1",
		},
	}
	if err = smr.UpdateValidateSets(validates); err != nil {
		t.Error("TestUpdateValidateSets error", err)
		return
	}
}
