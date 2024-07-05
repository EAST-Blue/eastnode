package runtime

import (
	"eastnode/indexer/repository"
	"eastnode/types"
	utils "eastnode/utils/store"
	"log"
	"os"
	"testing"
)

func getWasmRuntime() *WasmRuntime {
	instance := utils.GetFakeInstance(utils.SmartIndexDB)
	indexerRepo := repository.NewIndexerRepository(instance.Gorm)
	wr := &WasmRuntime{Store: *instance, IndexerRepo: indexerRepo}

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
	output := wr.RunWasmFunction("", wasmBytes, "", "processString", []string{"INPUT"}, types.Call)
	if output != "output for INPUT" {
		t.Error("output is incorrect")
	}
}
