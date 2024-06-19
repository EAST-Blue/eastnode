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

var s *Store

func GetInstance() *Store {
	if s == nil {
		lock.Lock()
		defer lock.Unlock()
		if s == nil {
			os.Mkdir("db", 0700)

			doltInstance, err := sql.Open("dolt", "file://"+utils.Cwd()+"/db?commitname=root&commitemail=root@east&multistatements=true")

			if err != nil {
				log.Panicln(err)
			}

			s = &Store{Instance: doltInstance}
			s.InitWasmDB()
			s.InitChainDb()
		}
	}

	return s
}
