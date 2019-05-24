package utxo

import (
	"container/list"
	"sync"
)

// UtxoCache is a in-memory cache of UTXO
type UtxoCache struct {
	// <ADDRESS, <UTXO_KEY, UTXO_ITEM>>
	Available map[string]map[string]*UtxoItem
	All       map[string]map[string]*UtxoItem
	List      *list.List
	nodes     map[string]*list.Element
	Limit     int
	mutex     *sync.Mutex
}

// NewUtxoCache create instance of UtxoCache
func NewUtxoCache(limit int) *UtxoCache {
	return &UtxoCache{
		Available: map[string]map[string]*UtxoItem{},
		All:       map[string]map[string]*UtxoItem{},
		List:      list.New(),
		nodes:     map[string]*list.Element{},
		Limit:     limit,
		mutex:     &sync.Mutex{},
	}
}

// Insert insert/update utxo cache
func (uv *UtxoCache) Insert(addr string, utxoKey string, item *UtxoItem) {
	uv.mutex.Lock()
	defer uv.mutex.Unlock()
	if _, exist := uv.All[addr]; !exist {
		uv.Available[addr] = map[string]*UtxoItem{}
		uv.All[addr] = map[string]*UtxoItem{}
	}
	uv.Available[addr][utxoKey] = item
	uv.All[addr][utxoKey] = item
	if node, ok := uv.nodes[addr]; ok {
		uv.List.MoveToFront(node) //挪到前面
	} else {
		ele := uv.List.PushFront(addr)
		uv.nodes[addr] = ele
	}
	if uv.List.Len() > uv.Limit {
		ele := uv.List.Back() //最近最少使用的address
		address := ele.Value.(string)
		delete(uv.Available, address)
		delete(uv.All, address)
		delete(uv.nodes, address)
		uv.List.Remove(ele)
	}
}

// Use mark a utxo key as used
func (uv *UtxoCache) Use(address string, utxoKey string) {
	if l2, exist := uv.Available[address]; exist {
		delete(l2, utxoKey)
	}
}

// Remove delete uxto key from cache
func (uv *UtxoCache) Remove(address string, utxoKey string) {
	uv.mutex.Lock()
	defer uv.mutex.Unlock()
	if l2, exist := uv.All[address]; exist {
		delete(l2, utxoKey)
		if len(l2) == 0 { //这个address的utxo删完了
			delete(uv.All, address)
			delete(uv.Available, address)
			if ele, ok := uv.nodes[address]; ok {
				delete(uv.nodes, address)
				uv.List.Remove(ele)
			}
		} else {
			if ele, ok := uv.nodes[address]; ok {
				uv.List.MoveToFront(ele) //挪到前面
			}
		}
	}
	if l2, exist := uv.Available[address]; exist {
		delete(l2, utxoKey)
	}
}

// Lock used to lock cache
func (uv *UtxoCache) Lock() {
	uv.mutex.Lock()
}

// Unlock used to unlock cache
func (uv *UtxoCache) Unlock() {
	uv.mutex.Unlock()
}
