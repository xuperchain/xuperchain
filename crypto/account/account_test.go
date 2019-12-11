package account

import (
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
