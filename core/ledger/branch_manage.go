package ledger

import (
	"bytes"
	"strconv"

	"github.com/xuperchain/xuperchain/core/kv/kvdb"
	"github.com/xuperchain/xuperchain/core/pb"
)

func (l *Ledger) updateBranchInfo(addedBlockid, deletedBlockid []byte, addedBlockHeight int64, batch kvdb.Batch) error {
	// delete deletedBlockid
	err := batch.Delete(append([]byte(pb.BranchInfoPrefix), deletedBlockid...))
	if err != nil {
		return err
	}
	// put addedBlockid
	addedBlockHeightStr := strconv.FormatInt(addedBlockHeight, 10)
	err = batch.Put(append([]byte(pb.BranchInfoPrefix), addedBlockid...), []byte(addedBlockHeightStr))
	if err != nil {
		return err
	}
	return nil
}

func (l *Ledger) GetBranchInfo(targetBlockid []byte, targetBlockHeight int64) ([]string, error) {
	result := []string{}
	it := l.baseDB.NewIteratorWithPrefix([]byte(pb.BranchInfoPrefix))
	defer it.Release()

	for it.Next() {
		key := it.Key()
		if len(key) < len(pb.BranchInfoPrefix)+1 {
			// 理论上不会出现这种情况
			continue
		}

		// key格式为:前缀+blockid，去掉前缀
		blkId := key[len(pb.BranchInfoPrefix):]

		value := string(it.Value())
		blockHeight, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, err
		}

		// only record block whose height is higher than target one
		if bytes.Equal(targetBlockid, blkId) {
			continue
		}
		if blockHeight > targetBlockHeight {
			result = append(result, string(blkId))
		}
	}
	if it.Error() != nil {
		return nil, it.Error()
	}

	return result, nil
}

func (l *Ledger) HandleFork(oldTip []byte, newTip []byte, batchWrite kvdb.Batch) (*pb.InternalBlock, error) {
	commonParentBlockid, err := l.GetCommonParentBlockid(oldTip, newTip)
	if err != nil {
		return nil, err
	}
	// 将老分支剪一下
	for !bytes.Equal(oldTip, commonParentBlockid) {
		oldBlock, oldBlockErr := l.fetchBlock(oldTip)
		if oldBlockErr != nil {
			return nil, oldBlockErr
		}
		oldBlock.InTrunk = false
		oldBlock.NextHash = []byte{}
		oldTip = oldBlock.PreHash
		saveErr := l.saveBlock(oldBlock, batchWrite)
		if saveErr != nil {
			return nil, saveErr
		}
	}
	// 将新分支修一下
	newBlock, newBlockErr := l.fetchBlock(newTip)
	if newBlockErr != nil {
		return nil, newBlockErr
	}
	newPreBlockid := newBlock.PreHash
	nextHash := []byte{}
	for !bytes.Equal(newTip, commonParentBlockid) {
		newBlock, newBlockErr := l.fetchBlock(newTip)
		if newBlockErr != nil {
			return nil, newBlockErr
		}
		newBlock.InTrunk = true
		cerr := l.correctTxsBlockid(newBlock.Blockid, batchWrite)
		if cerr != nil {
			return nil, cerr
		}
		newBlock.NextHash = nextHash
		nextHash = newBlock.Blockid
		saveErr := l.saveBlock(newBlock, batchWrite)
		if saveErr != nil {
			return nil, saveErr
		}
		newTip = nextHash
	}
	return l.fetchBlock(newPreBlockid)
}

func (l *Ledger) GetCommonParentBlockid(branch1Blockid, branch2Blockid []byte) ([]byte, error) {
	branch1Block, branch1Err := l.QueryBlock(branch1Blockid)
	if branch1Err != nil {
		return nil, branch1Err
	}
	branch2Block, branch2Err := l.QueryBlock(branch2Blockid)
	if branch2Err != nil {
		return nil, branch2Err
	}
	branch1BlockHeight := branch1Block.Height
	branch2BlockHeight := branch2Block.Height
	for branch1BlockHeight > branch2BlockHeight {
		branch1Block, branch1Err = l.QueryBlock(branch1Block.PreHash)
		// Is it necessary to consider about not found?
		if branch1Err != nil {
			return nil, branch1Err
		}
		branch1BlockHeight = branch1Block.Height
	}
	for branch2BlockHeight > branch1BlockHeight {
		branch2Block, branch2Err = l.QueryBlock(branch2Block.PreHash)
		if branch2Err != nil {
			return nil, branch2Err
		}
		branch2BlockHeight = branch2Block.Height
	}
	for !bytes.Equal(branch1Block.Blockid, branch2Block.Blockid) {
		branch1Block, branch1Err = l.QueryBlock(branch1Block.PreHash)
		if branch1Err != nil {
			return nil, branch1Err
		}
		branch2Block, branch2Err = l.QueryBlock(branch2Block.PreHash)
		if branch2Err != nil {
			return nil, branch2Err
		}
	}
	return branch1Block.Blockid, nil
}

func (l *Ledger) SetMeta(meta *pb.LedgerMeta) {
	l.meta = meta
}

func (l *Ledger) RemoveBlocks(fromBlockid []byte, toBlockid []byte, batch kvdb.Batch) error {
	return l.removeBlocks(fromBlockid, toBlockid, batch)
}
