package ledger

import (
	"strconv"

	"github.com/xuperchain/xuperunion/kv/kvdb"
	"github.com/xuperchain/xuperunion/pb"
)

func (l *Ledger) updateBranchInfo(addedBlockid, deletedBlockid []byte, addedBlockHeight int64, batch kvdb.Batch) error {
	// 删除preBlockid
	err := batch.Delete(append([]byte(pb.BranchInfoPrefix), deletedBlockid...))
	if err != nil {
		return err
	}
	// 更新addedBlockid
	addedBlockHeightStr := strconv.FormatInt(addedBlockHeight, 10)
	err = batch.Put(append([]byte(pb.BranchInfoPrefix), addedBlockid...), []byte(addedBlockHeightStr))
	return err
}
