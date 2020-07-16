package txhash

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/xuperchain/xuperchain/core/pb"
)

func readTxFile(tb testing.TB, name string) *pb.Transaction {
	buf, err := ioutil.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		tb.Fatal(err)
	}

	tx := new(pb.Transaction)
	err = proto.Unmarshal(buf, tx)
	if err != nil {
		tb.Fatal(err)
	}

	return tx
}
func BenchmarkTxHashV2(b *testing.B) {
	tx := readTxFile(b, "tx.pb")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TxSignature(tx, true)
	}
}

func BenchmarkTxHashV1(b *testing.B) {
	tx := readTxFile(b, "tx.pb")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MakeTransactionID(tx)
	}
}

func TestTxHashV2(t *testing.T) {
	tx := readTxFile(t, "tx.pb")
	txid := TxSignature(tx, true)
	t.Logf("txid = %x", txid)
}
