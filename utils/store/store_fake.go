package utils

import (
	"database/sql"
	indexerDb "eastnode/indexer/repository/db"
	"eastnode/utils"
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func GetFakeInstance(instanceType InstanceType, dumpFile string) *Store {
	if instanceType == SmartIndexDB && sSmartIndex != nil {
		return sSmartIndex
	} else if instanceType == ChainDB && sChain != nil {
		return sChain
	} else {
		lock.Lock()
		defer lock.Unlock()
		os.Mkdir("db_test", 0700)

		doltInstance, err := sql.Open("dolt", "file://"+utils.Cwd()+"/db_test?commitname=root&commitemail=root@east&multistatements=true")

		if err != nil {
			log.Panicln(err)
		}

		if instanceType == SmartIndexDB {
			sSmartIndex = &Store{Instance: doltInstance}

			if _, err := os.Stat(utils.Cwd() + "/db_test/indexer"); os.IsNotExist(err) {
				dump, _ := os.ReadFile(dumpFile)

				_, err = sSmartIndex.Instance.Exec(string(dump))
				if err != nil {
					panic(err)
				}
			}

			db, err := indexerDb.NewDB(mysql.New(mysql.Config{
				DriverName: "dolt",
				DSN:        "file://" + utils.Cwd() + "/db_test?commitname=root&commitemail=root@east&multistatements=true&database=indexer",
			}), &gorm.Config{})
			if err != nil {
				panic(err)
			}

			sSmartIndex.Gorm = db
			sSmartIndex.InitWasmDB()
			return sSmartIndex
		} else if instanceType == ChainDB {
			sChain = &Store{Instance: doltInstance}
			sChain.InitChainDb("/db_test")
			return sChain
		} else {
			sIndexer = &Store{Instance: doltInstance}
			sIndexer.InitIndexerDb("/db_test")
			return sIndexer
		}
	}
}
