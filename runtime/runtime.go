package runtime

import (
	"bytes"
	"context"
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
	Store store.Store
	Mod   api.Module
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
			// TODO: Remove mocked data
			result := types.BitcoinBlockHeader{
				ID:            1,
				Version:       1,
				Height:        0,
				PreviousBlock: "0000000000000000000000000000000000000000000000000000000000000000",
				MerkleRoot:    "4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b",
				Hash:          "000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f",
				Time:          1231006505000,
				Nonce:         2083236893,
				Bits:          486604799,
			}

			serializedResult, _ := json.Marshal(result)

			ptr := r.writeString(r.Mod.Memory(), string(serializedResult))

			return uint32(ptr)
		}).
		Export("getBlockByHeight").
		NewFunctionBuilder().
		WithFunc(func(block_hash int32) uint32 {
			// ^ the block_hash above is ptr for string
			// TODO: Remove mocked data
			result := []types.BitcoinTransaction{{

				ID:        1,
				Hash:      "3a6d490a7cf40819cdd826729d921ad5ab4b8347dfbec81179dd09aba0d25b37",
				BlockHash: "000000009a940db389f3a7cbb8405f4ec14342bed36073b60ee63ed7e117f193",
				BlockId:   189,
				LockTime:  0,
				Version:   1,
				Safe:      0,
			}}

			serializedResult, _ := json.Marshal(result)

			ptr := r.writeString(r.Mod.Memory(), string(serializedResult))

			return uint32(ptr)
		}).
		Export("getTransactionsByBlockHash").
		NewFunctionBuilder().
		WithFunc(func(tx_hash int32) uint32 {
			// ^ the tx_hash above is ptr for string
			// TODO: Remove mocked data
			result := []types.BitcoinOutpoint{{
				ID:              1,
				Value:           5000000000,
				SpendingTxId:    0,
				SpendingTxHash:  "",
				SpendingTxIndex: 0,
				Sequence:        0,
				FundingTxId:     194,
				FundingTxHash:   "3a6d490a7cf40819cdd826729d921ad5ab4b8347dfbec81179dd09aba0d25b37",
				FundingTxIndex:  0,
				SignatureScript: "",
				PkScript:        "410449fff9665bfda43017a27b3d32e986378befdd6fa5d4eb097626701ace807a2b3a43e74375dce4ed9028b3b62ba8485358cd48967e854a857a38ecdbfe5b62f8ac",
				Witness:         "",
				Spender:         "",
				Type:            "nonstandard",
			}}

			serializedResult, _ := json.Marshal(result)

			ptr := r.writeString(r.Mod.Memory(), string(serializedResult))

			return uint32(ptr)
		}).
		Export("getOutpointsByTransactionHash").
		NewFunctionBuilder().
		WithFunc(func(strPtr int32) {
			// DEBUG
			str := ToString(r.Mod.Memory(), int64(strPtr))

			*output = str
		}).
		Export("valueReturn").
		NewFunctionBuilder().
		WithFunc(func(strPtr int32) {
			// DEBUG
			str := ToString(r.Mod.Memory(), int64(strPtr))

			log.Panicln(str)
		}).
		Export("panic").
		NewFunctionBuilder().
		WithFunc(func(strPtr int32) {
			// DEBUG
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
