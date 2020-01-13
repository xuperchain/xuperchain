package xchaincore

import (
	"github.com/xuperchain/xuperchain/core/global"
	"github.com/xuperchain/xuperchain/core/pb"
)

// GenerateTx generate transaction from tx data
func (xc *XChainCore) GenerateTx(in *pb.TxData, hd *global.XContext) *pb.TxStatus {
	out := &pb.TxStatus{Header: in.Header}
	out.Header.Error = pb.XChainErrorEnum_SUCCESS
	if xc.Status() != global.Normal {
		out.Header.Error = pb.XChainErrorEnum_CONNECT_REFUSE // 拒绝
		return out
	}
	t, err := xc.Utxovm.GenerateTx(in)
	xc.Speed.Add("GenerateTx")
	if err != nil {
		out.Header.Error = HandlerUtxoError(err)
	} else {
		out.Tx = t
		out.Bcname = in.Bcname
		out.Txid = out.Tx.Txid
	}
	return out
}
