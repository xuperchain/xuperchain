package ledger

import (
	"bytes"
	"strconv"
	"strings"

	"github.com/xuperchain/xuperunion/kv/kvdb"
	"github.com/xuperchain/xuperunion/pb"
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
		key := string(it.Key())
		blockidStr := strings.Split(key, pb.BranchInfoPrefix)[1]
		value := string(it.Value())
		blockHeight, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, err
		}
		// only record block whose height is higher than target one
		if bytes.Equal(targetBlockid, []byte(blockidStr)) {
			continue
		}
		if blockHeight > targetBlockHeight {
			result = append(result, blockidStr)
		}
	}
	if it.Error() != nil {
		return nil, it.Error()
	}
	return result, nil
}
