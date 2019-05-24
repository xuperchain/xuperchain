/*
 * Copyright (c) 2019. Baidu Inc. All Rights Reserved.
 */

package common

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
