package chain

import (
	"eastnode/runtime"
	"eastnode/types"
	utils "eastnode/utils/store"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/btcsuite/btcd/btcec/v2"
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

func initChainTest() *Chain {
	clearChainTest()
	if err := os.Mkdir("db_test", os.ModeDir|0755); err != nil {
		log.Panicln(err)
	}

	bc := new(Chain)
	bc.Store = utils.GetFakeInstance(utils.ChainDB)
	bc.WasmRuntime = &runtime.WasmRuntime{Store: *utils.GetFakeInstance(utils.SmartIndexDB)}
	bc.Mempool = new(Mempool)

	if err := bc.Mempool.Init(bc.Store.KV); err != nil {
		log.Panicln(err)
	}

	bc.ProduceBlock()

	return bc
}

func clearChainTest() {
	if err := os.RemoveAll("db_test"); err != nil {
		log.Panicln(err)
	}
}

func TestProcessDeploy(t *testing.T) {
	bc := initChainTest()
	defer t.Cleanup(clearChainTest)
	// create sample tx

	_, publicKey := initKey()

	wasmBytes, _ := os.ReadFile("../build/release.wasm")
	actions := []types.Action{{
		Kind:         "deploy",
		FunctionName: "",
		Args:         []string{hex.EncodeToString(wasmBytes)},
	}}

	serializedActions, err := borsh.Serialize(actions)
	if err != nil {
		t.Error(err)
	}

	serializedActionsHex := hex.EncodeToString(serializedActions)

	transaction := types.Transaction{
		Signer:  publicKey.X().String(),
		Actions: serializedActionsHex,
	}

	smartIndexAddress := bc.ProcessDeploy(transaction, actions[0])
	fmt.Println(smartIndexAddress)

	var resultSmartIndexAddress string
	var resultOwnerAddress string
	var resultWasmBlob string

	sr := bc.Store.Instance.QueryRow(fmt.Sprintf("SELECT smart_index_address, owner_address, wasm_blob FROM smart_index WHERE smart_index_address = '%s';", smartIndexAddress))

	sr.Scan(&resultSmartIndexAddress, &resultOwnerAddress, &resultWasmBlob)

	fmt.Println("smartIndexRow", resultSmartIndexAddress, resultOwnerAddress)

	SmartIndexAddress = smartIndexAddress

	if resultSmartIndexAddress != smartIndexAddress {
		t.Error("Address differ")
	}
}

func TestProcessCall(t *testing.T) {
	bc := initChainTest()
	defer t.Cleanup(clearChainTest)

	TestProcessDeploy(t)

	// Test create table

	_, publicKey := initKey()

	actions := []types.Action{{
		Kind:         "call",
		FunctionName: "init",
		Args:         []string{},
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

	bc.ProcessCall(transaction, actions[0])

	res, err := bc.WasmRuntime.Store.Instance.Query("show tables;")

	if err != nil {
		t.Error(err)
	}

	var str string
	res.Scan(&str)

	res.Next()
	res.Scan(&str)

	if !strings.Contains(str, fmt.Sprintf("%s_ordinals", SmartIndexAddress)) {
		t.Error("Call failed")
	}
}

func TestProcessRedeploy(t *testing.T) {
	bc := initChainTest()
	defer t.Cleanup(clearChainTest)

	TestProcessDeploy(t)

	// create sample tx
	_, publicKey := initKey()

	actions := []types.Action{{
		Kind:         "deploy",
		FunctionName: "",
		Args:         []string{"41414141", SmartIndexAddress},
	}}

	serializedActions, err := borsh.Serialize(actions)
	if err != nil {
		t.Error(err)
	}
	serializedActionsHex := hex.EncodeToString(serializedActions)

	transaction := types.Transaction{
		Signer:  publicKey.X().String(),
		Actions: serializedActionsHex,
	}

	bc.ProcessDeploy(transaction, actions[0])

	var resultWasmBlob string
	sr := bc.Store.Instance.QueryRow(fmt.Sprintf("SELECT wasm_blob FROM smart_index WHERE smart_index_address = '%s';", SmartIndexAddress))
	sr.Scan(&resultWasmBlob)

	if resultWasmBlob != "AAAA" {
		t.Errorf("Contract not updated uhuy %s", resultWasmBlob)
	}
}
