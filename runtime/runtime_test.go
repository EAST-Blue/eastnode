package runtime

import (
	"eastnode/types"
	utils "eastnode/utils/store"
	"log"
	"os"
	"testing"
)

func clearRuntimeTest() {
	if err := os.RemoveAll("db_test"); err != nil {
		log.Panicln(err)
	}
}

func TestGetBlockByHeight(t *testing.T) {
	defer t.Cleanup(clearRuntimeTest)
	wasmBytes, _ := os.ReadFile("../build/release.wasm")

	wr := &WasmRuntime{Store: *utils.GetFakeInstance(utils.SmartIndexDB)}

	wr.RunWasmFunction("", wasmBytes, "", "index", []string{"1"}, types.Call)
}

func TestStringParamsAndResult(t *testing.T) {
	defer t.Cleanup(clearRuntimeTest)
	wasmBytes, _ := os.ReadFile("../build/release.wasm")

	wr := &WasmRuntime{Store: *utils.GetFakeInstance(utils.SmartIndexDB)}

	output := wr.RunWasmFunction("", wasmBytes, "", "processString", []string{"INPUT"}, types.Call)
	if output != "output for INPUT" {
		t.Error("output is incorrect")
	}
}
