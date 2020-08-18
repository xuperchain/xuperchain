package txhash

import (
	"encoding/hex"
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
		txDigestHashV2(tx, true)
	}
}

func BenchmarkTxHashV1(b *testing.B) {
	tx := readTxFile(b, "tx.pb")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MakeTransactionID(tx)
	}
}

func TestTxHashVersion(t *testing.T) {
	tx := readTxFile(t, "tx.pb")
	txids := map[int]string{
		1: "1dadd16344d1dd18f10ae63006feeac0c961e63336984d8bac7f16271bb0b2af",
		2: "852da114b2c7f8f2cd71df01f89f12e6a04a9282aeab1990d3f70b0cedea6a02",
		3: "30190490f724ed4b9ef68afc7d691c5c62a7c1e8acce04bf655d2adc75167b05",
	}
	for version, expect := range txids {
		tx.Version = int32(version)
		txid, err := MakeTransactionID(tx)
		if err != nil {
			t.Fatal(err)
		}
		txidstr := hex.EncodeToString(txid)
		if txidstr != expect {
			t.Fatalf("expect %s got %s when version = %d", expect, txidstr, version)
		}
	}

}
