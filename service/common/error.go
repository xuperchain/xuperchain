package common

import (
	"github.com/xuperchain/xuperchain/service/pb"
	ecom "github.com/xuperchain/xupercore/kernel/engines/xuperos/common"
)

// 错误映射配置
var StdErrToXchainErrMap = map[int]pb.XChainErrorEnum{
	ecom.ErrSuccess.Code:                  pb.XChainErrorEnum_SUCCESS,
	ecom.ErrInternal.Code:                 pb.XChainErrorEnum_UNKNOW_ERROR,
	ecom.ErrUnknown.Code:                  pb.XChainErrorEnum_UNKNOW_ERROR,
	ecom.ErrForbidden.Code:                pb.XChainErrorEnum_CONNECT_REFUSE,
	ecom.ErrUnauthorized.Code:             pb.XChainErrorEnum_CONNECT_REFUSE,
	ecom.ErrParameter.Code:                pb.XChainErrorEnum_CONNECT_REFUSE,
	ecom.ErrNewEngineCtxFailed.Code:       pb.XChainErrorEnum_UNKNOW_ERROR,
	ecom.ErrNotEngineType.Code:            pb.XChainErrorEnum_CONNECT_REFUSE,
	ecom.ErrLoadEngConfFailed.Code:        pb.XChainErrorEnum_UNKNOW_ERROR,
	ecom.ErrNewLogFailed.Code:             pb.XChainErrorEnum_UNKNOW_ERROR,
	ecom.ErrNewChainCtxFailed.Code:        pb.XChainErrorEnum_UNKNOW_ERROR,
	ecom.ErrChainExist.Code:               pb.XChainErrorEnum_CONNECT_REFUSE,
	ecom.ErrChainNotExist.Code:            pb.XChainErrorEnum_BLOCKCHAIN_NOTEXIST,
	ecom.ErrChainAlreadyExist.Code:        pb.XChainErrorEnum_CONNECT_REFUSE,
	ecom.ErrChainStatus.Code:              pb.XChainErrorEnum_NOT_READY_ERROR,
	ecom.ErrRootChainNotExist.Code:        pb.XChainErrorEnum_CONNECT_REFUSE,
	ecom.ErrLoadChainFailed.Code:          pb.XChainErrorEnum_UNKNOW_ERROR,
	ecom.ErrContractNewCtxFailed.Code:     pb.XChainErrorEnum_UNKNOW_ERROR,
	ecom.ErrContractInvokeFailed.Code:     pb.XChainErrorEnum_UNKNOW_ERROR,
	ecom.ErrContractNewSandboxFailed.Code: pb.XChainErrorEnum_UNKNOW_ERROR,
	ecom.ErrTxVerifyFailed.Code:           pb.XChainErrorEnum_TX_VERIFICATION_ERROR,
	ecom.ErrTxAlreadyExist.Code:           pb.XChainErrorEnum_TX_DUPLICATE_ERROR,
	ecom.ErrTxNotExist.Code:               pb.XChainErrorEnum_TX_NOT_FOUND_ERROR,
	ecom.ErrTxNotEnough.Code:              pb.XChainErrorEnum_NOT_ENOUGH_UTXO_ERROR,
	ecom.ErrSubmitTxFailed.Code:           pb.XChainErrorEnum_UNKNOW_ERROR,
	ecom.ErrBlockNotExist.Code:            pb.XChainErrorEnum_BLOCK_EXIST_ERROR,
	ecom.ErrProcBlockFailed.Code:          pb.XChainErrorEnum_UNKNOW_ERROR,
	ecom.ErrNewNetEventFailed.Code:        pb.XChainErrorEnum_UNKNOW_ERROR,
	ecom.ErrNewNetworkFailed.Code:         pb.XChainErrorEnum_UNKNOW_ERROR,
	ecom.ErrSendMessageFailed.Code:        pb.XChainErrorEnum_UNKNOW_ERROR,
	ecom.ErrNetworkNoResponse.Code:        pb.XChainErrorEnum_UNKNOW_ERROR,
}
