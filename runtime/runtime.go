package runtime

import (
	"bytes"
	"context"
	store "eastnode/utils/store"
	utils "eastnode/utils/store"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"unicode/utf16"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/assemblyscript"
)

var (
	LE = binary.LittleEndian
)

const (
	SizeOffset = -4
)

type Address string

type WasmRuntime struct {
	s   store.Store
	Mod api.Module
}

// ref: https://github.com/RPG-18/wasmer-go-assemblyscript/blob/main/assemblyscript/go.ts
// https://github.com/tetratelabs/wazero/blob/54cee893dac6fb85d9418b7f1e156974e7e05b00/imports/assemblyscript/assemblyscript.go#L302
func ToString(memory api.Memory, ptr int64) string {
	data, err := memory.Read(0, memory.Size())
	if err != true {
		log.Panicln("read failed")
	}
	len := LE.Uint32(data[ptr+SizeOffset:]) >> 1
	buf := bytes.NewReader(data[ptr:])

	tmp := make([]uint16, 0, len)
	for i := uint32(0); i < len; i++ {
		var j uint16
		_ = binary.Read(buf, LE, &j)
		tmp = append(tmp, j)
	}
	return string(utf16.Decode(tmp))
}

func (r *WasmRuntime) writeString(memory api.Memory, str string) uint32 {
	ctx := context.Background()

	alloc := r.Mod.ExportedFunction("allocate")

	result, err := alloc.Call(ctx, uint64(len(str)+4))

	if err != nil {
		log.Panicln(result)
	}

	lenOffset := uint32(result[0])
	stringOffset := (lenOffset + 4)

	memory.WriteUint32Le(lenOffset, uint32(len(str)<<1))

	runes := utf16.Encode([]rune(str))
	current_i := 0
	for _, r := range runes {
		bytes_1 := byte(r >> 8)
		bytes_0 := byte(r & 255)
		memory.WriteByte(stringOffset+uint32(current_i), bytes_0)
		memory.WriteByte(stringOffset+uint32(current_i+1), bytes_1)
		current_i += 2
	}

	return uint32(stringOffset)
}

// func (r *WasmRuntime) UploadWASM(wasmBytes []byte) int64 {
// 	ctx := context.Background()
//
// 	h := sha1.New()
// 	h.Write(wasmBytes)
// 	sha1Hash := hex.EncodeToString(h.Sum(nil))
//
// 	rowWasmBytes := &utils.WasmBytes{Hash: sha1Hash, Bytes: wasmBytes}
//
// 	var result = &utils.WasmBytes{}
// 	err := r.s.BunInstance.NewSelect().Model(result).Where("hash = ?", sha1Hash).Scan(ctx)
//
// 	if err != nil {
// 		_, err = r.s.BunInstance.NewInsert().Model(rowWasmBytes).Exec(ctx)
// 		if err != nil {
// 			log.Panicln(err)
// 		}
// 		r.s.BunInstance.NewSelect().Model(result).Where("hash = ?", sha1Hash).Scan(ctx)
// 		return result.ID
// 	} else {
// 		return result.ID
// 	}
//
// }

func (r *WasmRuntime) ParseAndRunTx() {
	// TODO: get caller address, contract, functionName, and params.
}

