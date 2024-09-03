package runtime

import (
	indexerDb "eastnode/indexer/repository/db"
	"eastnode/types"
	utils "eastnode/utils"
	store "eastnode/utils/store"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"sort"
	"strconv"
	"testing"
)

func getRuneWasmRuntime(dumpfile string) *WasmRuntime {
	os.Setenv("NETWORK", "regtest")

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
	Block  string `json:"block"`
	Tx     string `json:"tx"`
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
	wasmBytes, _ := os.ReadFile("../../runes-smart-index/smartindex/build/release.wasm")
	wr := getRuneWasmRuntime("../utils/store/test/rune/doltdump_1.sql")
	_, err := wr.RunWasmFunction("", wasmBytes, "temp", "init", []string{}, types.Call)
	if err != nil {
		t.Error(err)
	}

	_, err = wr.RunWasmFunction("", wasmBytes, "temp", "index", []string{"118", "139"}, types.Call)
	if err != nil {
		t.Error(err)
	}

	res, err := wr.RunSelectFunction("SELECT * from temp_outpoints where spent = 'false'", []string{})
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
	wasmBytes, _ := os.ReadFile("../../runes-smart-index/smartindex/build/release.wasm")
	wr := getRuneWasmRuntime("../utils/store/test/rune/doltdump_2.sql")
	_, err := wr.RunWasmFunction("", wasmBytes, "temp", "init", []string{}, types.Call)
	if err != nil {
		t.Error(err)
	}

	_, err = wr.RunWasmFunction("", wasmBytes, "temp", "index", []string{"118", "245"}, types.Call)
	if err != nil {
		t.Error(err)
	}

	res, err := wr.RunSelectFunction("SELECT * from temp_outpoints where spent = 'false'", []string{})
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

func TestRuneGetBalance(t *testing.T) {
	defer t.Cleanup(clearRuntimeTestRune)
	wasmBytes, _ := os.ReadFile("../../runes-smart-index/smartindex/build/release.wasm")
	wr := getRuneWasmRuntime("../utils/store/test/rune/doltdump_2.sql")
	_, err := wr.RunWasmFunction("", wasmBytes, "temp", "init", []string{}, types.Call)
	if err != nil {
		t.Error(err)
	}

	_, err = wr.RunWasmFunction("", wasmBytes, "temp", "index", []string{"118", "245"}, types.Call)
	if err != nil {
		t.Error(err)
	}

	res, err := wr.RunWasmFunction("", wasmBytes, "temp", "get_balance", []string{"118", "1", "bcrt1qdm7l5990ksrja0zkn46z5hpuk4rf93ym5ms9ar"}, types.Call)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(res)

	// TODO: assert the data
}

func TestRuneGetOutpointsByRuneId(t *testing.T) {
	defer t.Cleanup(clearRuntimeTestRune)
	wasmBytes, _ := os.ReadFile("../../runes-smart-index/smartindex/build/release.wasm")
	wr := getRuneWasmRuntime("../utils/store/test/rune/doltdump_2.sql")
	_, err := wr.RunWasmFunction("", wasmBytes, "temp", "init", []string{}, types.Call)
	if err != nil {
		t.Error(err)
	}

	_, err = wr.RunWasmFunction("", wasmBytes, "temp", "index", []string{"118", "245"}, types.Call)
	if err != nil {
		t.Error(err)
	}

	res, err := wr.RunWasmFunction("", wasmBytes, "temp", "get_outpoints_by_rune_id", []string{"118", "1"}, types.Call)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(res)
}

func TestRuneGetOutpointsByRuneIdAndAddress(t *testing.T) {
	defer t.Cleanup(clearRuntimeTestRune)
	wasmBytes, _ := os.ReadFile("../../runes-smart-index/smartindex/build/release.wasm")
	wr := getRuneWasmRuntime("../utils/store/test/rune/doltdump_2.sql")
	_, err := wr.RunWasmFunction("", wasmBytes, "temp", "init", []string{}, types.Call)
	if err != nil {
		t.Error(err)
	}

	_, err = wr.RunWasmFunction("", wasmBytes, "temp", "index", []string{"118", "245"}, types.Call)
	if err != nil {
		t.Error(err)
	}

	res, err := wr.RunWasmFunction("", wasmBytes, "temp", "get_outpoints_by_rune_id_and_address", []string{"118", "1", "bcrt1qdm7l5990ksrja0zkn46z5hpuk4rf93ym5ms9ar"}, types.Call)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(res)

	// TODO: assert the data
}

type runeEntry struct {
	Number       string `json:"number"`
	Block        string `json:"block"`
	Tx           string `json:"tx"`
	Minted       string `json:"minted"`
	Burned       string `json:"burned"`
	Divisibility string `json:"divisibility"`
	Premine      string `json:"premine"`
	Rune         string `json:"rune"`
	Spacers      string `json:"spacers"`
	Symbol       string `json:"symbol"`
	Turbo        string `json:"turbo"`
	Terms        string `json:"terms"`
	Amount       string `json:"amount"`
	Cap          string `json:"cap"`
	HeightStart  string `json:"height_start"`
	HeightEnd    string `json:"height_end"`
	OffsetStart  string `json:"offset_start"`
	OffsetEnd    string `json:"offset_end"`
}

type diffEntryOutpointData struct {
	Hash   string `json:"hash"`
	Index  string `json:"index"`
	Amount string `json:"amount"`
}

type diffEntryTermsData struct {
	Amount      *string `json:"amount"`
	Cap         *string `json:"cap"`
	HeightStart *int    `json:"height_start"`
	HeightEnd   *int    `json:"height_end"`
	OffsetStart *int    `json:"offset_start"`
	OffsetEnd   *int    `json:"offset_end"`
}

type diffEntryData struct {
	Number       int                     `json:"number"`
	Block        int                     `json:"block"`
	Tx           int                     `json:"tx"`
	Minted       string                  `json:"minted"`
	Burned       string                  `json:"burned"`
	Divisibility int                     `json:"divisibility"`
	Premine      string                  `json:"premine"`
	Rune         string                  `json:"rune"`
	Spacers      int                     `json:"spacers"`
	Symbol       *string                 `json:"symbol"`
	Turbo        bool                    `json:"turbo"`
	Terms        *diffEntryTermsData     `json:"terms"`
	Outpoints    []diffEntryOutpointData `json:"outpoints"`
}

func TestDumpRuneState840000_840010(t *testing.T) {
	defer t.Cleanup(clearRuntimeTestRune)

	log.Println("Preparing database & wasm runtime for testing")

	wasmBytes, _ := os.ReadFile("../../runes-smart-index/smartindex/build/release.wasm")
	wr := getRuneWasmRuntime("../utils/store/test/rune/indexer_840000_840010/doltdump.sql")
	_, err := wr.RunWasmFunction("", wasmBytes, "temp", "init", []string{}, types.Call)
	if err != nil {
		t.Error(err)
	}

	log.Println("Indexing data")
	_, err = wr.RunWasmFunction("", wasmBytes, "temp", "index", []string{"840000", "840000"}, types.Call)
	if err != nil {
		t.Error(err)
	}

	// getting outpoints
	resOutpointsRaw, err := wr.RunSelectFunction("SELECT * from temp_outpoints where spent = 'false'", []string{})
	if err != nil {
		t.Error(err)
	}

	refOutpoints := resOutpointsRaw.([]map[string]interface{})
	outpointsMap := make(map[string][]outpoint)
	for _, v := range refOutpoints {
		outpoint := outpoint{}
		_ = utils.MapToStruct(v, &outpoint)
		key := outpoint.Block + "_" + outpoint.Tx
		outpointsMap[key] = append(outpointsMap[key], outpoint)
	}

	// gettings rune entries
	resRuneEntriesRaw, err := wr.RunSelectFunction("SELECT * from temp_rune_entries", []string{})
	if err != nil {
		t.Error(err)
	}
	refRuneEntries := resRuneEntriesRaw.([]map[string]interface{})
	entries := make([]diffEntryData, len(refRuneEntries))

	for i, v := range refRuneEntries {
		entry := runeEntry{}
		_ = utils.MapToStruct(v, &entry)

		outpoints := outpointsMap[entry.Block+"_"+entry.Tx]
		entryOutpoints := make([]diffEntryOutpointData, len(outpoints))
		for i, outpoint := range outpoints {
			entryOutpoints[i] = diffEntryOutpointData{
				Hash:   outpoint.Hash,
				Index:  outpoint.Vout,
				Amount: outpoint.Amount,
			}
		}

		sort.Slice(entryOutpoints, func(i, j int) bool {
			if entryOutpoints[i].Hash != entryOutpoints[j].Hash {
				return entryOutpoints[i].Hash < entryOutpoints[j].Hash
			}
			iIndex, _ := strconv.Atoi(entryOutpoints[i].Index)
			jIndex, _ := strconv.Atoi(entryOutpoints[j].Index)
			return iIndex < jIndex
		})

		number, _ := strconv.Atoi(entry.Number)
		block, _ := strconv.Atoi(entry.Block)
		tx, _ := strconv.Atoi(entry.Tx)
		divisibility, _ := strconv.Atoi(entry.Divisibility)
		spacers, _ := strconv.Atoi(entry.Spacers)
		premine := "0"
		if entry.Premine != "\\N" {
			premine = entry.Premine
		}

		data := diffEntryData{
			Number:       number,
			Block:        block,
			Tx:           tx,
			Minted:       entry.Minted,
			Burned:       entry.Burned,
			Divisibility: divisibility,
			Premine:      premine,
			Rune:         entry.Rune,
			Spacers:      spacers,
			Symbol:       &entry.Symbol,
			Turbo:        entry.Turbo == "true",
			Outpoints:    entryOutpoints,
		}
		if entry.Terms == "true" {
			terms := &diffEntryTermsData{}

			terms.Amount = &entry.Amount
			terms.Cap = &entry.Cap

			if entry.HeightStart != "\\N" {
				heightStart, _ := strconv.Atoi(entry.HeightStart)
				terms.HeightStart = &heightStart
			}

			if entry.HeightEnd != "\\N" {
				heightEnd, _ := strconv.Atoi(entry.HeightEnd)
				terms.HeightEnd = &heightEnd
			}

			if entry.OffsetStart != "\\N" {
				offsetStart, _ := strconv.Atoi(entry.OffsetStart)
				terms.OffsetStart = &offsetStart
			}

			if entry.OffsetEnd != "\\N" {
				offsetEnd, _ := strconv.Atoi(entry.OffsetEnd)
				terms.OffsetEnd = &offsetEnd
			}

			data.Terms = terms
		}

		entries[i] = data
	}

	latestIndexedBlock, err := wr.RunWasmFunction("", wasmBytes, "temp", "get_latest_indexed_block", []string{}, types.Call)
	if err != nil {
		t.Error(err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Number < entries[j].Number
	})

	jsonStr, _ := json.MarshalIndent(entries, "", "  ")
	os.WriteFile(fmt.Sprintf("../../runes-diff-tests/smartindex-states/%s.json", latestIndexedBlock), []byte(jsonStr), 0644)
}
