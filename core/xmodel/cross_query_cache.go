package xmodel

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/xuperchain/xuperchain/core/common/log"
	crypto_client "github.com/xuperchain/xuperchain/core/crypto/client"
	"github.com/xuperchain/xuperchain/core/crypto/hash"
	"github.com/xuperchain/xuperchain/core/pb"
)

const (
	endorserTimeOut = 6 * time.Second
)

// CrossQueryCache cross query struct
type CrossQueryCache struct {
	crossQueryCaches []*pb.CrossQueryInfo
	crossQueryIdx    int
	isPenetrate      bool
}

type queryRes struct {
	queryRes *pb.CrossQueryResponse
	signs    *pb.SignatureInfo
}

// NewCrossQueryCache return CrossQuery instance while preexec
func NewCrossQueryCache() *CrossQueryCache {
	return &CrossQueryCache{
		isPenetrate: true,
	}
}

// NewCrossQueryCacheWithData return CrossQuery instance while posttx
func NewCrossQueryCacheWithData(crossQueries []*pb.CrossQueryInfo) *CrossQueryCache {
	return &CrossQueryCache{
		crossQueryCaches: crossQueries,
		isPenetrate:      false,
	}
}

// CrossQuery query contract from otherchain
func (cqc *CrossQueryCache) CrossQuery(
	crossQueryRequest *pb.CrossQueryRequest,
	queryMeta *pb.CrossQueryMeta) (*pb.ContractResponse, error) {
	log.Info("Receive CrossQuery", "crossQueryRequest", crossQueryRequest, "queryMeta", queryMeta)
	if !isQueryMetaValid(queryMeta) {
		return nil, fmt.Errorf("isQueryParamValid check failed")
	}
	// Call endorsor for responce
	if cqc.isPenetrate {
		queryInfo, err := crossQueryFromEndorsor(crossQueryRequest, queryMeta)
		if err != nil {
			log.Info("crossQueryFromEndorsor error", "error", err.Error())
			return nil, err
		}
		cqc.crossQueryCaches = append(cqc.crossQueryCaches, queryInfo)
		return queryInfo.GetResponse().GetResponse(), nil
	}

	// 验证背书规则、参数有效性、时间戳有效性
	if cqc.crossQueryIdx > len(cqc.crossQueryCaches)-1 {
		return nil, fmt.Errorf("len of crossQueryCaches not match the contract")
	}
	crossQuery := cqc.crossQueryCaches[cqc.crossQueryIdx]

	// 验证request、签名等信息
	if !isCossQueryValid(crossQueryRequest, queryMeta, crossQuery) {
		return nil, fmt.Errorf("isCossQueryValid check failed")
	}
	cqc.crossQueryIdx++
	return crossQuery.GetResponse().GetResponse(), nil
}

// GetCrossQueryRWSets get cross query rwsets
func (cqc *CrossQueryCache) GetCrossQueryRWSets() []*pb.CrossQueryInfo {
	return cqc.crossQueryCaches
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

func endorsorQueryWithGroup(req *pb.EndorserRequest, queryMeta *pb.CrossQueryMeta) (*pb.CrossQueryResponse, []*pb.SignatureInfo, error) {
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
	signs := []*pb.SignatureInfo{}
	var conRes *pb.CrossQueryResponse
	lenCh := len(msgChan)
	if lenCh <= 0 {
		return nil, nil, errors.New("endorsorQueryWithGroup res is nil")
	}

	breakFlag := 0
	for r := range msgChan {
		if breakFlag == 0 {
			conRes = r.queryRes
		} else {
			if !isCrossQueryResponseEqual(conRes, r.queryRes) {
				return conRes, signs, errors.New("endorsorQueryWithGroup ContractResponse different")
			}
		}
		signs = append(signs, r.signs)
		if breakFlag >= lenCh-1 {
			break
		}
		breakFlag++
	}
	return conRes, signs, nil
}

func endorsorQuery(req *pb.EndorserRequest, ce *pb.CrossEndorsor) (*queryRes, error) {
	ctx, _ := context.WithTimeout(context.TODO(), endorserTimeOut)
	conn, err := NewEndorsorConn(ce.GetHost())
	if err != nil {
		log.Info("endorsorQuery NewEndorsorClient error", "err", err.Error())
		return nil, err
	}
	defer conn.Close()
	cli := pb.NewXendorserClient(conn)
	endorsorRes, err := cli.EndorserCall(ctx, req)
	if err != nil {
		log.Info("endorsorQuery EndorserCall error", "err", err.Error())
		return nil, err
	}
	res := &pb.CrossQueryResponse{}
	err = json.Unmarshal(endorsorRes.GetResponseData(), res)
	if err != nil {
		log.Info("endorsorQuery Unmarshal error", "err", err)
		return nil, err
	}
	queryRes := &queryRes{
		queryRes: res,
		signs:    endorsorRes.GetEndorserSign(),
	}
	return queryRes, nil
}

// 验证CrossQuery背书信息
func isCossQueryValid(
	crossQueryRequest *pb.CrossQueryRequest,
	queryMeta *pb.CrossQueryMeta,
	queryInfo *pb.CrossQueryInfo) bool {
	// check req from bridge and req from preexec
	if !isMsgEqual(crossQueryRequest, queryInfo.GetRequest()) {
		log.Info("isCossQueryValid isMsgEqual not equal")
		return false
	}

	// check endorsor info
	signs, ok := isEndorsorInfoValid(queryMeta, queryInfo.GetSigns())
	if !ok {
		log.Info("isEndorsorInfoValid not ok")
		return false
	}

	// check endorsor sign
	if !isEndorsorSignValid(signs, queryInfo) {
		log.Info("isEndorsorSignValid not ok")
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
		log.Info("isEndorsorInfoValid failed")
		return nil, false
	}
	return signsValid, true
}

func isEndorsorSignValid(signsValid []*pb.SignatureInfo, queryInfo *pb.CrossQueryInfo) bool {
	reqData, err := json.Marshal(queryInfo.GetRequest())
	if err != nil {
		log.Info("Marshal Request failed", "err", err)
		return false
	}
	resData, err := json.Marshal(queryInfo.GetResponse())
	if err != nil {
		log.Info("Marshal Response failed", "err", err)
		return false
	}
	cryptoClient, err := crypto_client.CreateCryptoClient(crypto_client.CryptoTypeDefault)
	data := append(reqData[:], resData[:]...)
	digest := hash.UsingSha256(data)
	for idx := range signsValid {
		pk, err := cryptoClient.GetEcdsaPublicKeyFromJSON([]byte(signsValid[idx].GetPublicKey()))
		if err != nil {
			log.Info("GetEcdsaPublicKeyFromJSON failed")
			return false
		}
		ok, err := cryptoClient.VerifyECDSA(pk, signsValid[idx].GetSign(), digest)
		if !ok || err != nil {
			log.Info("VerifyECDSA failed", "ok", ok, "err", err)
			return false
		}
	}
	return true
}

func isCrossQueryResponseEqual(a, b *pb.CrossQueryResponse) bool {
	if a.GetResponse().GetStatus() != b.GetResponse().GetStatus() {
		return false
	}
	if a.GetResponse().GetMessage() != b.GetResponse().GetMessage() {
		return false
	}
	if !bytes.Equal(a.GetResponse().GetBody(), b.GetResponse().GetBody()) {
		return false
	}
	return true
}
