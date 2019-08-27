package ledger

import (
	"fmt"

	"github.com/xuperchain/xuperunion/pb"
)

func (l *Ledger) processFormatBlockForPOW(block *pb.InternalBlock, targetBits int32) (*pb.InternalBlock, error) {
	var gussNonce int32
	var gussCount int64
	valid := false
	var err error

	for {
		if gussCount%65535 == 0 && !l.GetPowBlockState() {
			break
		}
		if valid = IsProofed(block.Blockid, targetBits); !valid {
			gussNonce += 1
			block.Nonce = gussNonce
			block.Blockid, err = MakeBlockID(block)
			if err != nil {
				return nil, err
			}
			gussCount++
			continue
		}
		break
	}
	// valid为false说明还没挖到块
	// l.GetPowBlockState()为false说明被打断了
	// l.GetPowBlockState()为true说明还未被打断，此时valid不应该为false
	if !valid && !l.GetPowBlockState() {
		l.StartPowBlockState()
		l.xlog.Debug("I have been interrupted from a remote node, because it has a higher block")
		return nil, ErrTxDuplicated
	}
	l.xlog.Debug("I have generated a new block", "blockid->", fmt.Sprintf("%x", block.GetBlockid()))
	return block, nil
}
