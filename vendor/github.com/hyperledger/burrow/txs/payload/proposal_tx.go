package payload

import (
	"crypto/sha256"
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/burrow/encoding"
)

func NewProposalTx(propsal *Proposal) *ProposalTx {
	return &ProposalTx{
		Proposal: propsal,
	}
}

func (tx *ProposalTx) Type() Type {
	return TypeProposal
}

func (tx *ProposalTx) GetInputs() []*TxInput {
	return []*TxInput{tx.Input}
}

func (tx *ProposalTx) String() string {
	return fmt.Sprintf("ProposalTx{%v}", tx.Proposal)
}

func (tx *ProposalTx) Any() *Any {
	return &Any{
		ProposalTx: tx,
	}
}

func DecodeProposal(proposalBytes []byte) (*Proposal, error) {
	buf := proto.NewBuffer(proposalBytes)
	proposal := new(Proposal)
	err := buf.Unmarshal(proposal)
	if err != nil {
		return nil, err
	}
	return proposal, nil
}

func (p *Proposal) Encode() ([]byte, error) {
	buf := proto.NewBuffer(nil)
	err := buf.Marshal(p)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (p *Proposal) Hash() []byte {
	bs, err := p.Encode()
	if err != nil {
		panic("failed to encode Proposal")
	}

	hash := sha256.Sum256(bs)

	return hash[:]
}

func (p *Proposal) String() string {
	return ""
}

func (v *Vote) String() string {
	return v.Address.String()
}

func DecodeBallot(ballotBytes []byte) (*Ballot, error) {
	ballot := new(Ballot)
	err := encoding.Decode(ballotBytes, ballot)
	if err != nil {
		return nil, err
	}
	return ballot, nil
}

func (p *Ballot) Encode() ([]byte, error) {
	return encoding.Encode(p)
}
