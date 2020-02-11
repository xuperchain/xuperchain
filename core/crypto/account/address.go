/*
Copyright Baidu Inc. All Rights Reserved.
*/

package account

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"fmt"
	"reflect"

	"github.com/xuperchain/xuperchain/core/crypto/config"
	"github.com/xuperchain/xuperchain/core/crypto/hash"

	"github.com/btcsuite/btcutil/base58"
)

// GetAddressFromPublicKey 返回33位长度的地址
func GetAddressFromPublicKey(pub *ecdsa.PublicKey) (string, error) {
	if pub.X == nil || pub.Y == nil {
		return "", fmt.Errorf("public field is missing, pub.X:[%v], pub.Y:[%v]", pub.X, pub.Y)
	}
	//using SHA256 and Ripemd160 for hash summary
	data := elliptic.Marshal(pub.Curve, pub.X, pub.Y)
	outputSha256 := hash.UsingSha256(data)
	OutputRipemd160 := hash.UsingRipemd160(outputSha256)
	//暂时只支持一个字节长度，也就是uint8的密码学标志位
	// 判断是否是nist标准的私钥
	nVersion := config.Nist

	switch pub.Params().Name {
	case config.CurveNist: // NIST
	case config.CurveGm: // 国密
		nVersion = config.Gm
	case config.CurveNistSN: // NIST+Schnorr
		nVersion = config.NistSN
	default: // 不支持的密码学类型
		return "", fmt.Errorf("This cryptography[%v] has not been supported yet", pub.Params().Name)
	}

	bufVersion := []byte{byte(nVersion)}

	strSlice := make([]byte, len(bufVersion)+len(OutputRipemd160))
	copy(strSlice, bufVersion)
	copy(strSlice[len(bufVersion):], OutputRipemd160)

	//using double SHA256 for future risks
	checkCode := hash.DoubleSha256(strSlice)
	simpleCheckCode := checkCode[:4]

	slice := make([]byte, len(strSlice)+len(simpleCheckCode))
	copy(slice, strSlice)
	copy(slice[len(strSlice):], simpleCheckCode)

	//使用base58编码，手写不容易出错。
	//相比Base64，Base58不使用数字"0"，字母大写"O"，字母大写"I"，和字母小写"l"，以及"+"和"/"符号。
	strEnc := base58.Encode(slice)

	return strEnc, nil
}

// VerifyAddressUsingPublicKey 验证钱包地址是否和指定的公钥match
// 如果成功，返回true和对应的密码学标记位；如果失败，返回false和默认的密码学标记位0
func VerifyAddressUsingPublicKey(address string, pub *ecdsa.PublicKey) (bool, uint8) {
	//base58反解回byte[]数组
	slice := base58.Decode(address)

	//检查是否是合法的base58编码
	if len(slice) < 1 {
		return false, 0
	}
	//拿到密码学标记位
	byteVersion := slice[:1]
	nVersion := uint8(byteVersion[0])

	realAddress, err := GetAddressFromPublicKey(pub)
	if err != nil {
		return false, 0
	}

	if realAddress == address {
		return true, nVersion
	}

	return false, 0
}

// CheckAddressFormat 验证钱包地址是否是合法的格式
// 如果成功，返回true和对应的密码学标记位；如果失败，返回false和默认的密码学标记位0
func CheckAddressFormat(address string) (bool, uint8) {
	//base58反解回byte[]数组
	slice := base58.Decode(address)

	//检查是否是合法的base58编码
	if len(slice) < 1 {
		return false, 0
	}
	//拿到简单校验码
	simpleCheckCode := slice[len(slice)-4:]

	checkContent := slice[:len(slice)-4]
	checkCode := hash.DoubleSha256(checkContent)
	realSimpleCheckCode := checkCode[:4]

	byteVersion := slice[:1]
	nVersion := uint8(byteVersion[0])

	if reflect.DeepEqual(realSimpleCheckCode, simpleCheckCode) {
		return true, nVersion
	}

	return false, 0
}
