package xchaincore

import (
	"strings"
	"sync"
	"time"
)

const groupChainCacheUpdateWindow = 2

type GroupChainRegister interface {
	IsPeerInGroupChain(bcname, remotePeerID string) bool
	GetAllowedPeersWithBcname(bcname string) map[string]bool
}

type groupChainCache struct {
	// key: peerID value: ip + peerID
	StreamCache map[string]string
	// key: bcname value: map[string]bool
	StreamContractCache map[string]map[string]bool
	// key: bcname value: bool
	ChainContractCache map[string]bool
	Mutex              *sync.Mutex
}

// IsPeerInGroupChain 判断某条链下的某个节点是否在白名单中
func (xm *XChainMG) IsPeerInGroupChain(bcname, remotePeerID string) bool {
	if bcname == "" {
		return true
	}
	// 判断bcname是否支持群组
	xm.groupChainCache.Mutex.Lock()
	defer xm.groupChainCache.Mutex.Unlock()
	if _, groupExist := xm.groupChainCache.ChainContractCache[bcname]; !groupExist {
		return true
	}
	ipSet, ipSetExist := xm.groupChainCache.StreamContractCache[bcname]
	// ipSetExist代表是否有群组属性
	// len(ipSet)代表bcname的白名单数量
	if !ipSetExist {
		return true
	} else if len(ipSet) == 0 {
		return false
	}

	// 如果本地没有远程传来的节点id，直接拒绝
	ip, ipExist := xm.groupChainCache.StreamCache[remotePeerID]
	if !ipExist {
		return false
	}
	if _, exist := ipSet[ip]; !exist {
		return false
	}

	return true
}

func (xm *XChainMG) GetAllowedPeersWithBcname(bcname string) map[string]bool {
	allowedPeersMap := map[string]bool{}
	if bcname == "" {
		return allowedPeersMap
	}

	xm.groupChainCache.Mutex.Lock()
	defer xm.groupChainCache.Mutex.Unlock()
	if _, groupExist := xm.groupChainCache.ChainContractCache[bcname]; !groupExist {
		return allowedPeersMap
	}
	ipSet, exist := xm.groupChainCache.StreamContractCache[bcname]
	if !exist {
		return allowedPeersMap
	}
	for ip, _ := range ipSet {
		arr := strings.Split(ip, "/")
		if len(arr) <= 0 {
			continue
		}
		peerID := arr[len(arr)-1]
		localIP, exist := xm.groupChainCache.StreamCache[peerID]
		// 本地不存在，忽略
		if !exist {
			continue
		}
		// 群组合约存储的ip与本地ip一致，该stream是有效的
		if localIP == ip {
			allowedPeersMap[peerID] = true
		}
	}
	if len(allowedPeersMap) == 0 {
		allowedPeersMap["MAGIC_PEERID"] = true
	}

	return allowedPeersMap
}

func (xm *XChainMG) updateContractCache() {
	bc := xm.Get("xuper")
	if bc == nil {
		return
	}
	chainRes := bc.Utxovm.QueryChainInList()
	xm.groupChainCache.Mutex.Lock()
	xm.groupChainCache.ChainContractCache = chainRes

	bcnameSet := []string{}
	for bcname, _ := range xm.groupChainCache.ChainContractCache {
		bcnameSet = append(bcnameSet, bcname)
	}
	xm.groupChainCache.Mutex.Unlock()

	for _, bcname := range bcnameSet {
		ipSet := bc.Utxovm.QueryIPsInList(bcname)
		xm.groupChainCache.Mutex.Lock()
		xm.groupChainCache.StreamContractCache[bcname] = ipSet
		xm.groupChainCache.Mutex.Unlock()
	}
}

func (xm *XChainMG) updateStreamCache() {
	data := xm.P2pSvr.GetPeerIDAndUrls()
	// key: peerID, value: ip+peerID
	xm.groupChainCache.Mutex.Lock()
	defer xm.groupChainCache.Mutex.Unlock()
	xm.groupChainCache.StreamCache = data
}

func (xm *XChainMG) updateGroupChainCache() {
	for {
		xm.updateContractCache()
		xm.updateStreamCache()
		time.Sleep(groupChainCacheUpdateWindow * time.Second)
	}
}
