// Package utils defines some crypto/signature utilities
package utils

import (
	"crypto/aes"
	"encoding/base64"
	"github.com/xuperchain/xuperchain/core/crypto/aes/ecb"
	"github.com/xuperchain/xuperchain/core/crypto/aes/padding"
)

// AESEncrypt encrypt plaint text to encrypted text using key
func AESEncrypt(pt, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	mode := ecb.NewECBEncrypter(block)
	padder := padding.NewPkcs7Padding(mode.BlockSize()) //padding方式兼容php
	pt, err = padder.Pad(pt)
	if err != nil {
		return nil, err
	}
	ct := make([]byte, len(pt))
	mode.CryptBlocks(ct, pt)
	return ct, nil
}

// AESDecrypt decrypt encrypted text to plaint text using key
func AESDecrypt(ct, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	mode := ecb.NewECBDecrypter(block)
	pt := make([]byte, len(ct))
	mode.CryptBlocks(pt, ct)
	padder := padding.NewPkcs7Padding(mode.BlockSize()) //兼容php
	pt, err = padder.Unpad(pt)
	if err != nil {
		return nil, err
	}
	return pt, nil
}

// AESEncryptHex encrypt plaint text to base64-encoded encrypted text using key
func AESEncryptHex(pt, key []byte) (string, error) {
	ct, err := AESEncrypt(pt, key)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(ct), nil
}

// AESDecryptHex decrypt base64-encoded encrypted text to plaint text using key
func AESDecryptHex(ctHex string, key []byte) ([]byte, error) {
	ct, err := base64.StdEncoding.DecodeString(ctHex)
	if err != nil {
		return nil, err
	}
	return AESDecrypt(ct, key)
}
