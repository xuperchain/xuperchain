package account

import ()

// ECDSAInfo 助记词、随机字节数组、钱包地址
type ECDSAInfo struct {
	EntropyByte []byte
	Mnemonic    string
	Address     string
}

// ECDSAAccountToCloud 钱包地址、被加密后的私钥、被加密后的助记词、支付密码的明文
type ECDSAAccountToCloud struct {
	Address                 string
	JSONEncryptedPrivateKey string
	EncryptedMnemonic       string
	Password                string
}
