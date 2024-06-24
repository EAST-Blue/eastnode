package chain

import (
	"eastnode/types"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/near/borsh-go"
	bolt "go.etcd.io/bbolt"
)

func initMempoolTest() *Mempool {
	clearMempoolTest()
	if err := os.Mkdir("db_test", os.ModeDir|0755); err != nil {
		log.Panicln(err)
	}

	kv, err := bolt.Open("db_test/chain.db", 0600, nil)
	if err != nil {
		panic(err)
	}

	mempool := &Mempool{}

	if err := mempool.Init(kv); err != nil {
		log.Panicln(err)
	}

	return mempool
}

func clearMempoolTest() {
	if err := os.RemoveAll("db_test"); err != nil {
		log.Panicln(err)
	}
}

func TestMempool(t *testing.T) {
	mempool := initMempoolTest()
	defer t.Cleanup(clearMempoolTest)

	length := 15

	for i := 0; i < length; i++ {
		tx := types.Transaction{}

		serializedActions, err := borsh.Serialize(tx)
		if err != nil {
			t.Error(err)
		}
		serializedTx := hex.EncodeToString(serializedActions)

		signedTx := types.SignedTransaction{ID: fmt.Sprint(i), Signature: "signature", Transaction: serializedTx}
		mempool.Enqueue(signedTx)

	}

	if mempool.Length() != uint64(length) {
		t.Error("Length incorrect")
	}

	recoveredSignedTx := mempool.Get(0)
	if recoveredSignedTx.ID != "0" {
		t.Error("First item incorrect")
	}
	recoveredSignedTx = mempool.Get(1)
	if recoveredSignedTx.ID != "1" {
		t.Error("Second item incorrect")
	}

	recoveredSignedTx = mempool.Dequeue()
	if recoveredSignedTx.ID != "0" {
		t.Error("First item dequeue incorrect")
	}

	recoveredSignedTx = mempool.Get(0)
	if recoveredSignedTx.ID != "1" {
		t.Error("First item is incorrect after dequeue")
	}
}
