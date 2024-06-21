package jsonrpc

import (
	"eastnode/chain"
	"eastnode/types"
	"eastnode/utils"
	"encoding/hex"
	"net/http"

	"github.com/near/borsh-go"
)

type CommonServer struct {
	Chain *chain.Chain
}

func (s *CommonServer) Query(r *http.Request, params *string, reply *types.RpcReply) error {
	blockHeight := s.Chain.GetBlockHeight()
	blockHash := s.Chain.GetBlockHash(blockHeight)

	var decodedParams, _ = hex.DecodeString(*params)

	queryParams := new(types.CommonServerQuery)
	borsh.Deserialize(queryParams, decodedParams)

	if queryParams.FunctionName == "get_nonce" {
		pubKey := queryParams.Args[0]
		nonce := s.Chain.GetNonce(pubKey)

		*reply = types.RpcReply{
			BlockHash:   blockHash,
			BlockHeight: blockHeight,
			Result:      utils.Itob(nonce),
		}
	}

	return nil
}
