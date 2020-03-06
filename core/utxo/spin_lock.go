package utxo

import "sync"
import "sort"
import "strconv"
import "github.com/xuperchain/xuperchain/core/pb"

const (
	SharedLock    = 1 //可共享的情况
	ExclusiveLock = 2 //互斥的情况
)

type RefCounter struct {
	ctMap map[string]int
	mu    sync.Mutex
}

type SpinLock struct {
	m          *sync.Map
	refCounter *RefCounter
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

func (rc *RefCounter) Add(key string) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.ctMap[key] += 1
}

func (rc *RefCounter) Release(key string) int {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.ctMap[key] -= 1
	return rc.ctMap[key]
}

func NewSpinLock() *SpinLock {
	return &SpinLock{m: &sync.Map{}, refCounter: &RefCounter{ctMap: map[string]int{}}}
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
			if lkType == SharedLock && k.lockType == SharedLock { //读读共享
				sp.refCounter.Add(k.key)
				succLocked = append(succLocked, k)
				continue
			} else {
				return succLocked, false //读写冲突
			}
		}
		if k.lockType == SharedLock {
			sp.refCounter.Add(k.key)
		}
		succLocked = append(succLocked, k) //第一个抢到
	}
	return succLocked, true
}

func (sp *SpinLock) Unlock(lockKeys []*LockKey) {
	N := len(lockKeys)
	for i := N - 1; i >= 0; i-- {
		lkType := lockKeys[i].lockType
		k := lockKeys[i].key
		if lkType == ExclusiveLock {
			sp.m.Delete(k)
		} else if lkType == SharedLock { //共享锁要考虑引用计数
			if sp.refCounter.Release(k) == 0 {
				sp.m.Delete(lockKeys[i].key)
			}
		}
	}
}
