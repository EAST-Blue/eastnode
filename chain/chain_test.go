package chain

import (
	"eastnode/types"
	utils "eastnode/utils/store"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/near/borsh-go"
)

func initChainTest() *Chain {
	clearChainTest()
	if err := os.Mkdir("db_test", os.ModeDir|0755); err != nil {
		log.Panicln(err)
	}

	bc := new(Chain)
	bc.Store = utils.GetFakeInstance()
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
	keyHex := "7c67c815e1c4a25fe70d95aad9440b682bdcbe6e2baf34d460966e605705ea8e"

	privateKeyBytes, err := hex.DecodeString(keyHex)
	if err != nil {
		panic(err)
	}

	_, publicKey := btcec.PrivKeyFromBytes(privateKeyBytes)

	actions := []types.Action{{
		Kind:         "deploy",
		FunctionName: "",
		Args:         []string{"000102030405060708090A0B0C0D"},
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

	var resultSmartIndexAddress string
	var resultOwnerAddress string
	var resultWasmBlob string
	sr := bc.Store.Instance.QueryRow(fmt.Sprintf("SELECT smart_index_address, owner_address, wasm_blob FROM smart_index WHERE smart_index_address = '%s';", smartIndexAddress))

	sr.Scan(&resultSmartIndexAddress, &resultOwnerAddress, &resultWasmBlob)

	fmt.Println("smartIndexRow", resultSmartIndexAddress, resultOwnerAddress, resultWasmBlob)

	if resultSmartIndexAddress != smartIndexAddress {
		t.Error("Address differ")
	}
}

func TestProcessRedeploy(t *testing.T) {
	bc := initChainTest()
	defer t.Cleanup(clearChainTest)

	TestProcessDeploy(t)

	// create sample tx
	keyHex := "7c67c815e1c4a25fe70d95aad9440b682bdcbe6e2baf34d460966e605705ea8e"

	privateKeyBytes, err := hex.DecodeString(keyHex)
	if err != nil {
		panic(err)
	}

	_, publicKey := btcec.PrivKeyFromBytes(privateKeyBytes)

	actions := []types.Action{{
		Kind:         "deploy",
		FunctionName: "",
		Args:         []string{"41414141", "idx175sfmus32u8l6h2fyjj20lh0cek3un3j7xknww4rcufskgqxwecss0smcs"},
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

	var resultWasmBlob string
	sr := bc.Store.Instance.QueryRow(fmt.Sprintf("SELECT wasm_blob FROM smart_index WHERE smart_index_address = '%s';", smartIndexAddress))
	sr.Scan(&resultWasmBlob)

	if resultWasmBlob != "AAAA" {
		t.Error("Contract not updated")
	}
}
