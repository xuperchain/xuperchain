/*
Copyright Baidu Inc. All Rights Reserved.
*/

package main

import (
	"github.com/xuperchain/xuperchain/core/crypto/client/gm/gmclient"
)

func GetInstance() interface{} {
	gmCryptoClient := gmclient.GmCryptoClient{}
	return &gmCryptoClient
}
