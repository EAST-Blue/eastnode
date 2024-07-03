package chain

import (
	"eastnode/indexer/repository"
	"eastnode/runtime"
	"eastnode/types"
	"eastnode/utils"
	store "eastnode/utils/store"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/btcsuite/btcd/btcutil/bech32"
	"github.com/cbergoon/merkletree"
	"github.com/near/borsh-go"
	bolt "go.etcd.io/bbolt"
)

type GenesisAccount struct {
	Account string
	Value   string
}

type Chain struct {
	// TODO: use sync.Mutex https://go.dev/tour/concurrency/9
	Locked      bool
	Store       *store.Store
	Mempool     *Mempool
	WasmRuntime *runtime.WasmRuntime
}

func (c *Chain) Init() *Chain {
	c.Store = store.GetInstance(store.ChainDB)

	c.Mempool = new(Mempool)
	c.Mempool.Init(c.Store.KV)

	indexerInstance := store.GetInstance(store.IndexerDB)
	indexerRepo := repository.NewIndexerRepository(indexerInstance.Gorm)
	c.WasmRuntime = &runtime.WasmRuntime{Store: *store.GetInstance(store.SmartIndexDB), IndexerRepo: indexerRepo}

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

func (c *Chain) GetNonce(pubKey string) uint64 {
	nonce := uint64(0)

	err := c.Store.KV.View(func(tx *bolt.Tx) error {
		bNonce := tx.Bucket([]byte("nonce"))
		lastNonce := bNonce.Get([]byte(pubKey))

		if lastNonce != nil {
			nonce = utils.Btoi(lastNonce)
		}

		return nil
	})

	if err != nil {
		panic(err)
	}

	return nonce
}

func (c *Chain) CheckTx(signedTx types.SignedTransaction) error {
	// check signature valid
	verified := signedTx.IsValid()

	// unpack signedTx
	inputTx := signedTx.Unpack()

	err := c.Store.KV.View(func(tx *bolt.Tx) error {
		bNonce := tx.Bucket([]byte("nonce"))
		lastNonce := bNonce.Get([]byte(inputTx.Signer))

		// if lastNonce not exist, this is the first transaction
		if lastNonce == nil {
			return nil
		}

		// if lastNonce exist, check if current nonce larger than last nonce
		if inputTx.Nonce > utils.Btoi(lastNonce) {
			return nil
		}

		return errors.New("invalid nonce")
	})

	if verified && err == nil {
		return nil
	}

	if !verified {
		return errors.New("invalid signature")
	}

	return err
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
			pSignedTx := c.Mempool.Get(i) // ensure pushback

			txUnpacked := pSignedTx.Unpack()

			parsedActions := new([]types.Action)
			utils.DecodeHexAndBorshDeserialize(parsedActions, txUnpacked.Actions)

			// Process actions
			for i, action := range *parsedActions {
				if action.Kind == "deploy" || action.Kind == "redeploy" {
					c.ProcessDeploy(txUnpacked, action)
					// WORKAROUND: file is too large for column 'actions'
					(*parsedActions)[i].Args = []string{}
					txUnpacked.Actions = utils.BorshSerializeAndEncodeHex(parsedActions)
				} else if action.Kind == "call" {
					c.ProcessCall(txUnpacked, action)
				} else if action.Kind == "view" {
					// TODO: handle view function in the front, before going into produce block, so there'll be no view function here
				}
			}

			q := fmt.Sprintf(`
				INSERT INTO transactions (id, block_id, signer, receiver, actions, created_at)
				VALUES ('%s', '%d', '%s', '%s', '%s', '%d');`,
				pSignedTx.ID, blockHeight+1, txUnpacked.Signer, txUnpacked.Receiver, txUnpacked.Actions, blockTime,
			)

			// TODO: create another table for logs (tx_id, logs)

			_, err := c.Store.Instance.Exec(q)
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

		// REFACTOR: block commit to a new function. duplicate with genesis block up there.
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

func (c *Chain) ProcessWasmCall(signer string, smartIndexAddress string, functionName string, args []string, kind types.ActionKind) any {

	var resultWasmBlob []byte
	sr := c.Store.Instance.QueryRow(fmt.Sprintf("SELECT wasm_blob FROM smart_index WHERE smart_index_address = '%s';", smartIndexAddress))
	sr.Scan(&resultWasmBlob)

	return c.WasmRuntime.RunWasmFunction(runtime.Address(signer), resultWasmBlob, smartIndexAddress, functionName, args, kind)
}

func (c *Chain) ProcessCall(tx types.Transaction, action types.Action) {
	c.ProcessWasmCall(tx.Signer, tx.Receiver, action.FunctionName, action.Args, types.Call)
}

func (c *Chain) ProcessDeploy(tx types.Transaction, action types.Action) string {
	actionSerialized, err := borsh.Serialize(action)
	if err != nil {
		panic(err)
	}

	// TODO: Validate wasm file
	wasmBytes := action.Args[0]

	var q string
	var smartIndexAddress string

	if len(action.Args) == 1 { // new smart index
		// generate contract account based on the initial tx
		publicKey, err := hex.DecodeString(tx.Signer)
		hash, err := hex.DecodeString(utils.SHA256(append(actionSerialized, publicKey...)))
		if err != nil {
			panic(err)
		}

		smartIndexAddress, err = bech32.EncodeFromBase256("idx", hash)
		if err != nil {
			panic(err)
		}

		// maximum length is 64, trimmed this to 32 chars
		smartIndexAddress = smartIndexAddress[:32]

		// TODO: change this to query, exec to prevent sqli
		q = fmt.Sprintf(`
				INSERT INTO smart_index (smart_index_address, owner_address, wasm_blob)
				VALUES ('%s', '%s', X'%s');`, smartIndexAddress, tx.Signer, wasmBytes,
		)
	} else { // redeploy
		smartIndexAddress = action.Args[1]
		q = fmt.Sprintf(`
				UPDATE smart_index SET wasm_blob = x'%s' WHERE smart_index_address = '%s';`, wasmBytes, smartIndexAddress,
		)
	}

	_, err = c.Store.Instance.Exec(q)
	if err != nil {
		// TODO: handle error here, do not panic
		fmt.Println(err)
	}

	return smartIndexAddress
}

func (c *Chain) GetSmartIndexWasm(smartIndexAddress string) []byte {
	var wasmBlob []byte
	wasmBlobRaw := c.Store.Instance.QueryRow("SELECT wasm_blob FROM smart_index WHERE smart_index_address = ?", smartIndexAddress)
	wasmBlobRaw.Scan(&wasmBlob)

	return wasmBlob
}

func (c *Chain) GetTransaction(txId string) map[string]interface{} {

	var blockId string
	var signer string
	var receiver string
	var actions string
	var createdAt int64

	selectRaw := c.Store.Instance.QueryRow("SELECT block_id, signer, receiver, actions, created_at FROM transactions WHERE id = ?", txId)
	selectRaw.Scan(&blockId, &signer, &receiver, &actions, &createdAt)

	res := map[string]interface{}{
		"block_id":   blockId,
		"signer":     signer,
		"receiver":   receiver,
		"actions":    actions,
		"created_at": createdAt,
	}

	return res
}
