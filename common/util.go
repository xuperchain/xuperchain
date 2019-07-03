/*
 * Copyright (c) 2019. Baidu Inc. All Rights Reserved.
 */

package common

import (
	"fmt"
	"regexp"

	"github.com/xuperchain/xuperunion/permission/acl/utils"
)

var (
	contractNameRegex = regexp.MustCompile("^[a-zA-Z_]{1}[0-9a-zA-Z_.]+[0-9a-zA-Z_]$")
)

// UniqSlice de-duplication function `
func UniqSlice(slice []string) []string {
	var res []string
	tempMap := make(map[string]byte)
	for _, v := range slice {
		l := len(tempMap)
		tempMap[v] = 0
		if len(tempMap) != l {
			res = append(res, v)
		}
	}
	return res
}

func validContractName(contractName string) error {
	// param absence check
	// contract naming rule check
	contractSize := len(contractName)
	contractMaxSize := utils.GetContractNameMaxSize()
	contractMinSize := utils.GetContractNameMinSize()
	if contractSize > contractMaxSize || contractSize < contractMinSize {
		return fmt.Errorf("contract name length expect [%d~%d], actual: %d", contractMinSize, contractMaxSize, contractSize)
	}
	if !contractNameRegex.MatchString(contractName) {
		return fmt.Errorf("contract name does not fit the rule of contract name")
	}
	return nil
}

// ValidContractName check if contract name is ok
func ValidContractName(contractName string) error {
	return validContractName(contractName)
}
