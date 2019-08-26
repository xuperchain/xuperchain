package external

import (
	"github.com/xuperchain/xuperunion/consensus/common/chainedbft/utils"
	crypto_client "github.com/xuperchain/xuperunion/crypto/client"
	pb "github.com/xuperchain/xuperunion/pb"
)

// MockExternalConsensus mock the ExternalInterface
// Used in unit tests
type MockExternalConsensus struct {
}

// CallPreQc is the the given QC's ProposalMsg's JustifyQC
func (mec *MockExternalConsensus) CallPreQc(qc *pb.QuorumCert) (*pb.QuorumCert, error) {
	preQc := &pb.QuorumCert{
		ProposalId:  []byte("justify ProposalId"),
		ProposalMsg: []byte("justify ProposalMsg"),
		Type:        pb.QCState_PREPARE,
		ViewNumber:  1004,
		SignInfos: &pb.QCSignInfos{
			QCSignInfos: []*pb.SignInfo{},
		},
	}
	signInfo := &pb.SignInfo{
		Address:   `dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN`,
		PublicKey: `{"Curvname":"P-256","X":74695617477160058757747208220371236837474210247114418775262229497812962582435,"Y":51348715319124770392993866417088542497927816017012182211244120852620959209571}`,
	}
	privateKey := `{"Curvname":"P-256","X":74695617477160058757747208220371236837474210247114418775262229497812962582435,"Y":51348715319124770392993866417088542497927816017012182211244120852620959209571,"D":29079635126530934056640915735344231956621504557963207107451663058887647996601}`
	cryptoClient, err := crypto_client.CreateCryptoClient(crypto_client.CryptoTypeDefault)
	if err != nil {
		return nil, err
	}
	priKey, _ := cryptoClient.GetEcdsaPrivateKeyFromJSON([]byte(privateKey))

	signInfo, err = utils.MakeVoteMsgSign(cryptoClient, priKey, signInfo, preQc.GetProposalId())
	if err != nil {
		return nil, err
	}
	signInfos := []*pb.SignInfo{}
	signInfos = append(signInfos, signInfo)
	preQc.SignInfos.QCSignInfos = signInfos
	return preQc, nil
}

// CallPreProposalMsg call external consensus for the marshal format of proposalMsg's parent block
func (mec *MockExternalConsensus) CallPreProposalMsg(proposalMsg []byte) ([]byte, error) {
	return nil, nil
}

// CallPrePreProposalMsg call external consensus for the marshal format of proposalMsg's grandpa's block
func (mec *MockExternalConsensus) CallPrePreProposalMsg(proposalMsg []byte) ([]byte, error) {
	return nil, nil
}

// CallVerifyQc call external consensus for proposalMsg verify with the given QC
func (mec *MockExternalConsensus) CallVerifyQc(qc *pb.QuorumCert) (bool, error) {
	return true, nil
}

// CallProposalMsgWithProposalID call  external consensus for proposalMsg  with the given ProposalID
func (mec *MockExternalConsensus) CallProposalMsgWithProposalID(proposalID []byte) ([]byte, error) {
	return nil, nil
}
