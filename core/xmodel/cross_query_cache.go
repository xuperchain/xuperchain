package xmodel

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	crypto_client "github.com/xuperchain/xuperchain/core/crypto/client"
	"github.com/xuperchain/xuperchain/core/crypto/hash"
	"github.com/xuperchain/xuperchain/core/pb"
)

const (
	timeOut = 6
)

// CrossQueryCache cross query struct
type CrossQueryCache struct {
	crossQueryCaches []*pb.CrossQueryInfo
	crossQueryIdx    int
	isPenetrate      bool
}

type queryRes struct {
	*pb.ContractResponse
	*pb.SignatureInfo
}

// NewCrossQueryCache return CrossQuery instance while preexec
func NewCrossQueryCache() *CrossQueryCache {
	return &CrossQueryCache{
		isPenetrate: true,
	}
}

// NewCrossQueryCacheWirthData return CrossQuery instance while posttx
func NewCrossQueryCacheWirthData(crossQueries []*pb.CrossQueryInfo) *CrossQueryCache {
	return &CrossQueryCache{
		crossQueryCaches: crossQueries,
		isPenetrate:      false,
	}
}

// CrossQuery query contract from otherchain
func (cqc *CrossQueryCache) CrossQuery(
	crossQueryRequest *pb.CrossQueryRequest,
	queryMeta *pb.CrossQueryMeta) (*pb.ContractResponse, error) {
	if !isQueryMetaValid(queryMeta) {
		return nil, fmt.Errorf("isQueryParamValid check failed")
	}
	// Call endorsor for responce
	if cqc.isPenetrate {
		queryInfo, err := crossQueryFromEndorsor(crossQueryRequest, queryMeta)
		if err != nil {
			return nil, err
		}
		cqc.crossQueryCaches = append(cqc.crossQueryCaches, queryInfo)
		return queryInfo.GetResponse(), nil
	}

	// 验证背书规则、参数有效性、时间戳有效性
	crossQuery := cqc.crossQueryCaches[cqc.crossQueryIdx]

	// 验证request、签名等信息
	for isCossQueryValid(crossQueryRequest, queryMeta, crossQuery) {
		return nil, fmt.Errorf("isCossQueryValid check failed")
	}
	cqc.crossQueryIdx++
	return crossQuery.GetResponse(), nil
}

// isQueryParamValid 验证 query meta 背书策略是否有效
func isQueryMetaValid(queryMeta *pb.CrossQueryMeta) bool {
	return len(queryMeta.GetEndorsors()) >= int(queryMeta.GetChainMeta().GetMinEndorsorNum())
}

// crossQueryFromEndorsor will query cross from endorsor
func crossQueryFromEndorsor(
	crossQueryRequest *pb.CrossQueryRequest,
	queryMeta *pb.CrossQueryMeta) (*pb.CrossQueryInfo, error) {

	reqData, err := json.Marshal(crossQueryRequest)
	if err != nil {
		return nil, err
	}

	req := &pb.EndorserRequest{
		RequestName: "CrossQueryPreExec",
		BcName:      crossQueryRequest.GetBcname(),
		RequestData: reqData,
	}

	res, signs, err := endorsorQueryWithGroup(req, queryMeta)
	if err != nil {
		return nil, err
	}
	return &pb.CrossQueryInfo{
		Request:  crossQueryRequest,
		Response: res,
		Signs:    signs,
	}, nil
}

