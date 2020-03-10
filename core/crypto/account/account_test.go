package account

import (
	"encoding/base64"
	"fmt"
	"testing"
)

var oldMnem = "雕 轰 毒 氏 空 让 翻 肃 辞 料 克 收 稿 描 隐 童 辛 泥 萨 鸭 背 霍 持 贝"
var newMnem = "呈 仓 冯 滚 刚 伙 此 丈 锅 语 揭 弃 精 塘 界 戴 玩 爬 奶 滩 哀 极 样 费"

func Test_RetrieveAccountByMnem(t *testing.T) {
	// new version of mnemonic
	newAccount, err := GenerateAccountByMnemonic(newMnem, 1)
	if err != nil {
		t.Errorf("Generate account failed with new Mnem failed, err=%v\n", err)
		return
	}
	t.Logf("Generate account OK with new Mnem success, address=%s, private=%s\n", newAccount.Address, newAccount.JSONPrivateKey)

	// old version of mnemonic
	oldAccount, err := GenerateAccountByMnemonic(oldMnem, 1)
	if err != nil {
		t.Errorf("Generate account failed with old Mnem failed, err=%v\n", err)
		return
	}
	t.Logf("Generate account OK with old Mnem success, address=%s, private=%s\n", oldAccount.Address, oldAccount.JSONPrivateKey)
}

func Test_Encrpty(t *testing.T) {

	msg :="gjjtestxzhtestestestguangdongzhongshanhuojukaifaqu"
	key := randKey(msg)


	fmt.Println("密钥:"+string(key))


	bytes, _ := aesEncrypt([]byte(msg), key)


	fmt.Println("对称加密私钥后的密文"+base64.StdEncoding.EncodeToString(bytes))

	aes, _ := aesDecrypt(bytes, key)

	fmt.Println("对称解密后的私钥"+string(aes))


}

