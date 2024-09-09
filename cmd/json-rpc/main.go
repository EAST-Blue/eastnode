package main

import (
	"eastnode/chain"
	"eastnode/indexer/repository/bitcoin"
	indexerDb "eastnode/indexer/repository/db"
	"eastnode/jsonrpc"
	store "eastnode/utils/store"
	"log"
	"net/http"
	"os"

	_ "github.com/dolthub/driver"
	"github.com/gorilla/mux"
	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	log.Println("Initializing...")

	indexerDbInstance := store.GetInstance(store.IndexerDB)
	indexerDbRepo := indexerDb.NewDBRepository(indexerDbInstance.Gorm)

	blockchain := new(chain.Chain)
	bc := blockchain.Init(indexerDbRepo)

	rpcServer := rpc.NewServer()

	rpcServer.RegisterCodec(json.NewCodec(), "application/json")
	rpcServer.RegisterCodec(json.NewCodec(), "application/json;charset=UTF-8")

	runtimeServer := &jsonrpc.RuntimeServer{
		Chain: bc,
	}
	commonServer := &jsonrpc.CommonServer{
		Chain: bc,
	}
	btcServer := &jsonrpc.BitcoinServer{
		BitcoinRepo: bitcoin.NewBitcoinRepo(os.Getenv("BTC_RPC_URL"), "east", "east"),
	}

	rpcServer.RegisterService(runtimeServer, "Runtime")
	rpcServer.RegisterService(commonServer, "Common")

	router := mux.NewRouter()
	router.Handle("/", rpcServer)
	router.Handle("/btc", btcServer)

	log.Println("rpc is running")
	http.ListenAndServe(":4000", router)
}
