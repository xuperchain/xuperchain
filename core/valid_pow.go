package xchaincore

import (
	"fmt"

	"github.com/xuperchain/xuperunion/common/log"
	"github.com/xuperchain/xuperunion/consensus"
	"github.com/xuperchain/xuperunion/ledger"
	"github.com/xuperchain/xuperunion/pb"
)

func ValidPowBlock(block *pb.Block, xcore *XChainCore) bool {
	internalBlock := block.GetBlock()
	if xcore == nil || block == nil || internalBlock == nil {
		log.Warn("invalid: xcore or block or internalBlock is nil")
		return false
	}

	// validation for consensus of pow, if ok, tell the miner to stop mining
	newBlockHeight := internalBlock.GetHeight()
	if xcore.con.Type(newBlockHeight) == consensus.ConsensusTypePow {
		if newBlockHeight < xcore.Ledger.GetMeta().GetTrunkHeight() {
			log.Warn("invalid block: new block's height is not enough", "new block's height->", newBlockHeight, "miner trunk height->", xcore.Ledger.GetMeta().GetTrunkHeight())
			return false
		}
		actualTargetBits := internalBlock.GetTargetBits()
		if !ledger.IsProofed(internalBlock.Blockid, actualTargetBits) {
			log.Warn("receive a new block actual difficulty doesn't match blockid", "blockid->", fmt.Sprintf("%x", internalBlock.GetBlockid()), "proposer->", internalBlock.GetProposer())
			return false
		}

		//xcore.Ledger.StartPowBlockState()
		// a valid new block shows up, let's interrupt the process of the miner to welcome it.
		xcore.Ledger.AbortPowMinning()
	}

	return true
}
