package cmd

import (
	"os"
	"testing"

	"github.com/xuperchain/xuperchain/data/mock"
	_ "github.com/xuperchain/xupercore/lib/storage/kvdb/leveldb"
)

func TestPruneLedger(t *testing.T) {
	workspace := os.TempDir()
	os.RemoveAll(workspace)
	defer os.RemoveAll(workspace)
	econf, err := mock.NewEnvConfForTest()
	if err != nil {
		t.Fatal(err)
	}
	envdir := econf.GenConfFilePath("env.yaml")

	c := &PruneLedgerCommand{
		Name:    "xuper",
		Target:  "0354240c8335e10d8b48d76c0584e29ab604cfdb7b421d973f01a2a49bb67fee",
		Crypto:  "default",
		EnvConf: envdir,
	}
	err = c.pruneLedger(econf)
	if err != nil {
		t.Log("prune ledger fail.err:", err)
	} else {
		t.Log("prune ledger succ.blockid:", c.Target)
	}
}
