package utils

const (
	accountSize            = 16
	accountPrefix          = "XC"
	accountBucket          = "XCAccount"
	contractBucket         = "XCContract"
	contract2AccountBucket = "XCContract2Account"
	account2ContractBucket = "XCAccount2Contract"
	akLimit                = 1024
	aclSeperator           = "\x01"
	accountBcnameSep       = "@"
	accountContractValue   = "true"
	newAccountGasAmount    = 1000
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
	return accountName + aclSeperator + contractName
}

// MakeContractMethodKey generate contract and account mapping key
func MakeContractMethodKey(contractName string, methodName string) string {
	return contractName + aclSeperator + methodName
}

// MakeAccountKey generate account key using blockchain name and account number
func MakeAccountKey(bcname string, accountName string) string {
	return accountPrefix + accountName + accountBcnameSep + bcname
}

// GetAccountPrefix return the account prefix
func GetAccountPrefix() string {
	return accountPrefix
}

// GetAccountBucket return the account bucket name
func GetAccountBucket() string {
	return accountBucket
}

// GetACLSeperator return the acl seperator string
func GetACLSeperator() string {
	return aclSeperator
}

// GetAccountBcnameSep return the seperator string for account and blockchain name
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
