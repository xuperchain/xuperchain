package utxo

import (
//"os"
//"testing"

//"io/ioutil"
//"time"

//crypto_client "github.com/xuperchain/xuperunion/crypto/client"
//ledger_pkg "github.com/xuperchain/xuperunion/ledger"
)

/*
func TestAsyncBasic(t *testing.T) {
	workSpace, dirErr := ioutil.TempDir("/tmp", "")
	if dirErr != nil {
		t.Fatal(dirErr)
	}
	os.RemoveAll(workSpace)
	defer os.RemoveAll(workSpace)
	// 初始化一个账本
	ledger, err := ledger_pkg.NewLedger(workSpace, nil, nil, DefaultKVEngine, crypto_client.CryptoTypeDefault)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ledger)
	utxoVM, _ := NewUtxoVM("xuper", ledger, workSpace, minerPrivateKey, minerPublicKey, []byte(minerAddress),
		nil, false, DefaultKVEngine, crypto_client.CryptoTypeDefault)

	// test for IsAsync()
	isAsync := utxoVM.IsAsync()
	t.Log("is async ", isAsync)
	utxoVM.asyncMode = true
	isAsync = utxoVM.IsAsync()
	t.Log("is async ", isAsync)

	// test for SetBlockGenEvent()
	t.Log("asyncTryBlockGen ", utxoVM.asyncTryBlockGen)
	utxoVM.SetBlockGenEvent()
	t.Log("asyncTryBlockGen ", utxoVM.asyncTryBlockGen)

	// test for StartAsyncWriter()
	utxoVM.StartAsyncWriter()
	<-time.After(2 * time.Second)
	tx, _ := GenerateRootTx([]byte(`
        {
            "version" : "1"
            , "consensus" : {
                    "miner" : "0x00000000000"
            }
            , "predistribution":[
                    {
                            "address" : "` + BobAddress + `",
                            "quota" : "100"
                    },
                    {
                            "address" : "` + AliceAddress + `",
                            "quota" : "200"
                    }
            ]
            , "maxblocksize" : "128"
            , "period" : "5000"
            , "award" : "1000"
        }
    `))
	// utxoVM.GenerateTx(txReq)
	inboundTx := &InboundTx{tx: tx}
	for i := 0; i < 3001; i++ {
		utxoVM.inboundTxChan <- inboundTx
	}
	<-time.After(2 * time.Second)
	go utxoVM.asyncCancel()
	//utxoVM.inboundTxChan <- inboundTx
	// test for NotifyFinishBlockGen()
	t.Log("asyncMode ", utxoVM.asyncMode)
	utxoVM.NotifyFinishBlockGen()
	<-time.After(3 * time.Second)

}
*/
