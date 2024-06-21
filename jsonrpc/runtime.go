package jsonrpc

import (
	"eastnode/chain"
	"eastnode/types"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"

	"github.com/near/borsh-go"
)

type RuntimeServer struct {
	Chain *chain.Chain
}

func (s *RuntimeServer) Mutate(r *http.Request, params *string, reply *types.RpcReply) error {
	log.Printf("Running Call Function")

	blockHeight := s.Chain.GetBlockHeight()
	blockHash := s.Chain.GetBlockHash(blockHeight)

	var decodedParams, _ = hex.DecodeString(*params)

	newSignedTx := new(types.SignedTransaction)
	borsh.Deserialize(newSignedTx, decodedParams)

	err := s.Chain.CheckTx(*newSignedTx)

	if err == nil {
		// add to mempool, signal to produce new block
		log.Println("adding to mempool")

		s.Chain.Mempool.Enqueue(*newSignedTx)

		s.Chain.ProduceBlock()

		*reply = types.RpcReply{
			BlockHash:   blockHash,
			BlockHeight: blockHeight,
			Result:      []byte("true"),
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

	result, err := json.Marshal(params)

	if err != nil {
		panic(err)
	}

	*reply = types.ServerQueryReply{
		BlockHash:   blockHash,
		BlockHeight: blockHeight,
		Result:      string(result[:]),
	}

	return nil
}
