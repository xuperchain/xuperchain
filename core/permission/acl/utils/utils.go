package utils

const (
	accountSize            = 16
	contractNameMaxSize    = 16
	contractNameMinSize    = 4
	accountPrefix          = "XC"
	accountBucket          = "XCAccount"
	contractBucket         = "XCContract"
	contract2AccountBucket = "XCContract2Account"
	account2ContractBucket = "XCAccount2Contract"
	ak2AccountBucket       = "XCAK2Account"
	akLimit                = 1024
	aclSeparator           = "\x01"
	accountBcnameSep       = "@"
	addressAccountSep      = "\x01"
	accountContractValue   = "true"
	newAccountGasAmount    = 1000
	ak2AccountValue        = "true"
)

// GetContract2AccountBucket get the bucket name of contract to account map
func GetContract2AccountBucket() string {
	return contract2AccountBucket
}

// GetAccount2ContractBucket get the bucket name of account to contract map
func GetAccount2ContractBucket() string {
	return account2ContractBucket
}

// MakeAccountContractKey generate account and contract mapping key
func MakeAccountContractKey(accountName string, contractName string) string {
	return accountName + aclSeparator + contractName
}

// MakeContractMethodKey generate contract and account mapping key
func MakeContractMethodKey(contractName string, methodName string) string {
	return contractName + aclSeparator + methodName
}

// MakeAccountKey generate account key using blockchain name and account number
func MakeAccountKey(bcname string, accountName string) string {
	return accountPrefix + accountName + accountBcnameSep + bcname
}

// MakeAK2AccountKey generate key mixed ak with account as prefix key
func MakeAK2AccountKey(ak string, accountName string) string {
	return ak + addressAccountSep + accountName
}

// GetAccountPrefix return the account prefix
func GetAccountPrefix() string {
	return accountPrefix
}

// GetAccountBucket return the account bucket name
func GetAccountBucket() string {
	return accountBucket
}

// GetACLSeparator return the acl separator string
func GetACLSeparator() string {
	return aclSeparator
}

// GetAKAccountSeparator return the separator between address and account
func GetAKAccountSeparator() string {
	return addressAccountSep
}

// GetAccountBcnameSep return the separator string for account and blockchain name
func GetAccountBcnameSep() string {
	return accountBcnameSep
}

// GetContractBucket return the contract bucket name
func GetContractBucket() string {
	return contractBucket
}

// GetAccountSize return the account number size
func GetAccountSize() int {
	return accountSize
}

// GetAkLimit return maximum AK numbers of an ACL
func GetAkLimit() int {
	return akLimit
}

// GetAccountContractValue return accountContractValue
func GetAccountContractValue() string {
	return accountContractValue
}

// GetContractNameMaxSize return the contract name max size
func GetContractNameMaxSize() int {
	return contractNameMaxSize
}

// GetContractNameMinSize return the contract name min size
func GetContractNameMinSize() int {
	return contractNameMinSize
}

// GetAK2AccountBucket return the ak2Account bucket
func GetAK2AccountBucket() string {
	return ak2AccountBucket
}
