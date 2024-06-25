package runtime

import (
	"eastnode/types"
	utils "eastnode/utils/store"
	"os"
	"testing"
)

func TestGetBlockByHeight(t *testing.T) {
	wasmBytes, _ := os.ReadFile("../build/release.wasm")

	wr := &WasmRuntime{Store: *utils.GetFakeInstance(utils.SmartIndexDB)}

	wr.RunWasmFunction("", wasmBytes, "", "index", []string{"1"}, types.Call)
}

func TestStringParamsAndResult(t *testing.T) {
	wasmBytes, _ := os.ReadFile("../build/release.wasm")

	wr := &WasmRuntime{Store: *utils.GetFakeInstance(utils.SmartIndexDB)}

	output := wr.RunWasmFunction("", wasmBytes, "", "processString", []string{"INPUT"}, types.Call)
	if output != "output for INPUT" {
		t.Error("output is incorrect")
	}
}
