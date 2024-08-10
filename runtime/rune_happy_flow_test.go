package runtime

import (
	indexerDb "eastnode/indexer/repository/db"
	"eastnode/types"
	utils "eastnode/utils/store"
	"fmt"
	"log"
	"os"
	"testing"
)

func getRuneWasmRuntime() *WasmRuntime {
	instance := utils.GetFakeInstance(utils.SmartIndexDB, "../utils/store/test/rune-guguak-test/doltdump.sql")
	indexerDbRepo := indexerDb.NewDBRepository(instance.Gorm)
	wr := &WasmRuntime{Store: *instance, IndexerDbRepo: indexerDbRepo}

	return wr
}

func clearRuntimeTestRune() {
	if err := os.RemoveAll("db_test"); err != nil {
		log.Panicln(err)
	}
}

func TestRuneIndexFunction(t *testing.T) {
	defer t.Cleanup(clearRuntimeTestRune)
	wasmBytes, _ := os.ReadFile("../../example-op-return-rune/smartindex/build/release.wasm")
	wr := getRuneWasmRuntime()
	res, err := wr.RunWasmFunction("", wasmBytes, "temp", "init", []string{}, types.Call)
	if err != nil {
		t.Error(err)
	}

	res, err = wr.RunWasmFunction("", wasmBytes, "temp", "index", []string{"118", "126"}, types.Call)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(res)

	res, err = wr.RunSelectFunction("SELECT * from temp_rune_entries", []string{})
	if err != nil {
		t.Error(err)
	}
	fmt.Println(res)
	res, err = wr.RunSelectFunction("SELECT * from temp_outpoints", []string{})
	if err != nil {
		t.Error(err)
	}
	fmt.Println(res)
}
