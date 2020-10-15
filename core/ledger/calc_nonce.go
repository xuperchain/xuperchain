package ledger

import (
	"fmt"

	"github.com/xuperchain/xuperchain/core/pb"
)

const round = 65535

var powMinningHeight = int64(0)

func (l *Ledger) processFormatBlockForPOW(block *pb.InternalBlock, targetBits int32) (*pb.InternalBlock, error) {
	var gussNonce int32
	var gussCount int64
	valid := false
	var err error
	// 在每次挖矿时，设置为true
	l.StartPowMinning(block.GetHeight())
	for {
		if gussCount%round == 0 && !l.IsEnablePowMinning() {
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
	// l.IsEnablePowMinning() == true  --> 自己挖出块
	// l.IsEnablePowMinning() == false --> 被中断
	if !valid && !l.IsEnablePowMinning() {
		l.xlog.Debug("I have been interrupted from a remote node, because it has a higher block")
		return nil, ErrMinerInterrupt
	}
	l.xlog.Debug("I have generated a new block", "blockid->", fmt.Sprintf("%x", block.GetBlockid()))
	return block, nil
}