func (r *WasmRuntime) loadWasm(wasmBytes []byte, ctx context.Context) api.Module {
	wazeroRuntime := wazero.NewRuntime(ctx)

	envBuilder := wazeroRuntime.
		NewHostModuleBuilder("env").
		NewFunctionBuilder().
		WithFunc(func(tableName int32, primaryKey int32, tableSchema int32) int32 {
			tableNameStr := ToString(r.Mod.Memory(), int64(tableName))
			primaryKeyStr := ToString(r.Mod.Memory(), int64(primaryKey))
			tableSchemaStr := ToString(r.Mod.Memory(), int64(tableSchema))

			// TODO: get contract address from blockchain
			CreateTable(r.s, "contractAddress", tableNameStr, primaryKeyStr, tableSchemaStr)

			// WORKAROUND: showTables()
			r.s.ShowTables()
			return 0
		}).
		Export("createTable").
		NewFunctionBuilder().
		WithFunc(func(tableName int32, values int32) int32 {
			tableNameStr := ToString(r.Mod.Memory(), int64(tableName))
			valuesStr := ToString(r.Mod.Memory(), int64(values))

			// TODO: get contract address from blockchain
			Insert(r.s, "contractAddress", tableNameStr, valuesStr)

			// WORKAROUND: showTables()
			r.s.ShowTables()
			return 0
		}).
		Export("insertItem").
		NewFunctionBuilder().
		WithFunc(func(tableName int32, whereCondition int32, values int32) int32 {
			tableNameStr := ToString(r.Mod.Memory(), int64(tableName))
			whereConditionStr := ToString(r.Mod.Memory(), int64(whereCondition))
			valuesStr := ToString(r.Mod.Memory(), int64(values))

			// TODO: get contract address from blockchain
			Update(r.s, "contractAddress", tableNameStr, whereConditionStr, valuesStr)

			// WORKAROUND: showTables()
			r.s.ShowTables()
			return 0
		}).
		Export("updateItem").
		NewFunctionBuilder().
		WithFunc(func(tableName int32, whereCondition int32) int32 {
			tableNameStr := ToString(r.Mod.Memory(), int64(tableName))
			whereConditionStr := ToString(r.Mod.Memory(), int64(whereCondition))

			// TODO: get contract address from blockchain
			Delete(r.s, "contractAddress", tableNameStr, whereConditionStr)

			// WORKAROUND: showTables()
			r.s.ShowTables()
			return 0
		}).
		Export("deleteItem").
		NewFunctionBuilder().
		WithFunc(func(tableName int32, whereCondition int32) uint32 {
			tableNameStr := ToString(r.Mod.Memory(), int64(tableName))
			whereConditionStr := ToString(r.Mod.Memory(), int64(whereCondition))

			// TODO: get contract address from blockchain
			result := Select(r.s, "contractAddress", tableNameStr, whereConditionStr)

			ptr := r.writeString(r.Mod.Memory(), result)

			return uint32(ptr)
		}).
		Export("selectItem").
		NewFunctionBuilder().
		WithFunc(func(strPtr int32) {
			str := ToString(r.Mod.Memory(), int64(strPtr))

			fmt.Println("consoleLog", str)
		}).
		Export("consoleLog")

	assemblyscript.NewFunctionExporter().
		WithAbortMessageDisabled().
		ExportFunctions(envBuilder)

	_, err := envBuilder.Instantiate(ctx)

	if err != nil {
		log.Panicln(err)
	}

	mod, err := wazeroRuntime.InstantiateWithConfig(ctx, wasmBytes,
		wazero.NewModuleConfig().WithStdout(os.Stdout).WithStderr(os.Stderr))
	if err != nil {
		log.Panicln(err)
	}

	r.Mod = mod

	return mod
}

func (r *WasmRuntime) RunWasmFunction(caller Address, codeId int, functionName string, jsonParams string) any {
	ctx := context.Background()
	defer ctx.Done()

	var wasmBytes = &utils.WasmBytes{}
	r.s.BunInstance.NewSelect().Model(wasmBytes).Where("id = ?", codeId).Scan(ctx)

	var objmap map[string]any
	if err := json.Unmarshal([]byte(jsonParams), &objmap); err != nil {
		log.Fatal(err)
	}

	mod := r.loadWasm(wasmBytes.Bytes, ctx)
	f := mod.ExportedFunction(functionName)
	result, err := f.Call(ctx)
	// result, err := f.Call(ctx, uint64(objmap["first"].(float64)), uint64(objmap["second"].(float64)))
	// result, err := f.Call(ctx)

	if err != nil {
		log.Panicln(err)
	}

	return result
}
