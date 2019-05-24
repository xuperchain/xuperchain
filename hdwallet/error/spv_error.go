// Package spv_error is hdwallet error declaration
// Copyright (c) 2017 Baidu.com, Inc. All Rights Reserved

package spverror

import (
	"errors"
	"fmt"

	"github.com/xuperchain/xuperunion/pb"
)

// 定义全局errors
var (
	// 账户相关的错误
	// 账户文件不存在
	ErrFileNotExist = errors.New("file not exist")
	// 参数错误
	ErrParam = errors.New("param is illegal")
	// 请求账户管理系统失败
	ErrRequestFailed = errors.New("http request failed")
	// 请求账户管理系统返回格式错误
	ErrReadResponseFailed = errors.New("http read response failed")
	// 请求账户管理系统返回错误
	ErrResponseFailed = errors.New("http response failed")
	// 请求账户管理系统返回参数错误
	ErrRequestParam = errors.New("http request param is illegal")
	// 账户管理系统账户不存在
	ErrAccountNotExist = errors.New("account not exist")
	// 账户管理系统账户支付密码不存在
	ErrPwNotExist = errors.New("password not exist")
	// 账户管理系统账户支付密码已存在
	ErrPwExist = errors.New("password has been existed")
	// 账户管理系统返回数据库操作失败
	ErrDbFail = errors.New("Db failed")
	// 用户没有登录百度帐号
	ErrNotLogin = errors.New("user is not login")
	// 密码错误,解密失败
	ErrPwWrong = errors.New("password is wrong")
	// 交易相关错误
	// 无法连接到指定的全节点
	ErrConnectNodeFailure = errors.New("Cannot connect to the specified node")
	// 没有足够的UTXO
	ErrNotEnoughUTXO = errors.New("Cannot get enough utxo")
	// 不合法的钱包地址
	ErrInvalidAddress = errors.New("Address is invalid")
	// 不合法的手续费格式，手续费需要是正数
	ErrInvalidFeeAmount = errors.New("fee amount is invalid")
	// 链接拒绝
	ErrConnectRefuce = errors.New("Connect refuced by server")
	// 加密错误
	ErrUtxoEncrypt = errors.New("Utxo list encrypt error")
	// fee 不够
	ErrFeeNotEnough = errors.New("Fee is not enough")
	// 参数不合法
	ErrValidateParams = errors.New("Param not validate")
	// tx 签名验证失败
	ErrTxSign = errors.New("Tx sign verify error")
	// tx 已经在未确认表中
	ErrRepostTx = errors.New("Tx is already in unconfirm table")
	// ErrBlockchianNotExist used to return the error while name of blockchain not exist
	ErrBlockchianNotExist = fmt.Errorf("Blockchain not exist")
	// 未知错误
	ErrUnknow = errors.New("Unknown error from server")
)

