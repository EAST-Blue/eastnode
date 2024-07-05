package runtime

import (
	"bytes"
	"context"
	"eastnode/indexer/repository"
	"eastnode/types"
	store "eastnode/utils/store"
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
	Store       store.Store
	Mod         api.Module
	IndexerRepo *repository.IndexerRepository
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

func (r *WasmRuntime) loadWasm(wasmBytes []byte, ctx context.Context, smartIndexAddress string, signer Address, kind types.ActionKind, output *string) api.Module {
	wazeroRuntime := wazero.NewRuntime(ctx)

	envBuilder := wazeroRuntime.
		NewHostModuleBuilder("env").
		NewFunctionBuilder().
		WithFunc(func(tableName int32, primaryKey int32, tableSchema int32) int32 {
			if kind != types.Call {
				log.Panicln("Cannot call function on view")
				return 0
			}
			tableNameStr := ToString(r.Mod.Memory(), int64(tableName))
			primaryKeyStr := ToString(r.Mod.Memory(), int64(primaryKey))
			tableSchemaStr := ToString(r.Mod.Memory(), int64(tableSchema))

			CreateTable(r.Store, smartIndexAddress, tableNameStr, primaryKeyStr, tableSchemaStr)

			// WORKAROUND: showTables()
			r.Store.ShowTables()
			return 0
		}).
		Export("createTable").
		NewFunctionBuilder().
		WithFunc(func(tableName int32, values int32) int32 {
			if kind != types.Call {
				log.Panicln("Cannot call function on view")
				return 0
			}
			tableNameStr := ToString(r.Mod.Memory(), int64(tableName))
			valuesStr := ToString(r.Mod.Memory(), int64(values))

			Insert(r.Store, smartIndexAddress, tableNameStr, valuesStr)

			// WORKAROUND: showTables()
			r.Store.ShowTables()
			return 0
		}).
		Export("insertItem").
		NewFunctionBuilder().
		WithFunc(func(tableName int32, whereCondition int32, values int32) int32 {
			if kind != types.Call {
				log.Panicln("Cannot call function on view")
				return 0
			}
			tableNameStr := ToString(r.Mod.Memory(), int64(tableName))
			whereConditionStr := ToString(r.Mod.Memory(), int64(whereCondition))
			valuesStr := ToString(r.Mod.Memory(), int64(values))

			Update(r.Store, smartIndexAddress, tableNameStr, whereConditionStr, valuesStr)

			// WORKAROUND: showTables()
			r.Store.ShowTables()
			return 0
		}).
		Export("updateItem").
		NewFunctionBuilder().
		WithFunc(func(tableName int32, whereCondition int32) int32 {
			if kind != types.Call {
				log.Panicln("Cannot call function on view")
				return 0
			}
			tableNameStr := ToString(r.Mod.Memory(), int64(tableName))
			whereConditionStr := ToString(r.Mod.Memory(), int64(whereCondition))

			Delete(r.Store, smartIndexAddress, tableNameStr, whereConditionStr)

			// WORKAROUND: showTables()
			r.Store.ShowTables()
			return 0
		}).
		Export("deleteItem").
		NewFunctionBuilder().
		WithFunc(func(tableName int32, whereCondition int32) uint32 {
			tableNameStr := ToString(r.Mod.Memory(), int64(tableName))
			whereConditionStr := ToString(r.Mod.Memory(), int64(whereCondition))

			result := Select(r.Store, smartIndexAddress, tableNameStr, whereConditionStr)

			ptr := r.writeString(r.Mod.Memory(), result)

			return uint32(ptr)
		}).
		Export("selectItem").
		NewFunctionBuilder().
		WithFunc(func(height int64) uint32 {
			result, err := r.IndexerRepo.GetBlockByHeight(height)
			if err != nil {
				panic(err)
			}

			serializedResult, _ := json.Marshal(result)

			ptr := r.writeString(r.Mod.Memory(), string(serializedResult))

			return uint32(ptr)
		}).
		Export("getBlockByHeight").
		NewFunctionBuilder().
		WithFunc(func(blockHash int32) uint32 {
			blockHashStr := ToString(r.Mod.Memory(), int64(blockHash))
			result, err := r.IndexerRepo.GetTransactionsByBlockHash(blockHashStr)
			if err != nil {
				panic(err)
			}

			serializedResult, _ := json.Marshal(result)

			ptr := r.writeString(r.Mod.Memory(), string(serializedResult))

			return uint32(ptr)
		}).
		Export("getTransactionsByBlockHash").
		NewFunctionBuilder().
		WithFunc(func(transactionHash int32) uint32 {
			transactionHashStr := ToString(r.Mod.Memory(), int64(transactionHash))
			result, err := r.IndexerRepo.GetOutpointsByTransactionHash(transactionHashStr)
			if err != nil {
				panic(err)
			}
			serializedResult, _ := json.Marshal(result)

			ptr := r.writeString(r.Mod.Memory(), string(serializedResult))

			return uint32(ptr)
		}).
		Export("getOutpointsByTransactionHash").
		NewFunctionBuilder().
		WithFunc(func(strPtr int32) {
			str := ToString(r.Mod.Memory(), int64(strPtr))

			*output = str
		}).
		Export("valueReturn").
		NewFunctionBuilder().
		WithFunc(func(strPtr int32) {
			str := ToString(r.Mod.Memory(), int64(strPtr))

			log.Panicln(str)
		}).
		Export("panic").
		NewFunctionBuilder().
		WithFunc(func(strPtr int32) {
			str := ToString(r.Mod.Memory(), int64(strPtr))

			fmt.Println("consoleLog", str)
		}).
		Export("consoleLog")

	assemblyscript.NewFunctionExporter().
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

func (r *WasmRuntime) RunWasmFunction(signer Address, wasmBytes []byte, smartIndexAddress string, functionName string, args []string, kind types.ActionKind) any {
	ctx := context.Background()
	defer ctx.Done()

	var output string
	mod := r.loadWasm(wasmBytes, ctx, smartIndexAddress, signer, kind, &output)
	f := mod.ExportedFunction(functionName)

	// All arguments are stringified pointers
	argsPtr := make([]uint64, len(args))
	for i := range argsPtr {
		ptr := r.writeString(r.Mod.Memory(), args[i])
		argsPtr[i] = uint64(ptr)
	}

	_, err := f.Call(ctx, argsPtr...)

	if err != nil {
		fmt.Println("Error", err)
	}

	return output
}
