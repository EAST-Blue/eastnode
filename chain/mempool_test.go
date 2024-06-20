package chain

import (
	"eastnode/types"
	"fmt"
	"log"
	"os"
	"testing"

	bolt "go.etcd.io/bbolt"
)

func initTest() *Mempool {
	clearTest()
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

func clearTest() {
	if err := os.RemoveAll("db_test"); err != nil {
		log.Panicln(err)
	}
}

func TestMempool(t *testing.T) {
	mempool := initTest()
	defer t.Cleanup(clearTest)

	length := 15

	for i := 0; i < length; i++ {
		signedTx := types.SignedTransaction{ID: fmt.Sprint(i), Signature: "signature", Transaction: "transaction"}
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
