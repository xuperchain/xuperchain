package tdpos

import (
	"errors"
	"fmt"
	"strings"

	"strconv"

	"github.com/xuperchain/xuperchain/core/pb"
)

//GenCandidateBallotsPrefix generate candidate ballots prefix
func GenCandidateBallotsPrefix() string {
	return pb.ConsTDposPrefix + "_candidate_ballots_"
}

//GenCandidateBallotsKey generate candidate ballot key
func GenCandidateBallotsKey(address string) string {
	baseKey := GenCandidateBallotsPrefix()
	return baseKey + address
}

// GetCandidateInfoPrefix generate key prefix of candidate info
func GetCandidateInfoPrefix() string {
	return pb.ConsTDposPrefix + "_candidate_info_"
}

// GenCandidateInfoKey gen candidate info key
func GenCandidateInfoKey(address string) string {
	baseKey := GetCandidateInfoPrefix()
	return baseKey + address
}

//ParseCandidateBallotsKey parse candidate ballots key
func ParseCandidateBallotsKey(key string) (string, error) {
	subKeys := strings.Split(key, "_")
	if len(subKeys) != 4 {
		return "", errors.New("parse CandidateBallotsKey error")
	}
	return subKeys[3], nil
}

//GenNominateRecordsPrefix generate nominate records prefix
func GenNominateRecordsPrefix(addr string) string {
	return pb.ConsTDposPrefix + "_nominate_record_" + addr
}

//GenNominateRecordsKey generate nominate records key
func GenNominateRecordsKey(addrNominate, addrCandidate, txid string) string {
	baseKey := GenNominateRecordsPrefix(addrNominate)
	return baseKey + "_" + addrCandidate + "_" + txid
}

//ParseNominateRecordsKey parse nominate records key
func ParseNominateRecordsKey(key string) (string, string, error) {
	subKeys := strings.Split(key, "_")
	if len(subKeys) != 6 {
		return "", "", errors.New("parse NominateRecordsKey error")
	}
	return subKeys[4], subKeys[5], nil
}

// GenCandidateNominatePrefix generate nominate prefix
func GenCandidateNominatePrefix() string {
	return pb.ConsTDposPrefix + "_candidate_nominate_"
}

//GenCandidateNominateKey generate candidate nominate key
func GenCandidateNominateKey(address string) string {
	baseKey := GenCandidateNominatePrefix()
	return baseKey + address
}

//ParseCandidateNominateKey parse candidate nominate key 候选人池子
func ParseCandidateNominateKey(key string) (string, error) {
	subKeys := strings.Split(key, "_")
	if len(subKeys) != 4 {
		return "", errors.New("parse candidate nominate key error")
	}
	return subKeys[3], nil
}

//GenCandidateVotePrefix generate candidate vote prefix
func GenCandidateVotePrefix(addrCandi string) string {
	return pb.ConsTDposPrefix + "_candidate_vote_" + addrCandi + "_"
}

// GenCandidateVoteKey generate candidate vote key
func GenCandidateVoteKey(addrCandi, addrVoter, txid string) string {
	baseKey := GenCandidateVotePrefix(addrCandi)
	return baseKey + addrVoter + "_" + txid
}

//ParseCandidateVoteKey parse  candidate vote key
func ParseCandidateVoteKey(key string) (string, string, error) {
	subKeys := strings.Split(key, "_")
	if len(subKeys) != 6 {
		return "", "", errors.New("parse ParseCandidateVoteKey error")
	}
	return subKeys[4], subKeys[5], nil
}

//GenVoteCandidatePrefix generate candidate vote candidate prefix
func GenVoteCandidatePrefix(addrVoter string) string {
	return pb.ConsTDposPrefix + "_vote_candidate_" + addrVoter + "_"
}

//GenVoteCandidateKey generate vote candidate key
func GenVoteCandidateKey(addrVoter, addrCandi, txid string) string {
	baseKey := GenVoteCandidatePrefix(addrVoter)
	return baseKey + addrCandi + "_" + txid
}

//ParseVoteCandidateKey parse vote candidate key
func ParseVoteCandidateKey(key string) (string, string, error) {
	subKeys := strings.Split(key, "_")
	if len(subKeys) != 6 {
		return "", "", errors.New("parse ParseVoteCandidateKey error")
	}
	return subKeys[4], subKeys[5], nil
}

// GenTermCheckKeyPrefix gen term check key 检票信息, 与版本有关
func GenTermCheckKeyPrefix(version int64) string {
	return pb.ConsTDposPrefix + fmt.Sprintf("_%d", version)
}

//GenTermCheckKey generate term check key
func GenTermCheckKey(version, term int64) string {
	return fmt.Sprintf("%s_%020d", GenTermCheckKeyPrefix(version), term)
}

// ParseTermCheckKey parse term check key
func ParseTermCheckKey(key string) (int64, error) {
	subKeys := strings.Split(key, "_")
	if len(subKeys) != 3 {
		return 0, errors.New("parse parseTermCheckKey error")
	}
	return strconv.ParseInt(subKeys[2], 10, 64)
}

// GenRevokeCandidateKey generate revoke candidate key 候选人退出前投票记录, 便于回滚时恢复
func GenRevokeCandidateKey(address, txid string) string {
	baseKey := fmt.Sprintf("%s_%s", address, txid)
	return pb.ConsTDposPrefix + "_candidate_revoke_" + baseKey
}

// GenRevokeKey generate 生成撤销存储,避免重复撤销
func GenRevokeKey(txid string) string {
	return pb.ConsTDposPrefix + "_revoke_" + txid
}

func checkCandidateName(name string) bool {
	if name == "" {
		return false
	}
	return !strings.Contains(name, "_")
}
