package utxo

import "sync"
import "sort"
import "strconv"
import "github.com/xuperchain/xuperchain/core/pb"

const (
	SharedLock    = 1 //可共享的情况
	ExclusiveLock = 2 //互斥的情况
)

type SpinLock struct {
	m *sync.Map
}

type LockKey struct {
	lockType int
	key      string
}

func (lk *LockKey) String() string {
	if lk.lockType == SharedLock {
		return lk.key + ":S"
	} else if lk.lockType == ExclusiveLock {
		return lk.key + ":X"
	}
	return lk.key
}

func NewSpinLock() *SpinLock {
	return &SpinLock{m: &sync.Map{}}
}

func (sp *SpinLock) ExtractLockKeys(tx *pb.Transaction) []*LockKey {
	keys := []*LockKey{}
	for _, input := range tx.TxInputs {
		k := string(input.RefTxid) + "_" + strconv.Itoa(int(input.RefOffset))
		keys = append(keys, &LockKey{key: k, lockType: ExclusiveLock})
	}
	readKeys := map[string]bool{}
	writeKeys := map[string]bool{}
	for _, input := range tx.TxInputsExt {
		k := string(input.Bucket) + "/" + string(input.Key)
		readKeys[k] = true
	}
	for _, output := range tx.TxOutputsExt {
		k := string(output.Bucket) + "/" + string(output.Key)
		delete(readKeys, k)
		writeKeys[k] = true
	}
	for k := range readKeys {
		keys = append(keys, &LockKey{key: k, lockType: SharedLock})
	}
	for k := range writeKeys {
		keys = append(keys, &LockKey{key: k, lockType: ExclusiveLock})
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i].key < keys[j].key })
	lim := 0
	//dedup
	for i, k := range keys {
		if i == 0 || keys[i].key != keys[i-1].key {
			keys[lim] = k
			lim++
		}
	}
	return keys[:lim]
}

func (sp *SpinLock) TryLock(lockKeys []*LockKey) ([]*LockKey, bool) {
	succLocked := []*LockKey{}
	for _, k := range lockKeys {
		if lkType, occupiedByOthers := sp.m.LoadOrStore(k.key, k.lockType); occupiedByOthers {
			if lkType == SharedLock && k.lockType == SharedLock {
				succLocked = append(succLocked, k)
				continue
			} else {
				return succLocked, false
			}
		}
		succLocked = append(succLocked, k)
	}
	return succLocked, true
}

func (sp *SpinLock) Unlock(lockKeys []*LockKey) {
	N := len(lockKeys)
	for i := N - 1; i >= 0; i-- {
		sp.m.Delete(lockKeys[i].key)
	}
}
