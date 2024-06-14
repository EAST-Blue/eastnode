package main

import (
	"eastnode/chain"
	"log"
	"net/http"

	_ "github.com/dolthub/driver"
	"github.com/gorilla/mux"
	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"
)

func main() {
	blockchain := new(chain.Chain)
	bc := blockchain.Init()

	rpcServer := rpc.NewServer()

	rpcServer.RegisterCodec(json.NewCodec(), "application/json")
	rpcServer.RegisterCodec(json.NewCodec(), "application/json;charset=UTF-8")

	runtimeServer := &chain.Jsonrpc{
		Chain: bc,
	}

	rpcServer.RegisterService(runtimeServer, "runtime")

	router := mux.NewRouter()
	router.Handle("/", rpcServer)

	log.Println("rpc is running")
	http.ListenAndServe(":3000", router)
}
