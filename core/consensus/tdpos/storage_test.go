package tdpos

import "testing"

func TestGenCandidateBallotsKey(t *testing.T) {
	res := GenCandidateBallotsKey("addr1")
	if res != "D_candidate_ballots_addr1" {
		t.Error("genCandidateBallotsKey error")
	}
}

func TestParseCandidateBallotsKey(t *testing.T) {
	key := "D_candidate_ballots_addr1"
	addr, _ := ParseCandidateBallotsKey(key)
	if addr != "addr1" {
		t.Error("TestParseCandidateBallotsKey error")
	}
}

func TestGenNominateRecordsKey(t *testing.T) {
	res := GenNominateRecordsPrefix("add1")
	if res != "D_nominate_record_add1" {
		t.Error("TestGenNominateRecordsKey error")
	}
}

func TestParseNominateRecordsKey(t *testing.T) {
	key := "D_nominate_record_addrNominate_addrCandidate_txid"
	addrCandi, txID, _ := ParseNominateRecordsKey(key)
	if addrCandi != "addrCandidate" || txID != "txid" {
		t.Error("TestParseNominateRecordsKey error")
	}
}

func TestGenCandidateNominateKey(t *testing.T) {
	res := GenCandidateNominateKey("addr1")
	if res != "D_candidate_nominate_addr1" {
		t.Error("genCandidateNominateKey error")
	}
}

func TestGenCandidateVoteKey(t *testing.T) {
	res := GenCandidateVoteKey("addrCandi", "addrVoter", "ahvcqeuhcjhv")
	expected := "D_candidate_vote_addrCandi_addrVoter_ahvcqeuhcjhv"
	if res != expected {
		t.Error("genCandidateVoteKey error")
	}
}

func TestParseCandidateVoteKey(t *testing.T) {
	key := "D_candidate_vote_addrCandi_addrVoter_ahvcqeuhcjhv"
	addrVoter, txID, _ := ParseCandidateVoteKey(key)
	if addrVoter != "addrVoter" || txID != "ahvcqeuhcjhv" {
		t.Error("TestParseCandidateVoteKey error")
	}

}

func TestGenVoteCandidateKey(t *testing.T) {
	res := GenVoteCandidateKey("addrVoter", "addrCandi", "ahvcqeuhcjhv")
	expected := "D_vote_candidate_addrVoter_addrCandi_ahvcqeuhcjhv"
	if res != expected {
		t.Error("genVoteCandidateKey error")
	}
}

func TestParseVoteCandidateKey(t *testing.T) {
	key := "D_vote_candidate_addrVoter_addrCandi_ahvcqeuhcjhv"
	addrCandi, txID, _ := ParseVoteCandidateKey(key)
	if addrCandi != "addrCandi" || txID != "ahvcqeuhcjhv" {
		t.Error("TestParseVoteCandidateKey error")
	}
}

func TestGenRevokeKey(t *testing.T) {
	res := GenRevokeCandidateKey("addr1", "ahvcqeuhcjhv")
	if res != "D_candidate_revoke_addr1_ahvcqeuhcjhv" {
		t.Error("genRevokeKey error")
	}

	res2 := GenRevokeKey("ahvcqeuhcjhv")
	expect := "D_revoke_ahvcqeuhcjhv"
	if res2 != expect {
		t.Error("genRevokeKey error")
	}
}

func TestGenTermCheckKey(t *testing.T) {
	version := int64(11)
	term := int64(12)
	expect := "D_11_00000000000000000012"
	res := GenTermCheckKey(version, term)
	if res != expect {
		t.Error("TestGenTermCheckKey error")
	}
}

func TestParseTermCheckKey(t *testing.T) {
	key := "D_11_00000000000000000012"
	term, _ := ParseTermCheckKey(key)
	if term != 12 {
		t.Error("TestParseTermCheckKey error")
	}
}
