package utils

import (
	"testing"

	cons_base "github.com/xuperchain/xuperunion/consensus/base"
	"github.com/xuperchain/xuperunion/pb"
)

func TestIsInValidateSets(t *testing.T) {
	testCases := map[string]struct {
		validates []*cons_base.CandidateInfo
		addr      string
		expected  bool
	}{
		"case1": {
			validates: []*cons_base.CandidateInfo{
				&cons_base.CandidateInfo{
					Address: "addr1",
				},
				&cons_base.CandidateInfo{
					Address: "addr2",
				},
			},
			addr:     "addr1",
			expected: true,
		},
		"case2": {
			validates: []*cons_base.CandidateInfo{
				&cons_base.CandidateInfo{
					Address: "addr1",
				},
				&cons_base.CandidateInfo{
					Address: "addr2",
				},
			},
			addr:     "addr3",
			expected: false,
		},
		"case3": {
			validates: []*cons_base.CandidateInfo{},
			addr:      "addr3",
			expected:  false,
		},
	}

	for k, v := range testCases {
		ok := IsInValidateSets(v.validates, v.addr)
		if ok != v.expected {
			t.Error("test IsInValidateSets error", "casename",
				k, "expected", v.expected, "actual", ok)
		}
	}
}

func TestCheckIsVoted(t *testing.T) {
	testCases := map[string]struct {
		votedMsgs *pb.QCSignInfos
		voteMsg   *pb.SignInfo
		expected  bool
	}{
		"case1": {
			votedMsgs: &pb.QCSignInfos{
				QCSignInfos: []*pb.SignInfo{
					&pb.SignInfo{
						Address: "addr1",
					},
					&pb.SignInfo{
						Address: "addr2",
					},
				},
			},
			voteMsg: &pb.SignInfo{
				Address: "addr1",
			},
			expected: true,
		},
		"case2": {
			votedMsgs: &pb.QCSignInfos{
				QCSignInfos: []*pb.SignInfo{
					&pb.SignInfo{
						Address: "addr1",
					},
					&pb.SignInfo{
						Address: "addr2",
					},
				},
			},
			voteMsg: &pb.SignInfo{
				Address: "addr3",
			},
			expected: false,
		},
		"case3": {
			votedMsgs: &pb.QCSignInfos{},
			voteMsg:   &pb.SignInfo{},
			expected:  false,
		},
	}

	for k, v := range testCases {
		ok := CheckIsVoted(v.votedMsgs, v.voteMsg)
		if ok != v.expected {
			t.Error("test IsInValidateSets error", "casename",
				k, "expected", v.expected, "actual", ok)
		}
	}
}
