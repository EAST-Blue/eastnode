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

	wr.RunWasmFunction("", wasmBytes, "", "index", []uint64{1}, types.Call)
}
