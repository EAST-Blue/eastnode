package tests

import (
	"eastnode/chain"
	"eastnode/runtime"
	"eastnode/types"
	"eastnode/utils"
	store "eastnode/utils/store"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil/bech32"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/near/borsh-go"
)

var SmartIndexAddress string

func initKey() (*secp256k1.PrivateKey, *secp256k1.PublicKey) {
	keyHex := "7c67c815e1c4a25fe70d95aad9440b682bdcbe6e2baf34d460966e605705ea8e"

	privateKeyBytes, err := hex.DecodeString(keyHex)
	if err != nil {
		panic(err)
	}

	privateKey, publicKey := btcec.PrivKeyFromBytes(privateKeyBytes)

	return privateKey, publicKey
}

func initOrdinalsTest() *chain.Chain {
	clearOrdinalsTest()
	if err := os.Mkdir("db_test", os.ModeDir|0755); err != nil {
		log.Panicln(err)
	}

	bc := new(chain.Chain)
	bc.Store = store.GetFakeInstanceCustom(store.ChainDB, "../utils/store/test/ordinals_doltdump.sql")
	bc.WasmRuntime = &runtime.WasmRuntime{Store: *store.GetFakeInstanceCustom(store.SmartIndexDB, "../utils/store/test/ordinals_doltdump.sql")}
	bc.Mempool = new(chain.Mempool)

	if err := bc.Mempool.Init(bc.Store.KV); err != nil {
		log.Panicln(err)
	}

	bc.ProduceBlock()

	deployOrdinals(bc)

	return bc
}

func clearOrdinalsTest() {
	// if err := os.RemoveAll("db_test"); err != nil {
	// 	log.Panicln(err)
	// }
}

func deployOrdinals(bc *chain.Chain) {
	_, publicKey := initKey()

	wasmBytes, _ := os.ReadFile("./ordinals.wasm")
	actions := []types.Action{{
		Kind:         "deploy",
		FunctionName: "",
		Args:         []string{hex.EncodeToString(wasmBytes)},
	}}

	serializedActions, _ := borsh.Serialize(actions)

	serializedActionsHex := hex.EncodeToString(serializedActions)

	transaction := types.Transaction{
		Signer:  publicKey.X().String(),
		Actions: serializedActionsHex,
	}

	serializedTx, err := borsh.Serialize(transaction)
	if err != nil {
		fmt.Println(err)
	}
	serializedTxHex := hex.EncodeToString(serializedTx)

	signedTx := types.SignedTransaction{ID: utils.SHA256(serializedTx), Signature: "signature", Transaction: serializedTxHex}
	bc.Mempool.Enqueue(signedTx)
	bc.ProduceBlock()

	actionSerialized, _ := borsh.Serialize(actions[0])
	publicKeyNew, _ := hex.DecodeString(transaction.Signer)
	hash, err := hex.DecodeString(utils.SHA256(append(actionSerialized, publicKeyNew...)))
	if err != nil {
		log.Panic(err)
	}

	smartIndexAddress, _ := bech32.EncodeFromBase256("idx", hash)

	// maximum length is 64, trimmed this to 32 chars
	smartIndexAddress = smartIndexAddress[:32]
	SmartIndexAddress = smartIndexAddress
	fmt.Println(smartIndexAddress)

	actionsInit := []types.Action{{
		Kind:         "call",
		FunctionName: "init",
		Args:         []string{},
	}}

	serializedActionsInit, _ := borsh.Serialize(actionsInit)
	serializedActionsInitHex := hex.EncodeToString(serializedActionsInit)

	transactionInit := types.Transaction{
		Signer:   publicKey.X().String(),
		Receiver: SmartIndexAddress,
		Actions:  serializedActionsInitHex,
	}

	serializedTx, err = borsh.Serialize(transactionInit)
	if err != nil {
		fmt.Println(err)
	}
	serializedTxHex = hex.EncodeToString(serializedTx)

	signedTx = types.SignedTransaction{ID: utils.SHA256(serializedTx), Signature: "signature", Transaction: serializedTxHex}
	bc.Mempool.Enqueue(signedTx)
	bc.ProduceBlock()
}

func TestOrdinalsIndexMinting(t *testing.T) {
	bc := initOrdinalsTest()
	defer t.Cleanup(clearOrdinalsTest)

	_, publicKey := initKey()

	actions := []types.Action{{
		Kind:         "call",
		FunctionName: "index",
		Args:         []string{"138", "138"},
	}}

	serializedActions, err := borsh.Serialize(actions)
	if err != nil {
		t.Error(err)
	}
	serializedActionsHex := hex.EncodeToString(serializedActions)

	transaction := types.Transaction{
		Signer:   publicKey.X().String(),
		Receiver: SmartIndexAddress,
		Actions:  serializedActionsHex,
	}

	serializedTx, err := borsh.Serialize(transaction)
	if err != nil {
		fmt.Println(err)
	}
	serializedTxHex := hex.EncodeToString(serializedTx)

	signedTx := types.SignedTransaction{ID: utils.SHA256(serializedTx), Signature: "signature", Transaction: serializedTxHex}
	bc.Mempool.Enqueue(signedTx)
	bc.ProduceBlock()

	result, _ := bc.Store.SelectNative(
		fmt.Sprintf("SELECT * from %s_ordinals;", SmartIndexAddress), []string{},
	)

	t.Log(result)
}
