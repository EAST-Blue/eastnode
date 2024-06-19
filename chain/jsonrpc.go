package chain

import (
	"eastnode/types"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"

	"github.com/near/borsh-go"
)

type Jsonrpc struct {
	Chain *Chain
}

func (s *Jsonrpc) Mutate(r *http.Request, params *string, reply *types.RpcReply) error {
	log.Printf("Running Call Function")

	blockHeight := s.Chain.GetBlockHeight()
	blockHash := s.Chain.GetBlockHash(blockHeight)

	var decodedParams, _ = hex.DecodeString(*params)

	newSignedTx := new(types.SignedTransaction)
	borsh.Deserialize(newSignedTx, decodedParams)

	txIsValid := newSignedTx.IsValid()

	if txIsValid {
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
			Result:      []byte("false"),
		}
	}

	return nil
}

func (s *Jsonrpc) Query(r *http.Request, params *string, reply *types.RpcReply) error {
	log.Printf("Running Query Function")

	blockHeight := s.Chain.GetBlockHeight()
	blockHash := s.Chain.GetBlockHash(blockHeight)

	result, err := json.Marshal(params)

	if err != nil {
		panic(err)
	}

	*reply = types.RpcReply{
		BlockHash:   blockHash,
		BlockHeight: blockHeight,
		Result:      result,
	}

	return nil
}
