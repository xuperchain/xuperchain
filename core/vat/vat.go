// Package vat is Verifiable Autogen Tx package
package vat

import (
	"github.com/xuperchain/xuperunion/pb"
	"sort"
	"sync"
)

// VATInterface define the VAT interface
type VATInterface interface {
	GetVerifiableAutogenTx(blockHeight int64, maxCount int, timestamp int64) ([]*pb.Transaction, error)
	GetVATWhiteList() map[string]bool
}

// HandlerSlice the handler slice type
type HandlerSlice []string

// Len get the length of HandlerSlice
func (hs HandlerSlice) Len() int {
	return len(hs)
}

// Swap replace two elements in slice
func (hs HandlerSlice) Swap(i, j int) {
	hs[i], hs[j] = hs[j], hs[i]
}

// Less return true if element j is less than element i
func (hs HandlerSlice) Less(i, j int) bool {
	return hs[j] < hs[i]
}

// VATHandler define the VAT handler struct
type VATHandler struct {
	HandlerList HandlerSlice
	Handlers    map[string]VATInterface
	// map[module]map[method]bool
	WhiteList map[string]map[string]bool
	mutex     *sync.RWMutex
}

// NewVATHandler create instance of VATHandler
func NewVATHandler() *VATHandler {
	return &VATHandler{
		HandlerList: HandlerSlice{},
		Handlers:    make(map[string]VATInterface),
		WhiteList:   make(map[string]map[string]bool),
		mutex:       new(sync.RWMutex),
	}
}

// RegisterHandler add new handler into VATHandler
func (vh *VATHandler) RegisterHandler(name string, handler VATInterface, whiteList map[string]bool) {
	vh.mutex.Lock()
	defer vh.mutex.Unlock()
	if vh.Handlers[name] == nil {
		vh.Handlers[name] = handler
		vh.WhiteList[name] = whiteList
		vh.HandlerList = append(vh.HandlerList, name)
	} else {
		vh.Handlers[name] = handler
		vh.WhiteList[name] = whiteList
	}
	sort.Stable(vh.HandlerList)
}

// Remove delete handler of given name
func (vh *VATHandler) Remove(name string) {
	vh.mutex.Lock()
	defer vh.mutex.Unlock()
	delete(vh.Handlers, name)
	delete(vh.WhiteList, name)
	for i, v := range vh.HandlerList {
		if v == name {
			vh.HandlerList = append(vh.HandlerList[:i], vh.HandlerList[i+1:]...)
			break
		}
	}
}

// MustVAT check if the given module and method in whitelist
func (vh *VATHandler) MustVAT(module, method string) bool {
	if vh.WhiteList[module] == nil {
		return false
	}
	return vh.WhiteList[module][method]
}
