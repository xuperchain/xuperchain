package utils

import (
	cons_base "github.com/xuperchain/xuperunion/consensus/base"
	chainedbft_pb "github.com/xuperchain/xuperunion/consensus/common/chainedbft/pb"
)

// IsInValidateSets check whether addr in validates
func IsInValidateSets(validates []*cons_base.CandidateInfo, addr string) bool {
	// todo
	return false
}

// CheckIsVoted return whether the address have voted for this QC
func CheckIsVoted(votedMsgs *chainedbft_pb.QCSignInfos, voteMsg *chainedbft_pb.SignInfo) bool {
	for _, v := range votedMsgs.QCSignInfos {
		if v.GetAddress() == voteMsg.GetAddress() {
			return true
		}
	}
	return false
}
