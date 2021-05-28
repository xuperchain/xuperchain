package cmd

import (
	"github.com/xuperchain/xupercore/kernel/mock"
	"io/ioutil"
	"os"
	"testing"
)

func TestPruneLedger(t *testing.T) {
	workspace, dirErr := ioutil.TempDir("/tmp", "")
	if dirErr != nil {
		t.Fatal(dirErr)
	}
	os.RemoveAll(workspace)
	defer os.RemoveAll(workspace)

	econf, err := mock.NewEnvConfForTest()
	if err != nil {
		t.Fatal(err)
	}
	genesisConf := econf.GenDataAbsPath("genesis/xuper.json")
	envdir := econf.GenConfFilePath("env.yaml")

	c := &PruneLedgerCommand{
		Name:        "xuper",
		Target:      "11111",
		Crypto:      "default",
		GenesisConf: genesisConf,
		EnvConf:     envdir,
	}
	err = c.pruneLedger()
	if err != nil {
		t.Fatal("prune ledger fail.err:", err)
	}
	t.Log("prune ledger succ.blockid:", c.Target)
}
