package ledger

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/xuperchain/xuperunion/kv/kvdb"
	"github.com/xuperchain/xuperunion/pb"
)

func (l *Ledger) updateBranchInfo(addedBlockid, deletedBlockid []byte, addedBlockHeight int64, batch kvdb.Batch) error {
	addedBlockidStr := fmt.Sprintf("%x", addedBlockid)
	deletedBlockidStr := fmt.Sprintf("%x", deletedBlockid)
	// 删除preBlockid
	err := batch.Delete(append([]byte(pb.BranchInfoPrefix + deletedBlockidStr)))
	if err != nil {
		return err
	}
	// 更新addedBlockid
	addedBlockHeightStr := strconv.FormatInt(addedBlockHeight, 10)
	err = batch.Put(append([]byte(pb.BranchInfoPrefix+addedBlockidStr)), []byte(addedBlockHeightStr))
	return err
}

func (l *Ledger) GetTargetRangeBranchInfo(targetBlockidStr string, targetBlockHeight int64) ([]string, error) {
	result := []string{}
	iter := l.baseDB.NewIteratorWithPrefix([]byte(pb.BranchInfoPrefix))
	defer iter.Release()
	for iter.Next() {
		key := string(iter.Key())
		blockidStr := strings.Split(key, pb.BranchInfoPrefix)[0]
		value := string(iter.Value())
		blockHeight, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, err
		}
		// only record block whose height is higher than target one
		if blockidStr == targetBlockidStr {
			continue
		}
		if blockHeight > targetBlockHeight {
			result = append(result, blockidStr)
		}
	}
	if iter.Error() != nil {
		return nil, iter.Error()
	}
	return result, nil
}

func (l *Ledger) GetCommonParentBlockid(child1BlockidStr, child2BlockidStr []byte) ([]byte, error) {
	child1Block, child1Err := l.QueryBlock([]byte(child1BlockidStr))
	if child1Err != nil {
		return nil, child1Err
	}
	child2Block, child2Err := l.QueryBlock([]byte(child2BlockidStr))
	if child2Err != nil {
		return nil, child2Err
	}
	child1BlockHeight := child1Block.Height
	child2BlockHeight := child2Block.Height
	if child1BlockHeight > child2BlockHeight {
		for child1BlockHeight > child2BlockHeight {
			child1Block, child1Err = l.fetchBlock(child1Block.PreHash)
			if child1Err != nil {
				return nil, child1Err
			}
			child1BlockHeight = child1Block.Height
		}
	} else if child1BlockHeight < child2BlockHeight {
		for child2BlockHeight > child1BlockHeight {
			child2Block, child2Err = l.fetchBlock(child2Block.PreHash)
			if child2Err != nil {
				return nil, child2Err
			}
			child2BlockHeight = child2Block.Height
		}
	}
	for !bytes.Equal(child1Block.Blockid, child2Block.Blockid) {
		child1Block, child1Err = l.fetchBlock(child1Block.PreHash)
		if child1Err != nil {
			return nil, child1Err
		}
		child2Block, child2Err = l.fetchBlock(child2Block.PreHash)
		if child2Err != nil {
			return nil, child2Err
		}
	}
	return child1Block.Blockid, nil
}
