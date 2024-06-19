package chain

import (
	"eastnode/types"
	"eastnode/utils"
	store "eastnode/utils/store"
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/cbergoon/merkletree"
	"github.com/near/borsh-go"
	bolt "go.etcd.io/bbolt"
)

type GenesisAccount struct {
	Account string
	Value   string
}

type Chain struct {
	Locked  bool
	Store   *store.Store
	Mempool *Mempool
}

func (c *Chain) Init() *Chain {
	c.Store = store.GetInstance()

	c.Mempool = new(Mempool)
	c.Mempool.Init(c.Store.KV)

	log.Println("[+] chain initialized")

	c.ProduceBlock()

	return c
}

func (c *Chain) Lock() {
	c.Locked = true
}

func (c *Chain) Unlock() {
	c.Locked = false

	c.ProduceBlock()
}

func (c *Chain) Genesis() bool {
	genesisInit := true

	c.Store.KV.View(func(tx *bolt.Tx) error {
		bBlocks := tx.Bucket([]byte("blocks"))
		v := bBlocks.Get(utils.Itob(0))

		if v != nil {
			genesisInit = false
		}

		return nil
	})

	return genesisInit
}

func (c *Chain) GetBlockHeight() uint64 {
	blockHeight := new(uint64)

	c.Store.KV.View(func(tx *bolt.Tx) error {
		bBlocks := tx.Bucket([]byte("blocks"))

		*blockHeight = bBlocks.Sequence()

		return nil
	})

	return *blockHeight
}

func (c *Chain) GetBlockHash(blockHeight uint64) string {
	block := c.GetBlock(blockHeight)

	blockHeaderCompact, err := borsh.Serialize(block.Header)
	if err != nil {
		panic(err)
	}

	blockHeaderHash := utils.SHA256(blockHeaderCompact)

	return blockHeaderHash
}

func (c *Chain) GetBlock(blockHeight uint64) types.Block {
	block := new(types.Block)

	c.Store.KV.View(func(tx *bolt.Tx) error {
		bBlocks := tx.Bucket([]byte("blocks"))
		prevBlockRaw := bBlocks.Get([]byte(utils.Itob(blockHeight)))

		borsh.Deserialize(block, prevBlockRaw)

		return nil
	})

	return *block
}

