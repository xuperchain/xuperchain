package xmodel

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	log "github.com/xuperchain/log15"
	crypto_client "github.com/xuperchain/xuperchain/core/crypto/client"
	"github.com/xuperchain/xuperchain/core/crypto/hash"
	"github.com/xuperchain/xuperchain/core/pb"
)

const (
	timeOut = 6
)

// CrossQueryCache cross query struct
type CrossQueryCache struct {
	lg               log.Logger
	crossQueryCaches []*pb.CrossQueryInfo
	crossQueryIdx    int
	isPenetrate      bool
}

type queryRes struct {
	*pb.CrossQueryResponse
	*pb.SignatureInfo
}

// NewCrossQueryCache return CrossQuery instance while preexec
func NewCrossQueryCache() *CrossQueryCache {
	lg := log.New("module", "cross_query_cache")
	lg.SetHandler(log.StreamHandler(os.Stderr, log.LogfmtFormat()))
	return &CrossQueryCache{
		lg:          lg,
		isPenetrate: true,
	}
}

// NewCrossQueryCacheWirthData return CrossQuery instance while posttx
func NewCrossQueryCacheWirthData(crossQueries []*pb.CrossQueryInfo) *CrossQueryCache {
	lg := log.New("module", "cross_query_cache")
	lg.SetHandler(log.StreamHandler(os.Stderr, log.LogfmtFormat()))
	return &CrossQueryCache{
		lg:               lg,
		crossQueryCaches: crossQueries,
		isPenetrate:      false,
	}
}

// CrossQuery query contract from otherchain
func (cqc *CrossQueryCache) CrossQuery(
	crossQueryRequest *pb.CrossQueryRequest,
	queryMeta *pb.CrossQueryMeta) (*pb.ContractResponse, error) {
	cqc.lg.Info("Receive CrossQuery", "crossQueryRequest", crossQueryRequest, "queryMeta", queryMeta)
	if !isQueryMetaValid(queryMeta) {
		return nil, fmt.Errorf("isQueryParamValid check failed")
	}
	// Call endorsor for responce
	if cqc.isPenetrate {
		queryInfo, err := cqc.crossQueryFromEndorsor(crossQueryRequest, queryMeta)
		if err != nil {
			cqc.lg.Info("crossQueryFromEndorsor error", "error", err.Error())
			return nil, err
		}
		cqc.crossQueryCaches = append(cqc.crossQueryCaches, queryInfo)
		return queryInfo.GetResponse().GetResponse(), nil
	}

	// 验证背书规则、参数有效性、时间戳有效性
	crossQuery := cqc.crossQueryCaches[cqc.crossQueryIdx]

	// 验证request、签名等信息
	for !cqc.isCossQueryValid(crossQueryRequest, queryMeta, crossQuery) {
		return nil, fmt.Errorf("isCossQueryValid check failed")
	}
	cqc.crossQueryIdx++
	return crossQuery.GetResponse().GetResponse(), nil
}

// isQueryParamValid 验证 query meta 背书策略是否有效
func isQueryMetaValid(queryMeta *pb.CrossQueryMeta) bool {
	return len(queryMeta.GetEndorsors()) >= int(queryMeta.GetChainMeta().GetMinEndorsorNum())
}

// crossQueryFromEndorsor will query cross from endorsor
func (cqc *CrossQueryCache) crossQueryFromEndorsor(
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

	res, signs, err := cqc.endorsorQueryWithGroup(req, queryMeta)
	if err != nil {
		return nil, err
	}
	return &pb.CrossQueryInfo{
		Request:  crossQueryRequest,
		Response: res,
		Signs:    signs,
	}, nil
}

