package server

import (
	"encoding/json"
	"errors"
	context "golang.org/x/net/context"

	"github.com/xuperchain/xuperunion/pb"
)

// XEndorser is the interface for endorser service
type XEndorser interface {
	EndorserCall(ctx context.Context, req *pb.EndorserRequest) (*pb.EndorserResponse, error)
}

// DefaultXEndorser default implementation of XEndorser
type DefaultXEndorser struct {
	svr         *server
	requestType map[string]bool
}

// NewDefaultXEndorser create instance of DefaultXEndorser
func NewDefaultXEndorser(svr *server) *DefaultXEndorser {
	return &DefaultXEndorser{
		svr: svr,
		requestType: map[string]bool{
			"PreExecWithFee":  true,
			"ComplianceCheck": true,
		},
	}
}

// EndorserCall process endorser call
func (dxe *DefaultXEndorser) EndorserCall(ctx context.Context, req *pb.EndorserRequest) (*pb.EndorserResponse, error) {
	// make response header
	resHeader := &pb.Header{
		Error: pb.XChainErrorEnum_SUCCESS,
	}
	if req.GetHeader() == nil {
		resHeader.Logid = req.GetHeader().GetLogid()
	}

	// check param
	if _, ok := dxe.requestType[req.GetRequestName()]; !ok {
		resHeader.Error = pb.XChainErrorEnum_SERVICE_REFUSED_ERROR
		return dxe.generateErrorResponse(req, resHeader, errors.New("request name not supported"))
	}

	switch req.GetRequestName() {
	case "ComplianceCheck":
		success, errcode, err := dxe.processFee(ctx, req)
		if err != nil || !success {
			resHeader.Error = errcode
			return dxe.generateErrorResponse(req, resHeader, err)
		}
		return dxe.generateSuccessResponse(req, []byte(""), resHeader)
	case "PreExecWithFee":
		resData, errcode, err := dxe.getPreExecResult(ctx, req)
		if err != nil {
			resHeader.Error = errcode
			return dxe.generateErrorResponse(req, resHeader, err)
		}
		return dxe.generateSuccessResponse(req, resData, resHeader)
	}

	return nil, nil
}

func (dxe *DefaultXEndorser) getPreExecResult(ctx context.Context, req *pb.EndorserRequest) ([]byte, pb.XChainErrorEnum, error) {
	request := &pb.PreExecWithSelectUTXORequest{}
	err := json.Unmarshal(req.GetRequestData(), request)
	if err != nil {
		return nil, pb.XChainErrorEnum_SERVICE_REFUSED_ERROR, err
	}

	res, err := dxe.svr.PreExecWithSelectUTXO(ctx, request)
	if err != nil {
		return nil, res.GetHeader().GetError(), err
	}

	sData, err := json.Marshal(res)
	if err != nil {
		return nil, pb.XChainErrorEnum_SERVICE_REFUSED_ERROR, err
	}
	return sData, pb.XChainErrorEnum_SUCCESS, nil
}

func (dxe *DefaultXEndorser) processFee(ctx context.Context, req *pb.EndorserRequest) (bool, pb.XChainErrorEnum, error) {
	if req.GetFee() == nil {
		// no fee provided, default to true
		return true, pb.XChainErrorEnum_SUCCESS, nil
	}

	txStatus := &pb.TxStatus{
		Txid:   req.GetFee().GetTxid(),
		Bcname: req.GetBcName(),
		Tx:     req.GetFee(),
	}

	res, err := dxe.svr.PostTx(ctx, txStatus)
	if err != nil {
		return false, res.GetHeader().GetError(), err
	} else if res.GetHeader().GetError() != pb.XChainErrorEnum_SUCCESS {
		return false, res.GetHeader().GetError(), errors.New("Fee post to chain failed")
	}

	return true, pb.XChainErrorEnum_SUCCESS, nil
}

func (dxe *DefaultXEndorser) generateErrorResponse(req *pb.EndorserRequest, header *pb.Header,
	err error) (*pb.EndorserResponse, error) {
	res := &pb.EndorserResponse{
		Header:       header,
		ResponseName: req.GetRequestName(),
	}
	return res, err
}

func (dxe *DefaultXEndorser) generateSuccessResponse(req *pb.EndorserRequest, resData []byte,
	header *pb.Header) (*pb.EndorserResponse, error) {
	res := &pb.EndorserResponse{
		Header:       header,
		ResponseName: req.GetRequestName(),
		ResponseData: resData,
	}
	return res, nil
}
