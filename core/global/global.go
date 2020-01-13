package global

const (
	// SafeModel 表示安全的同步
	SafeModel = iota
	// Normal 表示正常状态
	Normal
)

const (
	// SRootChainName name of xuper chain
	SRootChainName = "xuper"
	// SBlockChainConfig configuration file name of xuper chain
	SBlockChainConfig = "xuper.json"
)

// XContext define the common context
type XContext struct {
	Timer *XTimer
}
