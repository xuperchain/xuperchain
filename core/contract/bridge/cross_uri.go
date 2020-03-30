package bridge

import (
	"fmt"
	"net/url"

	pb "github.com/xuperchain/xuperchain/core/contractsdk/go/pb"
	xchainpb "github.com/xuperchain/xuperchain/core/pb"
)

const (
	// XuperScheme define the xuper scheme
	XuperScheme = "xuper"
)

// CrossChainURI Standard
// [scheme:][//chain_name][path][?query]
// eg xuper://chain1?module=wasm&bcname=xuper&contract_name=counter&method_name=increase
type CrossChainURI struct {
	*url.URL
}

// ParseCrossChainURI will parse uri to cross chain uri instance
func ParseCrossChainURI(crossChainURI string) (*CrossChainURI, error) {
	uri, err := url.Parse(crossChainURI)
	if err != nil {
		return nil, err
	}

	return &CrossChainURI{
		uri,
	}, nil
}

// GetScheme return cross chain uri scheme
func (ccu *CrossChainURI) GetScheme() string {
	return ccu.URL.Scheme
}

// GetChainName return cross chain uri chain name
func (ccu *CrossChainURI) GetChainName() string {
	return ccu.URL.Host
}

// GetPath return cross chain uri path
func (ccu *CrossChainURI) GetPath() string {
	return ccu.URL.Path
}

// GetQuery return cross chain uri query
func (ccu *CrossChainURI) GetQuery() url.Values {
	return ccu.URL.Query()
}

// CrossChainScheme define the interface of CrossChainScheme
type CrossChainScheme interface {
	GetName() string
	GetCrossQueryRequest(*CrossChainURI, []*pb.ArgPair, string, []string) (*xchainpb.CrossQueryRequest, error)
}

// CrossXuperScheme define the xuper scheme
type CrossXuperScheme struct {
}

// GetCrossQueryRequest return XupeScheme instance with CrossChainURI
// [scheme:][//chain_name][?query]
// eg xuper://chain1?module=wasm&bcname=xuper&contract_name=counter&method_name=increase
func (cxs *CrossXuperScheme) GetCrossQueryRequest(crossChainURI *CrossChainURI,
	argPair []*pb.ArgPair, initiator string, authRequire []string) (*xchainpb.CrossQueryRequest, error) {
	if initiator == "" {
		return nil, fmt.Errorf("GetCrossQueryRequest initiator is nil")
	}

	querys := crossChainURI.GetQuery()
	module := querys.Get("module")
	bcname := querys.Get("bcname")
	contractName := querys.Get("contract_name")
	methodName := querys.Get("method_name")
	if module == "" || bcname == "" || contractName == "" || methodName == "" {
		return nil, fmt.Errorf("GetCrossQueryRequest query is nil")
	}
	args := make(map[string][]byte)
	for _, arg := range argPair {
		args[arg.GetKey()] = arg.GetValue()
	}

	crossQueryRequest := &xchainpb.CrossQueryRequest{
		Bcname:      bcname,
		Initiator:   initiator,
		AuthRequire: authRequire,
		Request: &xchainpb.InvokeRequest{
			ModuleName:   module,
			ContractName: contractName,
			MethodName:   methodName,
			Args:         args,
		},
	}
	return crossQueryRequest, nil
}

// GetName return cross xuper scheme name
func (cxs *CrossXuperScheme) GetName() string {
	return XuperScheme
}

// GetChainScheme return chain scheme by scheme
func GetChainScheme(scheme string) CrossChainScheme {
	switch scheme {
	case XuperScheme:
		return &CrossXuperScheme{}
	default:
		return &CrossXuperScheme{}
	}
}
