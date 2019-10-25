package pb

// common definition for KV prefix
const (
	BlocksTablePrefix        = "B" //表名prefix必须用一个字母，否则拼key在compare的时候会和后面的后缀无法区分开
	UTXOTablePrefix          = "U"
	UnconfirmedTablePrefix   = "N"
	ConfirmedTablePrefix     = "C"
	MetaTablePrefix          = "M"
	EVMMetaStatePrefix       = "S"
	EVMOutputPrefix          = "O"
	TriggerPrefix            = "T"
	VoteProposalPrefix       = "V"
	PlugConsPrefix           = "P"
	ConsTDposPrefix          = "D"
	PendingBlocksTablePrefix = "E"
	TxExtensionPrefix        = "X"
	WithdrawPrefix           = "W"
	ExtUtxoTablePrefix       = "ZU"
	ExtUtxoDelTablePrefix    = "ZD"
	BlockHeightPrefix        = "ZH"
	BranchInfoPrefix         = "ZI"
)
