package runtime

import (
	indexerDb "eastnode/indexer/repository/db"
	"eastnode/types"
	utils "eastnode/utils"
	store "eastnode/utils/store"
	"log"
	"os"
	"reflect"
	"sort"
	"testing"
)

func getRuneWasmRuntime(dumpfile string) *WasmRuntime {
	instance := store.GetFakeInstance(store.SmartIndexDB, dumpfile)
	indexerDbRepo := indexerDb.NewDBRepository(instance.Gorm)
	wr := &WasmRuntime{Store: *instance, IndexerDbRepo: indexerDbRepo}

	return wr
}

func clearRuntimeTestRune() {
	if err := os.RemoveAll("db_test"); err != nil {
		log.Panicln(err)
	}
}

type outpoint struct {
	Hash   string `json:"hash"`
	Vout   string `json:"vout"`
	Amount string `json:"amount"`
}

func sortOutpoints(outpoints []outpoint) {
	sort.Slice(outpoints, func(i, j int) bool {
		if outpoints[i].Hash != outpoints[j].Hash {
			return outpoints[i].Hash < outpoints[j].Hash
		}
		if outpoints[i].Vout != outpoints[j].Vout {
			return outpoints[i].Vout < outpoints[j].Vout
		}
		return outpoints[i].Amount < outpoints[j].Amount
	})
}

func TestRuneIndexOutput1(t *testing.T) {
	defer t.Cleanup(clearRuntimeTestRune)
	wasmBytes, _ := os.ReadFile("../../example-op-return-rune/smartindex/build/release.wasm")
	wr := getRuneWasmRuntime("../utils/store/test/rune/doltdump_1.sql")
	_, err := wr.RunWasmFunction("", wasmBytes, "temp", "init", []string{}, types.Call)
	if err != nil {
		t.Error(err)
	}

	_, err = wr.RunWasmFunction("", wasmBytes, "temp", "index", []string{"118", "139"}, types.Call)
	if err != nil {
		t.Error(err)
	}

	res, err := wr.RunSelectFunction("SELECT * from temp_outpoints", []string{})
	if err != nil {
		t.Error(err)
	}

	ref := res.([]map[string]interface{})
	outpoints := make([]outpoint, len(ref))

	for i, v := range ref {
		outpoint := outpoint{}
		_ = utils.MapToStruct(v, &outpoint)
		outpoints[i] = outpoint
	}

	expectedOutpoints := []outpoint{
		{
			Hash:   "428ac779b02094ae4009626d0b5e03727bec57a44207e025fbe8a641a9d9cb82",
			Vout:   "0",
			Amount: "35",
		},
		{
			Hash:   "428ac779b02094ae4009626d0b5e03727bec57a44207e025fbe8a641a9d9cb82",
			Vout:   "1",
			Amount: "13",
		},
		{
			Hash:   "428ac779b02094ae4009626d0b5e03727bec57a44207e025fbe8a641a9d9cb82",
			Vout:   "2",
			Amount: "13",
		},
		{
			Hash:   "428ac779b02094ae4009626d0b5e03727bec57a44207e025fbe8a641a9d9cb82",
			Vout:   "3",
			Amount: "13",
		},
		{
			Hash:   "428ac779b02094ae4009626d0b5e03727bec57a44207e025fbe8a641a9d9cb82",
			Vout:   "4",
			Amount: "13",
		},
		{
			Hash:   "428ac779b02094ae4009626d0b5e03727bec57a44207e025fbe8a641a9d9cb82",
			Vout:   "6",
			Amount: "13",
		},
		{
			Hash:   "ff419a53bc5acf9d12b2a175c88bfef651793065cd56b007458e43b9eb0e1185",
			Vout:   "0",
			Amount: "28",
		},
		{
			Hash:   "ff419a53bc5acf9d12b2a175c88bfef651793065cd56b007458e43b9eb0e1185",
			Vout:   "1",
			Amount: "83",
		},
		{
			Hash:   "ff419a53bc5acf9d12b2a175c88bfef651793065cd56b007458e43b9eb0e1185",
			Vout:   "2",
			Amount: "79",
		},
		{
			Hash:   "c3775672b925b7173dea8d9da5fc6db5b214658091de31818fee38701f6819a3",
			Vout:   "1",
			Amount: "10",
		},
	}

	sortOutpoints(outpoints)
	sortOutpoints(expectedOutpoints)

	if !reflect.DeepEqual(outpoints, expectedOutpoints) {
		t.Error("outputs are incorrect")
	}
}