// HandlerError 封装返回值
func HandlerError(err error) map[string]interface{} {
	switch err {
	case nil:
		return map[string]interface{}{
			"code": pb.ReturnCode_RETURNSUCCESS,
			"msg":  "success",
			"data": &map[string]string{},
		}
	case ErrFileNotExist:
		return map[string]interface{}{
			"code": pb.ReturnCode_FILENOTEXIST,
			"msg":  ErrFileNotExist.Error(),
			"data": &map[string]string{},
		}
	case ErrParam:
		return map[string]interface{}{
			"code": pb.ReturnCode_PARAMERR,
			"msg":  ErrParam.Error(),
			"data": &map[string]string{},
		}
	case ErrRequestFailed:
		return map[string]interface{}{
			"code": pb.ReturnCode_HTTPREQUERTFAIL,
			"msg":  ErrRequestFailed.Error(),
			"data": &map[string]string{},
		}
	case ErrReadResponseFailed:
		return map[string]interface{}{
			"code": pb.ReturnCode_HTTPRESPONSEFAIL,
			"msg":  ErrReadResponseFailed.Error(),
			"data": &map[string]string{},
		}
	case ErrResponseFailed:
		return map[string]interface{}{
			"code": pb.ReturnCode_HTTPRESPONSEFAIL,
			"msg":  ErrResponseFailed.Error(),
			"data": &map[string]string{},
		}
	case ErrRequestParam:
		return map[string]interface{}{
			"code": pb.ReturnCode_PARAMERR,
			"msg":  ErrRequestParam.Error(),
			"data": &map[string]string{},
		}
	case ErrAccountNotExist:
		return map[string]interface{}{
			"code": pb.ReturnCode_ACCOUNTNOTEXIST,
			"msg":  ErrAccountNotExist.Error(),
			"data": &map[string]string{},
		}
	case ErrPwNotExist:
		return map[string]interface{}{
			"code": pb.ReturnCode_PWNOTEXIST,
			"msg":  ErrPwNotExist.Error(),
			"data": &map[string]string{},
		}
	case ErrPwExist:
		return map[string]interface{}{
			"code": pb.ReturnCode_PWEXIST,
			"msg":  ErrPwExist.Error(),
			"data": &map[string]string{},
		}
	case ErrDbFail:
		return map[string]interface{}{
			"code": pb.ReturnCode_HTTPRESPONSEFAIL,
			"msg":  ErrDbFail.Error(),
			"data": &map[string]string{},
		}
	case ErrNotLogin:
		return map[string]interface{}{
			"code": pb.ReturnCode_NOTLOGIN,
			"msg":  ErrNotLogin.Error(),
			"data": &map[string]string{},
		}
	case ErrConnectNodeFailure:
		return map[string]interface{}{
			"code": pb.ReturnCode_CONNECTNODEFAIL,
			"msg":  ErrConnectNodeFailure.Error(),
			"data": &map[string]string{},
		}
	case ErrNotEnoughUTXO:
		return map[string]interface{}{
			"code": pb.ReturnCode_UTXONOTENOUGH,
			"msg":  ErrNotEnoughUTXO.Error(),
			"data": &map[string]string{},
		}
	case ErrInvalidAddress:
		return map[string]interface{}{
			"code": pb.ReturnCode_ADDRESSINVALID,
			"msg":  ErrInvalidAddress.Error(),
			"data": &map[string]string{},
		}
	case ErrInvalidFeeAmount:
		return map[string]interface{}{
			"code": pb.ReturnCode_FEEINVALID,
			"msg":  ErrInvalidFeeAmount.Error(),
			"data": &map[string]string{},
		}
	case ErrConnectRefuce:
		return map[string]interface{}{
			"code": pb.ReturnCode_CONNECTREFUSED,
			"msg":  ErrConnectRefuce.Error(),
			"data": &map[string]string{},
		}
	case ErrUtxoEncrypt:
		return map[string]interface{}{
			"code": pb.ReturnCode_UTXOENCRYPTERR,
			"msg":  ErrConnectRefuce.Error(),
			"data": &map[string]string{},
		}
	case ErrFeeNotEnough:
		return map[string]interface{}{
			"code": pb.ReturnCode_FEENOTENOUGN,
			"msg":  ErrFeeNotEnough.Error(),
			"data": &map[string]string{},
		}
	case ErrValidateParams:
		return map[string]interface{}{
			"code": pb.ReturnCode_PARAMSINVALID,
			"msg":  ErrValidateParams.Error(),
			"data": &map[string]string{},
		}
	case ErrTxSign:
		return map[string]interface{}{
			"code": pb.ReturnCode_TXSIGNERR,
			"msg":  ErrTxSign.Error(),
			"data": &map[string]string{},
		}
	case ErrRepostTx:
		return map[string]interface{}{
			"code": pb.ReturnCode_REPOSTTXERR,
			"msg":  ErrRepostTx.Error(),
			"data": &map[string]string{},
		}
	case ErrBlockchianNotExist:
		return map[string]interface{}{
			"code": pb.ReturnCode_BLOCKCHAINNOTEXIST,
			"msg":  ErrBlockchianNotExist.Error(),
			"data": &map[string]string{},
		}
	case ErrUnknow:
		return map[string]interface{}{
			"code": pb.ReturnCode_SERVERERR,
			"msg":  ErrUnknow.Error(),
			"data": &map[string]string{},
		}
	default:
		return map[string]interface{}{
			"code": pb.ReturnCode_INTERNALERR,
			"msg":  err.Error(),
			"data": &map[string]string{},
		}
	}
}
