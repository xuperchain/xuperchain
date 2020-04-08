package ledger

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/xuperchain/xuperchain/core/global"
	"math"
	"sort"

	"github.com/xuperchain/xuperchain/core/crypto/hash"
	"github.com/xuperchain/xuperchain/core/pb"
)

func getLeafSize(txCount int) int {
	if txCount&(txCount-1) == 0 { // 刚好是2的次幂
		return txCount
	}
	exponent := uint(math.Log2(float64(txCount))) + 1
	return 1 << exponent // 2^exponent
}

// MakeMerkleTree generate merkele-tree
func MakeMerkleTree(txList []*pb.Transaction) [][]byte {
	txCount := len(txList)
	if txCount == 0 {
		return nil
	}
	leafSize := getLeafSize(txCount) //需要补充为完全树
	treeSize := leafSize*2 - 1       //整个树的节点个数
	tree := make([][]byte, treeSize)
	for i, tx := range txList {
		tree[i] = tx.Txid //用现有的txid填充部分叶子节点
	}
	noneLeafOffset := leafSize //非叶子节点的插入点
	for i := 0; i < treeSize-1; i += 2 {
		switch {
		case tree[i] == nil: //没有左孩子
			tree[noneLeafOffset] = nil
		case tree[i+1] == nil: //没有右孩子
			concat := bytes.Join([][]byte{tree[i], tree[i]}, []byte{})
			tree[noneLeafOffset] = hash.DoubleSha256(concat)
		default: //左右都有
			concat := bytes.Join([][]byte{tree[i], tree[i+1]}, []byte{})
			tree[noneLeafOffset] = hash.DoubleSha256(concat)
		}
		noneLeafOffset++
	}
	return tree
}

//序列化系统合约失败的Txs
func encodeFailedTxs(buf *bytes.Buffer, block *pb.InternalBlock) error {
	txids := []string{}
	for txid := range block.FailedTxs {
		txids = append(txids, txid)
	}
	sort.Strings(txids) //ascii increasing order
	for _, txid := range txids {
		txErr := block.FailedTxs[txid]
		err := binary.Write(buf, binary.LittleEndian, []byte(txErr))
		if err != nil {
			return err
		}
	}
	return nil
}

func encodeJustify(buf *bytes.Buffer, block *pb.InternalBlock) error {
	if block.Justify == nil {
		// no justify field
		return nil
	}
	err := binary.Write(buf, binary.LittleEndian, block.Justify.ProposalId)
	if err != nil {
		return err
	}
	err = binary.Write(buf, binary.LittleEndian, block.Justify.ProposalMsg)
	if err != nil {
		return err
	}
	err = binary.Write(buf, binary.LittleEndian, block.Justify.Type)
	if err != nil {
		return err
	}
	err = binary.Write(buf, binary.LittleEndian, block.Justify.ViewNumber)
	if err != nil {
		return err
	}
	if block.Justify.SignInfos != nil {
		for _, sign := range block.Justify.SignInfos.QCSignInfos {
			err = binary.Write(buf, binary.LittleEndian, []byte(sign.Address))
			if err != nil {
				return err
			}
			err = binary.Write(buf, binary.LittleEndian, []byte(sign.PublicKey))
			if err != nil {
				return err
			}
			err = binary.Write(buf, binary.LittleEndian, sign.Sign)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// VerifyMerkle
func VerifyMerkle(block *pb.InternalBlock) error {
	blockid := block.Blockid
	merkleTree := MakeMerkleTree(block.Transactions)
	if len(merkleTree) > 0 {
		merkleRoot := merkleTree[len(merkleTree)-1]
		if !(bytes.Equal(merkleRoot, block.MerkleRoot)) {
			return errors.New("merkle root is wrong, block id:" + global.F(blockid) + ",block merkle root:" + global.F(block.MerkleRoot) + ", make merkle root:" + global.F(merkleRoot))
		}
		return nil
	} else {
		return errors.New("can not make merkle tree , block id:" + global.F(blockid))
	}
}

// MakeBlockID generate BlockID
func MakeBlockID(block *pb.InternalBlock) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, block.Version)
	if err != nil {
		return nil, err
	}
	err = binary.Write(buf, binary.LittleEndian, block.Nonce)
	if err != nil {
		return nil, err
	}
	err = binary.Write(buf, binary.LittleEndian, block.TxCount)
	if err != nil {
		return nil, err
	}
	if block.Proposer != nil {
		err = binary.Write(buf, binary.LittleEndian, block.Proposer)
		if err != nil {
			return nil, err
		}
	}
	err = binary.Write(buf, binary.LittleEndian, block.Timestamp)
	if err != nil {
		return nil, err
	}
	if block.Pubkey != nil {
		err = binary.Write(buf, binary.LittleEndian, block.Pubkey)
		if err != nil {
			return nil, err
		}
	}
	err = binary.Write(buf, binary.LittleEndian, block.PreHash)
	if err != nil {
		return nil, err
	}
	err = binary.Write(buf, binary.LittleEndian, block.MerkleRoot)
	if err != nil {
		return nil, err
	}
	err = encodeFailedTxs(buf, block)
	if err != nil {
		return nil, err
	}
	err = binary.Write(buf, binary.LittleEndian, block.CurTerm)
	if err != nil {
		return nil, err
	}
	err = binary.Write(buf, binary.LittleEndian, block.CurBlockNum)
	if err != nil {
		return nil, err
	}
	if block.TargetBits > 0 {
		err = binary.Write(buf, binary.LittleEndian, block.TargetBits)
		if err != nil {
			return nil, err
		}
	}
	err = encodeJustify(buf, block)
	if err != nil {
		return nil, fmt.Errorf("encodeJustify failed, err=%v", err)
	}
	return hash.DoubleSha256(buf.Bytes()), nil
}
