package event

import (
	"regexp"

	"github.com/xuperchain/xuperchain/core/pb"
)

type txFilterFunc func(*blockFilter, *pb.Transaction) bool
type contractEventFilterFunc func(*blockFilter, *pb.ContractEvent) bool

var txFilterFuncs = []txFilterFunc{
	matchContractName,
	matchIterator,
	matchAuthRequire,
	matchFromAddr,
	matchToAddr,
}

var contractEventFilterFuncs = []contractEventFilterFunc{
	matchEventName,
}

type compiledBlockFilter struct {
	Contract    *regexp.Regexp
	EventName   *regexp.Regexp
	Initiator   *regexp.Regexp
	AuthRequire *regexp.Regexp
	FromAddr    *regexp.Regexp
	ToAddr      *regexp.Regexp
}

type blockFilter struct {
	*pb.BlockFilter
	compiled compiledBlockFilter
}

func newBlockFilter(ori *pb.BlockFilter) (*blockFilter, error) {
	var c compiledBlockFilter
	var err error
	if c.Contract, err = compileString(ori.GetContract()); err != nil {
		return nil, err
	}
	if c.EventName, err = compileString(ori.GetEventName()); err != nil {
		return nil, err
	}
	if c.Initiator, err = compileString(ori.GetInitiator()); err != nil {
		return nil, err
	}
	if c.AuthRequire, err = compileString(ori.GetAuthRequire()); err != nil {
		return nil, err
	}
	if c.FromAddr, err = compileString(ori.GetFromAddr()); err != nil {
		return nil, err
	}
	if c.ToAddr, err = compileString(ori.GetToAddr()); err != nil {
		return nil, err
	}

	return &blockFilter{
		BlockFilter: ori,
		compiled:    c,
	}, nil
}

func compileString(regstr string) (*regexp.Regexp, error) {
	if regstr == "" {
		return nil, nil
	}
	return regexp.Compile(regstr)
}

func matchString(filter *regexp.Regexp, target string) bool {
	return filter == nil || filter.MatchString(target)
}

func matchBytes(filter *regexp.Regexp, target []byte) bool {
	return filter == nil || filter.Match(target)
}

func matchIterator(filter *blockFilter, tx *pb.Transaction) bool {
	return matchString(filter.compiled.Initiator, tx.GetInitiator())
}

func matchAuthRequire(filter *blockFilter, tx *pb.Transaction) bool {
	if filter.GetAuthRequire() == "" {
		return true
	}
	for _, addr := range tx.GetAuthRequire() {
		if matchString(filter.compiled.AuthRequire, addr) {
			return true
		}
	}
	return false
}

func matchContractName(filter *blockFilter, tx *pb.Transaction) bool {
	if filter.GetContract() == "" {
		return true
	}
	for _, req := range tx.GetContractRequests() {
		if matchString(filter.compiled.Contract, req.GetContractName()) {
			return true
		}
	}
	return false
}

func matchFromAddr(filter *blockFilter, tx *pb.Transaction) bool {
	if len(filter.GetFromAddr()) == 0 {
		return true
	}
	for _, input := range tx.GetTxInputs() {
		if matchBytes(filter.compiled.FromAddr, input.GetFromAddr()) {
			return true
		}
	}
	return false
}

func matchToAddr(filter *blockFilter, tx *pb.Transaction) bool {
	if len(filter.GetToAddr()) == 0 {
		return true
	}
	for _, output := range tx.GetTxOutputs() {
		if matchBytes(filter.compiled.ToAddr, output.GetToAddr()) {
			return true
		}
	}
	return false
}

func matchEventName(filter *blockFilter, event *pb.ContractEvent) bool {
	return matchString(filter.compiled.EventName, event.GetName())
}

func matchTx(filter *blockFilter, tx *pb.Transaction) bool {
	for _, filterFunc := range txFilterFuncs {
		match := filterFunc(filter, tx)
		if !match {
			return false
		}
	}
	return true
}

func hasEventFilter(filter *blockFilter) bool {
	return filter.EventName != ""
}

func matchEvent(filter *blockFilter, event *pb.ContractEvent) bool {
	for _, filterFunc := range contractEventFilterFuncs {
		match := filterFunc(filter, event)
		if !match {
			return false
		}
	}
	return true
}
