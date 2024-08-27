package runtime

import (
	"bytes"
	"context"
	indexerDb "eastnode/indexer/repository/db"
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

type Address string

type WasmRuntime struct {
	Store         store.Store
	Mod           api.Module
	IndexerDbRepo *indexerDb.DBRepository
}

// ref: https://github.com/RPG-18/wasmer-go-assemblyscript/blob/main/assemblyscript/go.ts
// https://github.com/tetratelabs/wazero/blob/54cee893dac6fb85d9418b7f1e156974e7e05b00/imports/assemblyscript/assemblyscript.go#L302
func ToString(memory api.Memory, ptr uint32) string {
	data, err := memory.Read(ptr-4, 8)
	if err != true {
		log.Panicln("read failed")
	}
	len := LE.Uint32(data) >> 1

	dataBuf, err := memory.Read(ptr, len*2)
	if err != true {
		log.Panicln("read failed")
	}
	buf := bytes.NewReader(dataBuf)

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

func (r *WasmRuntime) loadWasm(wasmBytes []byte, ctx context.Context, smartIndexAddress string, signer Address, kind types.ActionKind, output *string, errorMessage *error) api.Module {
	wazeroRuntime := wazero.NewRuntime(ctx)

	envBuilder := wazeroRuntime.
		NewHostModuleBuilder("env").
		NewFunctionBuilder().
		WithFunc(func(tableName uint32, tableSchema uint32, option uint32) uint32 {
			if kind != types.Call {
				log.Panicln("Cannot call function on view")
				return 0
			}
			tableNameStr := ToString(r.Mod.Memory(), tableName)
			tableSchemaStr := ToString(r.Mod.Memory(), tableSchema)
			optionStr := ToString(r.Mod.Memory(), option)

			CreateTable(r.Store, smartIndexAddress, tableNameStr, tableSchemaStr, optionStr)

			return 0
		}).
		Export("createTable").
		NewFunctionBuilder().
		WithFunc(func(tableName uint32, values uint32) uint32 {
			if kind != types.Call {
				log.Panicln("Cannot call function on view")
				return 0
			}
			tableNameStr := ToString(r.Mod.Memory(), tableName)
			valuesStr := ToString(r.Mod.Memory(), values)

			Insert(r.Store, smartIndexAddress, tableNameStr, valuesStr)

			return 0
		}).
		Export("insertItem").
		NewFunctionBuilder().
		WithFunc(func(tableName uint32, whereCondition uint32, values uint32) uint32 {
			if kind != types.Call {
				log.Panicln("Cannot call function on view")
				return 0
			}
			tableNameStr := ToString(r.Mod.Memory(), tableName)
			whereConditionStr := ToString(r.Mod.Memory(), whereCondition)
			valuesStr := ToString(r.Mod.Memory(), values)

			Update(r.Store, smartIndexAddress, tableNameStr, whereConditionStr, valuesStr)

			return 0
		}).
		Export("updateItem").
		NewFunctionBuilder().
		WithFunc(func(tableName uint32, whereCondition uint32) uint32 {
			if kind != types.Call {
				log.Panicln("Cannot call function on view")
				return 0
			}
			tableNameStr := ToString(r.Mod.Memory(), tableName)
			whereConditionStr := ToString(r.Mod.Memory(), whereCondition)

			Delete(r.Store, smartIndexAddress, tableNameStr, whereConditionStr)

			return 0
		}).
		Export("deleteItem").
		NewFunctionBuilder().
		WithFunc(func(tableName uint32, whereCondition uint32) uint32 {
			tableNameStr := ToString(r.Mod.Memory(), tableName)
			whereConditionStr := ToString(r.Mod.Memory(), whereCondition)

			result, err := Select(r.Store, smartIndexAddress, tableNameStr, whereConditionStr)

			if err != nil {
				*errorMessage = err
				ptr := r.writeString(r.Mod.Memory(), err.Error())

				return uint32(ptr)
			} else {
				ptr := r.writeString(r.Mod.Memory(), result)

				return uint32(ptr)
			}
		}).
		Export("selectItem").
		NewFunctionBuilder().
		WithFunc(func(statement uint32, args uint32) uint32 {
			statementStr := ToString(r.Mod.Memory(), statement)
			argsStr := ToString(r.Mod.Memory(), args)

			var argsArray []string
			json.Unmarshal([]byte(argsStr), &argsArray)

			result, err := SelectNative(r.Store, statementStr, argsArray)
			resultStr, err := json.Marshal(result)

			if err != nil {
				*errorMessage = err
				ptr := r.writeString(r.Mod.Memory(), err.Error())

				return uint32(ptr)
			} else {
				ptr := r.writeString(r.Mod.Memory(), string(resultStr))

				return uint32(ptr)
			}
		}).
		Export("selectNative").
		NewFunctionBuilder().
		WithFunc(func(height int64) uint32 {
			result, err := r.IndexerDbRepo.GetBlockByHeight(height)
			if err != nil {
				panic(err)
			}
			serializedResult, _ := json.Marshal(result)

			ptr := r.writeString(r.Mod.Memory(), string(serializedResult))

			return uint32(ptr)
		}).
		Export("getBlockByHeight").
		NewFunctionBuilder().
		WithFunc(func(blockHash uint32) uint32 {
			blockHashStr := ToString(r.Mod.Memory(), blockHash)
			result, err := r.IndexerDbRepo.GetTransactionsByBlockHash(blockHashStr)
			if err != nil {
				panic(err)
			}

			serializedResult, _ := json.Marshal(result)

			ptr := r.writeString(r.Mod.Memory(), string(serializedResult))

			return uint32(ptr)
		}).
		Export("getTransactionsByBlockHash").
		NewFunctionBuilder().
		WithFunc(func(transactionHash uint32) uint32 {
			transactionHashStr := ToString(r.Mod.Memory(), transactionHash)
			result, err := r.IndexerDbRepo.GetOutpointsByTransactionHash(transactionHashStr)
			if err != nil {
				panic(err)
			}
			serializedResult, _ := json.Marshal(result)

			ptr := r.writeString(r.Mod.Memory(), string(serializedResult))

			return uint32(ptr)
		}).
		Export("getOutpointsByTransactionHash").
		NewFunctionBuilder().
		WithFunc(func(height uint64) uint32 {
			result, err := r.IndexerDbRepo.GetTransactionV1sByBlockHeight(height)
			if err != nil {
				panic(err)
			}
			serializedResult, _ := json.Marshal(result)

			ptr := r.writeString(r.Mod.Memory(), string(serializedResult))

			return uint32(ptr)
		}).
		Export("getTransactionV1sByBlockHeight").
		NewFunctionBuilder().
		WithFunc(func(strPtr uint32) {
			str := ToString(r.Mod.Memory(), strPtr)

			*output = str
		}).
		Export("valueReturn").
		NewFunctionBuilder().
		WithFunc(func() uint32 {
			ptr := r.writeString(r.Mod.Memory(), string(smartIndexAddress))

			return uint32(ptr)
		}).
		Export("contractAddress").
		NewFunctionBuilder().
		WithFunc(func(strPtr uint32) {
			str := ToString(r.Mod.Memory(), strPtr)

			log.Panicln(str)
		}).
		Export("panic").
		NewFunctionBuilder().
		WithFunc(func(strPtr uint32) {
			str := ToString(r.Mod.Memory(), strPtr)

			fmt.Println("consoleLog", str)
		}).
		Export("consoleLog").
		NewFunctionBuilder().
		WithFunc(func() uint32 {
			network := os.Getenv("NETWORK")
			if network == "" {
				network = "regtest"
			}

			ptr := r.writeString(r.Mod.Memory(), network)

			return uint32(ptr)
		}).
		Export("getNetwork")

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

func (r *WasmRuntime) RunWasmFunction(signer Address, wasmBytes []byte, smartIndexAddress string, functionName string, args []string, kind types.ActionKind) (any, error) {
	ctx := context.Background()
	defer ctx.Done()

	var output string
	var errorMessage error
	mod := r.loadWasm(wasmBytes, ctx, smartIndexAddress, signer, kind, &output, &errorMessage)
	f := mod.ExportedFunction(functionName)

	if f == nil {
		return "", fmt.Errorf("Function not found")
	}

	// All arguments are stringified pointers
	argsPtr := make([]uint64, len(args))
	for i := range argsPtr {
		ptr := r.writeString(r.Mod.Memory(), args[i])
		argsPtr[i] = uint64(ptr)
	}

	_, err := f.Call(ctx, argsPtr...)

	if err != nil {
		return "", err
	}

	if errorMessage != nil {
		return output, nil
	}

	return output, nil
}

func (r *WasmRuntime) RunSelectFunction(statement string, args []string) (any, error) {
	return r.Store.SelectNative(statement, args)
}
