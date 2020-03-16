package utxo

import "testing"
import "time"
import "fmt"
import "github.com/xuperchain/xuperchain/core/pb"

func TestSpinLock(t *testing.T) {
	sp := NewSpinLock()
	tx1 := &pb.Transaction{
		Txid: []byte("tx1"),
		TxInputs: []*pb.TxInput{
			&pb.TxInput{
				RefTxid: []byte("tx0"),
			},
			&pb.TxInput{
				RefTxid:   []byte("tx3"),
				RefOffset: 1,
			},
		},
		TxOutputs: []*pb.TxOutput{
			&pb.TxOutput{},
		},
		TxInputsExt: []*pb.TxInputExt{
			&pb.TxInputExt{
				Bucket: "bk2",
				Key:    []byte("key2"),
			},
		},
		TxOutputsExt: []*pb.TxOutputExt{
			&pb.TxOutputExt{
				Bucket: "bk1",
				Key:    []byte("key1"),
			},
		},
	}
	tx2 := &pb.Transaction{
		TxInputsExt: []*pb.TxInputExt{
			&pb.TxInputExt{
				Bucket: "bk2",
				Key:    []byte("key2"),
			},
		},
		TxInputs: []*pb.TxInput{
			&pb.TxInput{
				RefTxid: []byte("tx3"),
			},
		},
	}
	lockKeys1 := sp.ExtractLockKeys(tx1)
	lockKeys2 := sp.ExtractLockKeys(tx2)
	t.Log(lockKeys1)
	t.Log(lockKeys2)
	if fmt.Sprintf("%v", lockKeys1) != "[bk1/key1:X bk2/key2:S tx0_0:X tx1_0:X tx3_1:X]" {
		t.Fatal("tx1 lock error")
	}
	if fmt.Sprintf("%v", lockKeys2) != "[bk2/key2:S tx3_0:X]" {
		t.Fatal("tx2 lock error")
	}
	go func() {
		succLks, ok := sp.TryLock(lockKeys2)
		t.Log("tx2 got lock", succLks, ok)
		sp.Unlock(succLks)
	}()
	sp.TryLock(lockKeys1)
	if !sp.IsLocked("tx1_0") {
		t.Fatal("tx1_0 is expected to be locked")
	}
	time.Sleep(1 * time.Second)
	sp.Unlock(lockKeys1)
	t.Log("tx1 unlock")
}
