package runtime

import (
	indexerDb "eastnode/indexer/repository/db"
	"eastnode/types"
	utils "eastnode/utils/store"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"
)

func getWasmRuntime() *WasmRuntime {
	instance := utils.GetFakeInstance(utils.SmartIndexDB, "../utils/store/test/doltdump.sql")
	indexerDbRepo := indexerDb.NewDBRepository(instance.Gorm)
	wr := &WasmRuntime{Store: *instance, IndexerDbRepo: indexerDbRepo}

	return wr
}

func clearRuntimeTest() {
	if err := os.RemoveAll("db_test"); err != nil {
		log.Panicln(err)
	}
}

func TestGetBlockByHeight(t *testing.T) {
	defer t.Cleanup(clearRuntimeTest)
	wasmBytes, _ := os.ReadFile("../build/release.wasm")

	wr := getWasmRuntime()
	wr.RunWasmFunction("", wasmBytes, "", "index", []string{"20"}, types.Call)
	wr.RunWasmFunction("", wasmBytes, "", "index", []string{"111"}, types.Call)
}

func TestStringParamsAndResult(t *testing.T) {
	defer t.Cleanup(clearRuntimeTest)
	wasmBytes, _ := os.ReadFile("../build/release.wasm")

	wr := getWasmRuntime()
	output, _ := wr.RunWasmFunction("", wasmBytes, "", "processString", []string{"INPUT"}, types.Call)
	if output != "output for INPUT" {
		t.Error("output is incorrect")
	}
}

func TestRunSelectFunction(t *testing.T) {
	defer t.Cleanup(clearRuntimeTest)
	wasmBytes, _ := os.ReadFile("../build/release.wasm")
	wr := getWasmRuntime()
	wr.RunWasmFunction("", wasmBytes, "temp", "init", []string{}, types.Call)
	wr.RunWasmFunction("", wasmBytes, "temp", "insertItemTest", []string{}, types.Call)

	res, err := wr.RunSelectFunction("SELECT * from temp_ordinals", []string{})

	if err != nil {
		t.Error(err)
	}
	fmt.Println(res)

	// result can be json marshalled
	marshalled, _ := json.Marshal(res)
	fmt.Println(string(marshalled))

	// selectNative can also be used from cross-index
	wr.RunWasmFunction("", wasmBytes, "temp", "selectNativeTest", []string{}, types.Call)
}
