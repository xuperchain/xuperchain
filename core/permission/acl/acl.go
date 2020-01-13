package acl

import (
	"fmt"
	"strings"

	"github.com/xuperchain/xuperchain/core/permission/acl/utils"
)

// IsAccount check the type of name
// return : -1 if name is invalid
//           1 if name is account
//           0 if name is AK
func IsAccount(name string) int {
	if name == "" {
		return -1
	}
	if !strings.HasPrefix(name, utils.GetAccountPrefix()) {
		return 0
	}
	prefix := strings.Split(name, utils.GetAccountBcnameSep())[0]
	prefix = prefix[len(utils.GetAccountPrefix()):]
	if err := ValidRawAccount(prefix); err != nil {
		return 0
	}
	return 1
}

// ValidRawAccount validate account number
func ValidRawAccount(accountName string) error {
	// param absence check
	if accountName == "" {
		return fmt.Errorf("invoke NewAccount failed, account name is empty")
	}
	// account naming rule check
	if len(accountName) != utils.GetAccountSize() {
		return fmt.Errorf("invoke NewAccount failed, account name length expect %d, actual: %d", utils.GetAccountSize(), len(accountName))
	}
	for i := 0; i < utils.GetAccountSize(); i++ {
		if accountName[i] >= '0' && accountName[i] <= '9' {
			continue
		} else {
			return fmt.Errorf("invoke NewAccount failed, account name expect continuous %d number", utils.GetAccountSize())
		}
	}
	return nil
}