func (c *Chain) ProduceBlock() error {
	// process all tx
	transactions := []types.SignedTransaction{}

	// if genesis
	if c.Genesis() {
		// processing genesis block
		log.Println("Processing Genesis Block")

		// get genesis config
		genAccounts := []GenesisAccount{
			{
				Account: "bc1ph02hv4dc9afhcycs04vtawkmmm055j3g39w7mqur6d2x5ng4dgmshavfvj",
				Value:   "1000",
			},
			{
				Account: "bc1pkskdm7qk0z4gr8cgy38ysa00gyftj364gmf3uruse80c6gzunf6s0ywcsh",
				Value:   "5000",
			},
			{
				Account: "bc1pkskdm7qk0z4gr8cgy38ysa00gyftj364gmf3uruse80c6gzunf6s0ywcsh",
				Value:   "10000",
			},
		}

		genesisTime := time.Now().UnixMilli()

		txMerkleTree := []merkletree.Content{}

		for _, v := range genAccounts {
			actions := []types.Action{}
			actions = append(actions, types.Action{
				Kind: "genesis",
				Args: []string{v.Account, v.Value},
			})

			actionsPacked, err := borsh.Serialize(actions)
			if err != nil {
				panic(err)
			}

			gTx := types.Transaction{
				Signer:   "genesis",
				Receiver: v.Account,
				Actions:  hex.EncodeToString(actionsPacked),
			}
			gTxPacked, err := borsh.Serialize(gTx)
			if err != nil {
				panic(err)
			}

			gSignedTx := types.SignedTransaction{
				ID:          "GENESIS_" + utils.SHA256(gTxPacked),
				Signature:   "GENESIS",
				Transaction: hex.EncodeToString(gTxPacked),
			}

			_, err = c.Store.Instance.Exec(fmt.Sprintf(`
				INSERT INTO transactions (id, block_id, signer, receiver, actions, created_at)
				VALUES ('%s', '%d', '%s', '%s', '%x', '%d');`,
				gSignedTx.ID, 0, gTx.Signer, gTx.Receiver, actionsPacked, genesisTime,
			))
			if err != nil {
				panic(err)
			}
			transactions = append(transactions, gSignedTx)
			txMerkleTree = append(txMerkleTree, types.MerkleTreeContent{
				Value: gSignedTx.ID,
			})
		}
		var workingEngineHash string
		workingEngineHashRaw := c.Store.Instance.QueryRow("SELECT dolt_hashof_db('WORKING')")
		workingEngineHashRaw.Scan(&workingEngineHash)

		newBlock := new(types.Block)

		blockData, err := borsh.Serialize(transactions)
		if err != nil {
			panic(err)
		}

		t, err := merkletree.NewTree(txMerkleTree)
		if err != nil {
			panic(err)
		}

		newBlock.Header = types.BlockHeader{
			ChainID:     "Eastblue",
			Height:      0,
			LastBlockID: []byte(""),
			DataHash:    t.MerkleRoot(),
			Time:        genesisTime,
			StorageHash: []byte(workingEngineHash),
		}

		newBlock.Data = blockData

		blockHeaderCompact, err := borsh.Serialize(newBlock.Header)
		if err != nil {
			panic(err)
		}

		blockHeaderHash := utils.SHA256(blockHeaderCompact)
		log.Println(workingEngineHash)
		log.Println("block hash: " + blockHeaderHash)

		// consensus done
		// commit block
		_, err = c.Store.Instance.Exec("CALL DOLT_COMMIT('-Am', 'commit genesis block');")

		if err != nil {
			panic(err)
		}

		c.Store.KV.Update(func(tx *bolt.Tx) error {
			bBlocks := tx.Bucket([]byte("blocks"))
			bCommon := tx.Bucket([]byte("common"))

			blockBuf, err := borsh.Serialize(newBlock)
			if err != nil {
				panic(err)
			}
			bBlocks.Put([]byte(utils.Itob(newBlock.Header.Height)), blockBuf)

			bCommon.Put([]byte(blockHeaderHash), []byte(utils.Itob(newBlock.Header.Height)))

			return nil
		})
	} else {
		// read from mempool & product block
		blockHeight := c.GetBlockHeight()

		pendingTx := c.Mempool.Length()

		if pendingTx == 0 || c.Locked {
			return nil
		}

		pendingTx = uint64(math.Min(float64(pendingTx), float64(10)))

		log.Println("Processing new block")

		c.Lock()

		lastBlock := c.GetBlock(blockHeight)

		txMerkleTree := []merkletree.Content{}

		blockTime := time.Now().UnixMilli()

		// take the first 10 transactions
		for i := uint64(0); i < pendingTx; i++ {
			pSignedTx := c.Mempool.Get(i)

			txInHex, err := hex.DecodeString(pSignedTx.Transaction)
			if err != nil {
				panic(err)
			}

			txUnpacked := new(types.Transaction)
			borsh.Deserialize(*txUnpacked, txInHex)

			q := fmt.Sprintf(`
				INSERT INTO transactions (id, block_id, signer, receiver, actions, created_at)
				VALUES ('%s', '%d', '%s', '%s', '%s', '%d');`,
				pSignedTx.ID, blockHeight+1, txUnpacked.Signer, txUnpacked.Receiver, txUnpacked.Actions, blockTime,
			)

			_, err = c.Store.Instance.Exec(q)
			if err != nil {
				panic(err)
			}

			transactions = append(transactions, pSignedTx)
			txMerkleTree = append(txMerkleTree, types.MerkleTreeContent{
				Value: pSignedTx.ID,
			})
		}

		var workingEngineHash string
		workingEngineHashRaw := c.Store.Instance.QueryRow("SELECT dolt_hashof_db('WORKING')")
		workingEngineHashRaw.Scan(&workingEngineHash)

		log.Println("new storage hash: " + workingEngineHash)

		newBlock := new(types.Block)

		blockData, err := borsh.Serialize(transactions)

		if err != nil {
			panic(err)
		}

		prevBlockHeaderCompact, err := borsh.Serialize(lastBlock.Header)
		if err != nil {
			panic(err)
		}

		prevBlockHeaderHash := utils.SHA256(prevBlockHeaderCompact)

		t, err := merkletree.NewTree(txMerkleTree)
		if err != nil {
			panic(err)
		}

		newBlock.Header = types.BlockHeader{
			ChainID:     "Eastblue",
			Height:      blockHeight + 1,
			LastBlockID: []byte(prevBlockHeaderHash),
			DataHash:    t.MerkleRoot(),
			Time:        blockTime,
			StorageHash: []byte(workingEngineHash),
		}
		newBlock.Data = blockData

		blockHeaderCompact, err := borsh.Serialize(newBlock.Header)
		if err != nil {
			panic(err)
		}

		blockHeaderHash := utils.SHA256(blockHeaderCompact)
		log.Println("new block hash: " + blockHeaderHash)

		// consensus done
		// commit block
		_, err = c.Store.Instance.Exec(fmt.Sprintf(`
			CALL DOLT_COMMIT('-Am', 'commit new block %d');
		`, newBlock.Header.Height))

		// remove pending transaction
		for i := uint64(0); i < (pendingTx); i++ {
			c.Mempool.Dequeue()
		}

		if err != nil {
			panic(err)
		}

		var cEngineHash string
		cEngineHashRaw := c.Store.Instance.QueryRow("SELECT dolt_hashof_db()")
		cEngineHashRaw.Scan(&cEngineHash)

		log.Println("check storage hash: " + cEngineHash)

		c.Store.KV.Update(func(tx *bolt.Tx) error {
			bBlocks := tx.Bucket([]byte("blocks"))
			bCommon := tx.Bucket([]byte("common"))

			blockBuf, err := borsh.Serialize(newBlock)
			if err != nil {
				panic(err)
			}
			// update block height
			bBlocks.NextSequence()
			bBlocks.Put([]byte(utils.Itob(newBlock.Header.Height)), blockBuf)

			bCommon.Put([]byte(blockHeaderHash), blockBuf)

			return nil
		})

		c.Unlock()
	}

	return nil
}
