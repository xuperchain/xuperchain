package utils

import (
	cons_base "github.com/xuperchain/xuperunion/consensus/base"
	pb "github.com/xuperchain/xuperunion/pb"
)

// IsInValidateSets check whether addr in validates
func IsInValidateSets(validates []*cons_base.CandidateInfo, addr string) bool {
	for _, v := range validates {
		if v.Address == addr {
			return true
		}
	}
	return false
}

// CheckIsVoted return whether the address have voted for this QC
func CheckIsVoted(votedMsgs *pb.QCSignInfos, voteMsg *pb.SignInfo) bool {
	for _, v := range votedMsgs.QCSignInfos {
		if v.GetAddress() == voteMsg.GetAddress() {
			return true
		}
	}
	return false
}