func (cqc *CrossQueryCache) endorsorQueryWithGroup(req *pb.EndorserRequest, queryMeta *pb.CrossQueryMeta) (*pb.CrossQueryResponse, []*pb.SignatureInfo, error) {
	wg := sync.WaitGroup{}
	msgChan := make(chan *queryRes, len(queryMeta.GetEndorsors()))

	for idx := range queryMeta.GetEndorsors() {
		wg.Add(1)
		go func(req *pb.EndorserRequest, ce *pb.CrossEndorsor) {
			defer wg.Done()
			res, err := cqc.endorsorQuery(req, ce)
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

	i := 0
	for r := range msgChan {
		if i == 0 {
			conRes = r.CrossQueryResponse
		} else {
			if !isMsgEqual(conRes, r.CrossQueryResponse) {
				return conRes, signs, errors.New("endorsorQueryWithGroup ContractResponse different")
			}
		}
		signs = append(signs, r.SignatureInfo)
		if i >= lenCh-1 {
			break
		}
		i++
	}
	return conRes, signs, nil
}

func (cqc *CrossQueryCache) endorsorQuery(req *pb.EndorserRequest, ce *pb.CrossEndorsor) (*queryRes, error) {
	ctx, _ := context.WithTimeout(context.TODO(), timeOut*time.Second)
	conn, err := NewEndorsorConn(ce.GetHost())
	if err != nil {
		cqc.lg.Info("endorsorQuery NewEndorsorClient error", "err", err.Error())
		return nil, err
	}
	defer conn.Close()
	cli := pb.NewXendorserClient(conn)
	endorsorRes, err := cli.EndorserCall(ctx, req)
	if err != nil {
		cqc.lg.Info("endorsorQuery EndorserCall error", "err", err.Error())
		return nil, err
	}
	res := &pb.CrossQueryResponse{}
	err = json.Unmarshal(endorsorRes.GetResponseData(), res)
	if err != nil {
		cqc.lg.Info("endorsorQuery Unmarshal error", "err", err)
		return nil, err
	}
	queryRes := &queryRes{
		res,
		endorsorRes.GetEndorserSign(),
	}
	return queryRes, nil
}

// 验证CrossQuery背书信息
func (cqc *CrossQueryCache) isCossQueryValid(
	crossQueryRequest *pb.CrossQueryRequest,
	queryMeta *pb.CrossQueryMeta,
	queryInfo *pb.CrossQueryInfo) bool {
	// check req from bridge and req from preexec
	if !isMsgEqual(crossQueryRequest, queryInfo.GetRequest()) {
		cqc.lg.Info("isCossQueryValid isMsgEqual not equal")
		return false
	}

	// check endorsor info
	signs, ok := cqc.isEndorsorInfoValid(queryMeta, queryInfo.GetSigns())
	if !ok {
		cqc.lg.Info("isEndorsorInfoValid not ok")
		return false
	}

	// check endorsor sign
	if !cqc.isEndorsorSignValid(signs, queryInfo) {
		cqc.lg.Info("isEndorsorSignValid not ok")
		return false
	}
	return true
}

func (cqc *CrossQueryCache) isEndorsorInfoValid(queryMeta *pb.CrossQueryMeta, signs []*pb.SignatureInfo) ([]*pb.SignatureInfo, bool) {
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
		cqc.lg.Info("isEndorsorInfoValid failed")
		return nil, false
	}
	return signsValid, true
}

func (cqc *CrossQueryCache) isEndorsorSignValid(signsValid []*pb.SignatureInfo, queryInfo *pb.CrossQueryInfo) bool {
	reqData, err := json.Marshal(queryInfo.GetRequest())
	if err != nil {
		cqc.lg.Info("Marshal Request failed", "err", err)
		return false
	}
	resData, err := json.Marshal(queryInfo.GetResponse())
	if err != nil {
		cqc.lg.Info("Marshal Response failed", "err", err)
		return false
	}
	cryptoClient, err := crypto_client.CreateCryptoClient(crypto_client.CryptoTypeDefault)
	data := append(reqData[:], resData[:]...)
	digest := hash.UsingSha256(data)
	for idx := range signsValid {
		pk, err := cryptoClient.GetEcdsaPublicKeyFromJSON([]byte(signsValid[idx].GetPublicKey()))
		if err != nil {
			cqc.lg.Info("GetEcdsaPublicKeyFromJSON failed")
			return false
		}
		ok, err := cryptoClient.VerifyECDSA(pk, signsValid[idx].GetSign(), digest)
		if !ok || err != nil {
			cqc.lg.Info("VerifyECDSA failed", "ok", ok, "err", err)
			return false
		}
	}
	return true
}

// GetCrossQueryRWSets get cross query rwsets
func (cqc *CrossQueryCache) GetCrossQueryRWSets() []*pb.CrossQueryInfo {
	return cqc.crossQueryCaches
}
