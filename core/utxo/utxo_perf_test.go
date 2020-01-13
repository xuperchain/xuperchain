package utxo

import (
	"os"
	"testing"
	"time"

	crypto_client "github.com/xuperchain/xuperchain/core/crypto/client"
	ledger_pkg "github.com/xuperchain/xuperchain/core/ledger"
	"github.com/xuperchain/xuperchain/core/pb"
)

func BenchmarkPerformance(t *testing.B) {
	workspace := "/tmp/utxo_perf"
	os.RemoveAll(workspace)
	defer os.RemoveAll(workspace)
	ledger, err := ledger_pkg.NewLedger(workspace, nil, nil,
		DefaultKVEngine, crypto_client.CryptoTypeDefault)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ledger)
	utxoVM, _ := MakeUtxoVM("xuper", ledger, workspace, minerPrivateKey, minerPublicKey, []byte(minerAddress), nil, 5000,
		60, 500, nil, false, DefaultKVEngine, crypto_client.CryptoTypeDefault)
	//创建链的时候分配财富
	tx, err := GenerateRootTx([]byte(`
       {
        "version" : "1"
        , "consensus" : {
                "miner" : "0x00000000000"
        }
        , "predistribution":[
                {
                        "address" : "bob",
                        "quota" : "1000000"
                },
				{
                        "address" : "alice",
                        "quota" : "20000"
                }

        ]
        , "maxblocksize" : "128"
        , "period" : "5000"
        , "award" : "1000"
		} 
    `))
	if err != nil {
		t.Fatal(err)
	}
	block, _ := ledger.FormatRootBlock([]*pb.Transaction{tx})
	t.Logf("blockid %x", block.Blockid)
	confirmStatus := ledger.ConfirmBlock(block, true)
	if !confirmStatus.Succ {
		t.Fatal("confirm block fail")
	}
	playErr := utxoVM.Play(block.Blockid)
	if playErr != nil {
		t.Fatal(playErr)
	}
	txReq := &pb.TxData{}
	txReq.Bcname = "xuper-chain"
	txReq.FromAddr = "bob"
	txReq.FromPubkey = `{"Curvname":"P-256","X":19273820916200149317740158826280160284701834923355092786023370397946429017217,"Y":47876664752361508437602518868161483671473797134385536823847125838767712769836}`
	txReq.FromScrkey = `{"Curvname":"P-256","X":19273820916200149317740158826280160284701834923355092786023370397946429017217,"Y":47876664752361508437602518868161483671473797134385536823847125838767712769836,"D":33958682631561082403597909917154628412303439091483110433683501048213444529537}`
	txReq.Nonce = "nonce"
	txReq.Timestamp = time.Now().UnixNano()
	//bob给alice转1
	txReq.Account = []*pb.TxDataAccount{
		{Address: "alice", Amount: "1"},
	}
	for i := 0; i < t.N; i++ {
		tx, err := utxoVM.GenerateTx(txReq)
		if err != nil {
			t.Fatal(err)
		}
		errDo := utxoVM.DoTx(tx)
		if errDo != nil {
			t.Fatal(errDo)
		}
	}

	/*for i := 0; i < t.N; i++ {
	        t.RunParallel(func(pb *testing.PB) {
	            tx, err := utxoVM.GenerateTx(txReq)
	            if err != nil {
	                t.Fatal(err)
	            }
	            errDo := utxoVM.DoTx(tx)
	            if errDo != nil {
	                t.Fatal(errDo)
	            }
	        })
		}*/
}
