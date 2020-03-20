package utils

import (
	"bytes"
	"errors"
	"reflect"
)

var (
	// 密码错误,解密失败
	ErrPwWrong = errors.New("password is wrong")
)

//BytesCombine concatenates byte array
func BytesCombine(pBytes ...[]byte) []byte {
	var buffer bytes.Buffer

	for i := 0; i < len(pBytes); i++ {
		buffer.Write(pBytes[i])
	}

	return buffer.Bytes()
}

//BytesCompare compares two byte arrays, give the result whether they are the same or not
func BytesCompare(bytesA, bytesB []byte) bool {
	if reflect.DeepEqual(bytesA, bytesB) {
		return true
	}
	return false
}

// 把slice的长度补齐到指定字节的长度
func BytesPad(pBytes []byte, length int) []byte {
	newSlice := make([]byte, length-len(pBytes))
	return append(newSlice, pBytes...)
}

// PKCS7Padding:缺N个字节就补N个字节的0，
// PKCS5Padding:缺N个字节就补充N个字节的N，例如缺8个字节，就补充8个字节的数字8
func BytesPKCS5Padding(cipherData []byte, blockSize int) []byte {
	padLength := blockSize - len(cipherData)%blockSize
	padData := bytes.Repeat([]byte{byte(padLength)}, padLength)

	return append(cipherData, padData...)
}

// PKCS7Padding:缺N个字节就补N个字节的0，
// PKCS5Padding:缺N个字节就补充N个字节的N，例如缺8个字节，就补充8个字节的数字8
// PKCS5UnPadding:获取最后一个字节，转换为数字N，然后剔除掉最后N个字节，例如最后一个字节是数字8，就剔除掉最后8个字节
func BytesPKCS5UnPadding(originalData []byte) ([]byte, error) {
	dataLength := len(originalData)
	unpadLength := int(originalData[dataLength-1])

	// 无法按照PKCS5UnPadding来正确解密
	if dataLength-unpadLength <= 0 {
		return nil, ErrPwWrong
	}

	return originalData[:(dataLength - unpadLength)], nil
}

/**
// 比较两个字节数组的内容是否完全一致
func BytesCompare(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
**/
