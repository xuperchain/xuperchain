package smr

import (
	"testing"

	"github.com/xuperchain/xuperunion/consensus/common/chainedbft/utils"
	"github.com/xuperchain/xuperunion/pb"
)

func TestSafeProposal(t *testing.T) {
	smr, err := MakeSmr(t)
	if err != nil {
		t.Error("TestSafeProposal make smr error", "error", err)
		return
	}

	propsQC := &pb.QuorumCert{
		ProposalId:  []byte("propsQC ProposalId"),
		ProposalMsg: []byte("propsQC ProposalMsg"),
		Type:        pb.QCState_PREPARE,
		ViewNumber:  1002,
	}
	justify := &pb.QuorumCert{
		ProposalId:  []byte("justify ProposalId"),
		ProposalMsg: []byte("justify ProposalMsg"),
		Type:        pb.QCState_PREPARE,
		ViewNumber:  1001,
		SignInfos: &pb.QCSignInfos{
			QCSignInfos: []*pb.SignInfo{},
		},
	}
	signInfo := &pb.SignInfo{
		Address:   `dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN`,
		PublicKey: `{"Curvname":"P-256","X":74695617477160058757747208220371236837474210247114418775262229497812962582435,"Y":51348715319124770392993866417088542497927816017012182211244120852620959209571}`,
	}
	privateKey := `{"Curvname":"P-256","X":74695617477160058757747208220371236837474210247114418775262229497812962582435,"Y":51348715319124770392993866417088542497927816017012182211244120852620959209571,"D":29079635126530934056640915735344231956621504557963207107451663058887647996601}`
	priKey, _ := smr.cryptoClient.GetEcdsaPrivateKeyFromJSON([]byte(privateKey))

	signInfo, err = utils.MakeVoteMsgSign(smr.cryptoClient, priKey, signInfo, justify.GetProposalId())
	if err != nil {
		t.Error("TestSafeProposal MakeVoteMsgSign error")
		return
	}
	signInfos := []*pb.SignInfo{}
	signInfos = append(signInfos, signInfo)
	justify.SignInfos.QCSignInfos = signInfos
	if _, err = smr.safeProposal(propsQC, justify); err == nil {
		t.Error("TestSafeProposal case1 safeProposal error")
		return
	}
	smr.lockedQC.ViewNumber = 1000
	if _, err = smr.safeProposal(propsQC, justify); err != nil {
		t.Error("TestSafeProposal case2 safeProposal error")
		return
	}
}
