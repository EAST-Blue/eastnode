package chain

import (
	"eastnode/types"
	"eastnode/utils"
	"encoding/json"
	"log"

	bolt "go.etcd.io/bbolt"
)

type Mempool struct {
	db *bolt.DB
}

func (q *Mempool) Init(db *bolt.DB) error {
	q.db = db
	err := q.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("mempool"))

		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists([]byte("mempool-meta"))

		if err != nil {
			return err
		}

		bMempoolMeta := tx.Bucket([]byte("mempool-meta"))

		head := bMempoolMeta.Get([]byte("head"))

		if head == nil {
			bMempoolMeta.Put([]byte("head"), []byte(utils.Itob(0)))
		}

		return nil
	})

	if err != nil {
		return err
	}
	return nil
}

func (q *Mempool) Enqueue(inputTx types.SignedTransaction) {
	err := q.db.Update(func(tx *bolt.Tx) error {
		bMempool := tx.Bucket([]byte("mempool"))

		// ENSURE: make sure always start with 0 to ...
		index, err := bMempool.NextSequence()

		if err != nil {
			log.Println(err)
			return err
		}

		inputTxBuf, err := json.Marshal(inputTx)
		head := index - 1

		if err != nil {
			return err
		}

		bMempool.Put([]byte(utils.Itob(head)), inputTxBuf)

		return nil
	})

	if err != nil {
		panic(err)
	}
}

func (q *Mempool) Dequeue() types.SignedTransaction {
	btx := new(types.SignedTransaction)

	err := q.db.Update(func(tx *bolt.Tx) error {
		bMempool := tx.Bucket([]byte("mempool"))
		bMempoolMeta := tx.Bucket([]byte("mempool-meta"))

		head := bMempoolMeta.Get([]byte("head"))

		json.Unmarshal(bMempool.Get(head), btx)

		newHead := utils.Btoi(head) + 1

		bMempoolMeta.Put([]byte("head"), utils.Itob(newHead))
		bMempool.Delete(head)

		return nil
	})

	if err != nil {
		panic(err)
	}

	return *btx
}

func (q *Mempool) Get(index uint64) types.SignedTransaction {
	btx := new(types.SignedTransaction)

	err := q.db.View(func(tx *bolt.Tx) error {
		bMempool := tx.Bucket([]byte("mempool"))
		bMempoolMeta := tx.Bucket([]byte("mempool-meta"))

		head := bMempoolMeta.Get([]byte("head"))

		key := utils.Btoi(head) + index

		json.Unmarshal(bMempool.Get(utils.Itob(key)), btx)

		return nil
	})

	if err != nil {
		panic(err)
	}

	return *btx
}

func (q *Mempool) Head() {}

func (q *Mempool) Tail() {}

func (q *Mempool) Length() uint64 {
	length := new(uint64)

	err := q.db.View(func(tx *bolt.Tx) error {
		bMempool := tx.Bucket([]byte("mempool"))
		tail := bMempool.Sequence()

		bMempoolMeta := tx.Bucket([]byte("mempool-meta"))
		head := bMempoolMeta.Get([]byte("head"))

		*length = tail - utils.Btoi(head)

		return nil
	})

	if err != nil {
		panic(err)
	}

	return *length
}
