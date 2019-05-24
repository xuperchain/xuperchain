package utxo

import (
	"fmt"
	"math/big"

	"github.com/xuperchain/xuperunion/common"
	crypto_client "github.com/xuperchain/xuperunion/crypto/client"
	"github.com/xuperchain/xuperunion/pb"
	"github.com/xuperchain/xuperunion/utxo/txhash"
)

// GenerateTx 根据一个原始订单, 得到UTXO格式的交易, 相当于预执行, 会在内存中锁定一段时间UTXO, 但是不修改kv存储
func (uv *UtxoVM) GenerateTx(txReq *pb.TxData) (*pb.Transaction, error) {
	cryptoClient, cryptoErr := crypto_client.CreateCryptoClient(crypto_client.CryptoTypeDefault)
	if cryptoErr != nil {
		return nil, cryptoErr
	}
	utxoTx := &pb.Transaction{}
	utxoTx.Desc = txReq.Desc
	utxoTx.Nonce = txReq.Nonce
	utxoTx.Timestamp = txReq.Timestamp
	utxoTx.Version = TxVersion
	utxoTx.AuthRequire = append(utxoTx.AuthRequire, txReq.FromAddr)
	utxoTx.Initiator = txReq.FromAddr
	totalNeed := big.NewInt(0) // 需要支付的总额
	// 这个for累加一下 一共需要多少钱发起交易
	for _, txAccount := range txReq.Account {
		amount := big.NewInt(0)
		amount.SetString(txAccount.Amount, 10) // 10进制转换大整数
		if amount.Cmp(big.NewInt(0)) < 0 {
			return nil, ErrNegativeAmount
		}
		totalNeed.Add(totalNeed, amount)
		txOutput := &pb.TxOutput{}
		txOutput.ToAddr = []byte(txAccount.Address)
		txOutput.Amount = amount.Bytes()
		txOutput.FrozenHeight = txAccount.FrozenHeight
		utxoTx.TxOutputs = append(utxoTx.TxOutputs, txOutput)
	}
	// 一般的交易
	utxoTx.Coinbase = false
	txInputs, _, utxoTotal, selectErr := uv.SelectUtxos(txReq.FromAddr, txReq.FromPubkey, totalNeed, true, false)
	if selectErr != nil {
		uv.xlog.Warn("select utxos error", "err", selectErr)
		return nil, selectErr
	}
	utxoTx.TxInputs = txInputs
	// 多出来的utxo需要再转给自己
	if utxoTotal.Cmp(totalNeed) > 0 {
		delta := utxoTotal.Sub(utxoTotal, totalNeed)
		txOutput := &pb.TxOutput{}
		txOutput.ToAddr = []byte(txReq.FromAddr) // 收款人就是汇款人自己
		txOutput.Amount = delta.Bytes()
		utxoTx.TxOutputs = append(utxoTx.TxOutputs, txOutput)
	}
	signTx, signErr := txhash.ProcessSignTx(cryptoClient, utxoTx, []byte(txReq.FromScrkey))
	if signErr != nil {
		return nil, signErr
	}
	signInfo := &pb.SignatureInfo{
		PublicKey: txReq.FromPubkey,
		Sign:      signTx,
	}
	utxoTx.InitiatorSigns = append(utxoTx.InitiatorSigns, signInfo)
	utxoTx.AuthRequireSigns = utxoTx.InitiatorSigns
	utxoTx.Txid, _ = txhash.MakeTransactionID(utxoTx)

	// check if size limit exceeded
	txSize, err := common.GetTxSerializedSize(utxoTx)
	if nil != err {
		uv.xlog.Warn("failed to GetTxSerializedSize", "err", err)
		return nil, err
	}
	if txSize > uv.ledger.GetMaxBlockSize() {
		uv.xlog.Warn("tx size limit exceeded", "txSize", txSize)
		return nil, ErrTxSizeLimitExceeded
	}
	uv.xlog.Trace("make txid done", "txid", fmt.Sprintf("%x", utxoTx.Txid))
	return utxoTx, nil
}