func TestRuneIndexOutput2(t *testing.T) {
	defer t.Cleanup(clearRuntimeTestRune)
	wasmBytes, _ := os.ReadFile("../../example-op-return-rune/smartindex/build/release.wasm")
	wr := getRuneWasmRuntime("../utils/store/test/rune/doltdump_2.sql")
	_, err := wr.RunWasmFunction("", wasmBytes, "temp", "init", []string{}, types.Call)
	if err != nil {
		t.Error(err)
	}

	_, err = wr.RunWasmFunction("", wasmBytes, "temp", "index", []string{"118", "245"}, types.Call)
	if err != nil {
		t.Error(err)
	}

	res, err := wr.RunSelectFunction("SELECT * from temp_outpoints", []string{})
	if err != nil {
		t.Error(err)
	}

	ref := res.([]map[string]interface{})
	outpoints := make([]outpoint, len(ref))

	for i, v := range ref {
		outpoint := outpoint{}
		_ = utils.MapToStruct(v, &outpoint)
		outpoints[i] = outpoint
	}

	expectedOutpoints := []outpoint{
		{
			Hash:   "59f5afee2338faeeaab3d4403d8099b511ee826e6f6b508cf767fef26c2aa91f",
			Vout:   "1",
			Amount: "50",
		},
		{
			Hash:   "59f5afee2338faeeaab3d4403d8099b511ee826e6f6b508cf767fef26c2aa91f",
			Vout:   "2",
			Amount: "50",
		},
		{
			Hash:   "59f5afee2338faeeaab3d4403d8099b511ee826e6f6b508cf767fef26c2aa91f",
			Vout:   "3",
			Amount: "50",
		},
		{
			Hash:   "59f5afee2338faeeaab3d4403d8099b511ee826e6f6b508cf767fef26c2aa91f",
			Vout:   "4",
			Amount: "50",
		},
		{
			Hash:   "59f5afee2338faeeaab3d4403d8099b511ee826e6f6b508cf767fef26c2aa91f",
			Vout:   "6",
			Amount: "50",
		},
		{
			Hash:   "d9889c573ff5eb2befc28f0b45e374fe41a462b8273d5e1bdc419795e93a6152",
			Vout:   "0",
			Amount: "9",
		},
		{
			Hash:   "d9889c573ff5eb2befc28f0b45e374fe41a462b8273d5e1bdc419795e93a6152",
			Vout:   "1",
			Amount: "9",
		},
		{
			Hash:   "d9889c573ff5eb2befc28f0b45e374fe41a462b8273d5e1bdc419795e93a6152",
			Vout:   "2",
			Amount: "8",
		},
		{
			Hash:   "d9889c573ff5eb2befc28f0b45e374fe41a462b8273d5e1bdc419795e93a6152",
			Vout:   "3",
			Amount: "8",
		},
		{
			Hash:   "d9889c573ff5eb2befc28f0b45e374fe41a462b8273d5e1bdc419795e93a6152",
			Vout:   "4",
			Amount: "8",
		},
		{
			Hash:   "d9889c573ff5eb2befc28f0b45e374fe41a462b8273d5e1bdc419795e93a6152",
			Vout:   "6",
			Amount: "8",
		},
	}

	sortOutpoints(outpoints)
	sortOutpoints(expectedOutpoints)

	if !reflect.DeepEqual(outpoints, expectedOutpoints) {
		t.Error("outputs are incorrect")
	}
}