func endorsorQueryWithGroup(req *pb.EndorserRequest, queryMeta *pb.CrossQueryMeta) (*pb.ContractResponse, []*pb.SignatureInfo, error) {
	wg := sync.WaitGroup{}
	msgChan := make(chan *queryRes, len(queryMeta.GetEndorsors()))

	for idx := range queryMeta.GetEndorsors() {
		wg.Add(1)
		go func(req *pb.EndorserRequest, ce *pb.CrossEndorsor) {
			defer wg.Done()
			res, err := endorsorQuery(req, ce)
			if err != nil {
				return
			}
			msgChan <- res
		}(req, queryMeta.GetEndorsors()[idx])
	}
	wg.Wait()
	// 处理所有请求结果
	lenCh := len(msgChan)
	if lenCh <= 0 {
		return nil, nil, errors.New("endorsorQueryWithGroup res is nil")
	}
	signs := []*pb.SignatureInfo{}
	var conRes *pb.ContractResponse
	i := 0
	for r := range msgChan {
		if i == 0 {
			conRes = r.ContractResponse
		} else {
			if !isMsgEqual(conRes, r.ContractResponse) {
				return conRes, signs, errors.New("endorsorQueryWithGroup ContractResponse different")
			}
		}
		signs = append(signs, r.SignatureInfo)
	}

	return conRes, signs, nil
}

func endorsorQuery(req *pb.EndorserRequest, ce *pb.CrossEndorsor) (*queryRes, error) {
	ctx, _ := context.WithTimeout(context.TODO(), timeOut)
	cli, err := NewEndorsorClient(ce.GetHost())
	if err != nil {
		return nil, err
	}
	endorsorRes, err := cli.EndorserCall(ctx, req)
	if err != nil {
		return nil, err
	}
	res := &pb.ContractResponse{}
	err = json.Unmarshal(endorsorRes.GetResponseData(), res)
	if err != nil {
		return nil, err
	}

	return &queryRes{
		res,
		endorsorRes.GetEndorserSign(),
	}, nil
}

// 验证CrossQuery背书信息
func isCossQueryValid(
	crossQueryRequest *pb.CrossQueryRequest,
	queryMeta *pb.CrossQueryMeta,
	queryInfo *pb.CrossQueryInfo) bool {
	// check req from bridge and req from preexec
	if !isMsgEqual(crossQueryRequest, queryInfo.GetRequest()) {
		return false
	}

	// check endorsor info
	signs, ok := isEndorsorInfoValid(queryMeta, queryInfo.GetSigns())
	if !ok {
		return false
	}

	// check endorsor sign
	if !isEndorsorSignValid(signs, queryInfo) {
		return false
	}
	return true
}

func isEndorsorInfoValid(queryMeta *pb.CrossQueryMeta, signs []*pb.SignatureInfo) ([]*pb.SignatureInfo, bool) {
	signMap := map[string]*pb.SignatureInfo{}
	for idx := range signs {
		signMap[signs[idx].GetPublicKey()] = signs[idx]
	}

	endorsorMap := map[string]*pb.CrossEndorsor{}
	endorsors := queryMeta.GetEndorsors()
	for idx := range endorsors {
		endorsorMap[endorsors[idx].GetPubKey()] = endorsors[idx]
	}
	signsValid := []*pb.SignatureInfo{}
	for k, v := range signMap {
		if endorsorMap[k] != nil {
			signsValid = append(signsValid, v)
		}
	}
	if len(signsValid) < int(queryMeta.GetChainMeta().GetMinEndorsorNum()) {
		return nil, false
	}
	return signsValid, true
}

func isEndorsorSignValid(signsValid []*pb.SignatureInfo, queryInfo *pb.CrossQueryInfo) bool {
	reqData, err := json.Marshal(queryInfo.GetRequest())
	if err != nil {
		return false
	}
	resData, err := json.Marshal(queryInfo.GetRequest())
	if err != nil {
		return false
	}
	cryptoClient, err := crypto_client.CreateCryptoClient(crypto_client.CryptoTypeDefault)
	data := append(reqData[:], resData[:]...)
	digest := hash.UsingSha256(data)
	for idx := range signsValid {
		pk, err := cryptoClient.GetEcdsaPublicKeyFromJSON([]byte(signsValid[idx].GetPublicKey()))
		if err != nil {
			return false
		}
		ok, err := cryptoClient.VerifyECDSA(pk, signsValid[idx].GetSign(), digest)
		if !ok || err != nil {
			return false
		}
	}
	return true
}

// GetCrossQueryRWSets get cross query rwsets
func (cqc *CrossQueryCache) GetCrossQueryRWSets() []*pb.CrossQueryInfo {
	return cqc.crossQueryCaches
}
