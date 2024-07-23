package jsonrpc

import (
	"eastnode/chain"
	"eastnode/types"
	"eastnode/utils"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
)

type RuntimeServer struct {
	Chain *chain.Chain
}

func (s *RuntimeServer) Mutate(r *http.Request, params *string, reply *types.RpcReply) error {
	log.Printf("Running Call Function")

	blockHeight := s.Chain.GetBlockHeight()
	blockHash := s.Chain.GetBlockHash(blockHeight)

	newSignedTx := new(types.SignedTransaction)
	utils.DecodeHexAndBorshDeserialize(newSignedTx, *params)

	err := s.Chain.CheckTx(*newSignedTx)

	if err == nil {

		// add to mempool, signal to produce new block
		log.Println("adding to mempool")

		s.Chain.Mempool.Enqueue(*newSignedTx)

		s.Chain.ProduceBlock()

		*reply = types.RpcReply{
			BlockHash:   blockHash,
			BlockHeight: blockHeight,
			Result:      []byte(newSignedTx.ID),
		}
	} else {
		*reply = types.RpcReply{
			BlockHash:   blockHash,
			BlockHeight: blockHeight,
			Result:      []byte(err.Error()),
		}
	}

	return nil
}

func (s *RuntimeServer) Query(r *http.Request, params *types.RuntimeServerQuery, reply *types.ServerQueryReply) error {
	log.Printf("Running Query Function")

	blockHeight := s.Chain.GetBlockHeight()
	blockHash := s.Chain.GetBlockHash(blockHeight)

	if params.FunctionName == "get_smart_index_wasm" {
		smartIndexAddress := params.Args[0]

		smartIndexWasm := s.Chain.GetSmartIndexWasm(smartIndexAddress)

		*reply = types.ServerQueryReply{
			BlockHash:   blockHash,
			BlockHeight: blockHeight,
			Result:      hex.EncodeToString(smartIndexWasm),
		}

	} else if params.FunctionName == "view_function" {
		smartIndexAddress := params.Target
		functionName := params.Args[0]
		res, err := s.Chain.ProcessWasmCall("", smartIndexAddress, functionName, params.Args[1:], types.View)

		if err != nil {
			*reply = types.ServerQueryReply{
				BlockHash:   blockHash,
				BlockHeight: blockHeight,
				Result:      hex.EncodeToString([]byte(err.Error())),
			}

		} else {
			result, _ := json.Marshal(res)

			*reply = types.ServerQueryReply{
				BlockHash:   blockHash,
				BlockHeight: blockHeight,
				Result:      hex.EncodeToString(result),
			}
		}

	} else if params.FunctionName == "get_transaction" {
		txId := params.Args[0]

		result, _ := json.Marshal(s.Chain.GetTransaction(txId))

		*reply = types.ServerQueryReply{
			BlockHash:   blockHash,
			BlockHeight: blockHeight,
			Result:      hex.EncodeToString(result),
		}
	} else if params.FunctionName == "select_native_sql" {
		res, err := s.Chain.WasmRuntime.RunSelectFunction(params.Args[0], params.Args[1:])

		if err != nil {
			*reply = types.ServerQueryReply{
				BlockHash:   blockHash,
				BlockHeight: blockHeight,
				Result:      hex.EncodeToString([]byte(err.Error())),
			}

		} else {
			result, _ := json.Marshal(res)

			*reply = types.ServerQueryReply{
				BlockHash:   blockHash,
				BlockHeight: blockHeight,
				Result:      hex.EncodeToString(result),
			}
		}

	}

	return nil
}
