package bridge

import xchainpb "github.com/xuperchain/xuperchain/core/pb"

// TODO:zq

const (
	// XuperScheme define the xuper scheme
	XuperScheme = "xuper"
)

// CrossChainURI Standard
// [scheme:][//chain_name][path][?query]
// eg xuper://chain1/wasm?contract_name=counter&method_name=increase
type CrossChainURI struct {
	scheme    string
	chainName string
	path      string
	query     string
}

// ParseCrossChainURI will parse uri to cross chain uri instance
func ParseCrossChainURI(crossChainURI string) (*CrossChainURI, error) {
	return nil, nil
}

// GetScheme return cross chain uri scheme
func (ccu *CrossChainURI) GetScheme() string {
	return ""
}

// GetChainName return cross chain uri chain name
func (ccu *CrossChainURI) GetChainName() string {
	return ""
}

// GetPath return cross chain uri path
func (ccu *CrossChainURI) GetPath() string {
	return ""
}

// GetQuery return cross chain uri query
func (ccu *CrossChainURI) GetQuery() string {
	return ""
}

// CrossChainScheme define the interface of CrossChainScheme
type CrossChainScheme interface {
	GetCrossQueryRequest(*CrossChainURI) (*xchainpb.CrossQueryRequest, error)
}

// CrossXuperScheme define the xuper scheme
type CrossXuperScheme struct {
}

// GetCrossQueryRequest return XupeScheme instance with CrossChainURI
func (cxs *CrossXuperScheme) GetCrossQueryRequest(crossChainURI *CrossChainURI) (*xchainpb.CrossQueryRequest, error) {
	return nil, nil
}
