package utils

import (
	"database/sql"
	"eastnode/utils"
	"log"
	"os"
	"sync"

	"github.com/uptrace/bun"
	"go.etcd.io/bbolt"
)

type Store struct {
	KV          *bbolt.DB
	Instance    *sql.DB
	BunInstance *bun.DB
}

var lock = &sync.Mutex{}

var sChain *Store
var sSmartIndex *Store

type InstanceType int

const (
	SmartIndexDB InstanceType = 0
	ChainDB      InstanceType = 1
)

func GetInstance(instanceType InstanceType) *Store {

	if instanceType == SmartIndexDB && sSmartIndex != nil {
		return sSmartIndex
	} else if instanceType == ChainDB && sChain != nil {
		return sChain
	} else {
		lock.Lock()
		defer lock.Unlock()
		os.Mkdir("db", 0700)

		doltInstance, err := sql.Open("dolt", "file://"+utils.Cwd()+"/db?commitname=root&commitemail=root@east&multistatements=true")

		if err != nil {
			log.Panicln(err)
		}

		if instanceType == SmartIndexDB {
			sSmartIndex = &Store{Instance: doltInstance}
			sSmartIndex.InitWasmDB()
			return sSmartIndex
		} else {
			sChain = &Store{Instance: doltInstance}
			sChain.InitChainDb()
			return sChain
		}
	}
}
