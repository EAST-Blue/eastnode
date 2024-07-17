package utils

import (
	"database/sql"
	"eastnode/utils"
	"log"
	"os"
	"sync"

	"github.com/uptrace/bun"
	"go.etcd.io/bbolt"
	"gorm.io/gorm"
)

type Store struct {
	KV          *bbolt.DB
	Instance    *sql.DB
	BunInstance *bun.DB
	Gorm        *gorm.DB
}

var lock = &sync.Mutex{}

var sChain *Store
var sSmartIndex *Store
var sIndexer *Store

type InstanceType int

const (
	SmartIndexDB InstanceType = 0
	ChainDB      InstanceType = 1
	IndexerDB    InstanceType = 2
)

func GetInstance(instanceType InstanceType) *Store {

	if instanceType == SmartIndexDB && sSmartIndex != nil {
		return sSmartIndex
	} else if instanceType == ChainDB && sChain != nil {
		return sChain
	} else if instanceType == IndexerDB && sIndexer != nil {
		return sIndexer
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
		} else if instanceType == ChainDB {
			sChain = &Store{Instance: doltInstance}
			sChain.InitChainDb()
			return sChain
		} else {
			sIndexer = &Store{Instance: doltInstance}
			sIndexer.InitIndexerDb()
			return sIndexer
		}
	}
}
