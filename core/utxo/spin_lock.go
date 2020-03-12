package utxo

import (
	"github.com/xuperchain/xuperchain/core/pb"
	"sort"
	"strconv"
	"sync"
)

const (
	sharedLock    = 1 //可共享的情况
	exclusiveLock = 2 //互斥的情况
)

type refCounter struct {
	ctMap map[string]int
	mu    sync.Mutex
}

//SpinLock is a collections of small locks on special keys
type SpinLock struct {
	m          *sync.Map
	refCounter *refCounter
}

// LockKey is a lock item with lock type and key
type LockKey struct {
	lockType int
	key      string
}

// String returns readable string for a lock item
func (lk *LockKey) String() string {
	if lk.lockType == sharedLock {
		return lk.key + ":S"
	} else if lk.lockType == exclusiveLock {
		return lk.key + ":X"
	}
	return lk.key
}

func (rc *refCounter) Add(key string) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.ctMap[key]++
}

func (rc *refCounter) Release(key string) int {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.ctMap[key]--
	return rc.ctMap[key]
}

// NewSpinLock returns a new spinlock instance
func NewSpinLock() *SpinLock {
	return &SpinLock{m: &sync.Map{}, refCounter: &refCounter{ctMap: map[string]int{}}}
}

//ExtractLockKeys extract lock items from a transaction
func (sp *SpinLock) ExtractLockKeys(tx *pb.Transaction) []*LockKey {
	keys := []*LockKey{}
	for _, input := range tx.TxInputs {
		k := string(input.RefTxid) + "_" + strconv.Itoa(int(input.RefOffset))
		keys = append(keys, &LockKey{key: k, lockType: exclusiveLock})
	}
	for offset := range tx.TxOutputs {
		k := string(tx.Txid) + "_" + strconv.Itoa(offset)
		keys = append(keys, &LockKey{key: k, lockType: exclusiveLock})
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
		keys = append(keys, &LockKey{key: k, lockType: sharedLock})
	}
	for k := range writeKeys {
		keys = append(keys, &LockKey{key: k, lockType: exclusiveLock})
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

//IsLocked returns whether a key is locked
func (sp *SpinLock) IsLocked(key string) bool {
	_, locked := sp.m.Load(key)
	return locked
}

//TryLock try to lock some keys
func (sp *SpinLock) TryLock(lockKeys []*LockKey) ([]*LockKey, bool) {
	succLocked := []*LockKey{}
	for _, k := range lockKeys {
		if lkType, occupiedByOthers := sp.m.LoadOrStore(k.key, k.lockType); occupiedByOthers {
			if lkType == sharedLock && k.lockType == sharedLock { //读读共享
				sp.refCounter.Add(k.key)
				succLocked = append(succLocked, k)
				continue
			} else {
				return succLocked, false //读写冲突
			}
		}
		if k.lockType == sharedLock {
			sp.refCounter.Add(k.key)
		}
		succLocked = append(succLocked, k) //第一个抢到
	}
	return succLocked, true
}

//Unlock release the locks on some keys
func (sp *SpinLock) Unlock(lockKeys []*LockKey) {
	N := len(lockKeys)
	for i := N - 1; i >= 0; i-- {
		lkType := lockKeys[i].lockType
		k := lockKeys[i].key
		if lkType == exclusiveLock {
			sp.m.Delete(k)
		} else if lkType == sharedLock { //共享锁要考虑引用计数
			if sp.refCounter.Release(k) == 0 {
				sp.m.Delete(lockKeys[i].key)
			}
		}
	}
}
