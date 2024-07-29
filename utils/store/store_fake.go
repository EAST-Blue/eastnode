package utils

import (
	"database/sql"
	indexerDb "eastnode/indexer/repository/db"
	"eastnode/utils"
	"log"
	"os"

	bolt "go.etcd.io/bbolt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func GetFakeInstance(instanceType InstanceType) *Store {
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
				dump, _ := os.ReadFile("../utils/store/test/doltdump.sql")

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
		} else {
			sChain = &Store{Instance: doltInstance}

			sChain.Instance.Exec(`
				CREATE DATABASE core;
				USE core;
				CREATE TABLE kv (
					k varchar(255),
					v varchar(255),
					primary key(k)
				);
				CREATE TABLE transactions (
					id VARCHAR(255),
					block_id BIGINT,
					signer VARCHAR(255),
					receiver VARCHAR(255),
					actions VARBINARY(1024),
					created_at BIGINT,
					primary key(id)
				);
				CREATE TABLE smart_index (
					smart_index_address VARCHAR(255),
					owner_address VARCHAR(255),
					wasm_blob LONGBLOB,
					primary key(smart_index_address)
				);
				CREATE TABLE transaction_logs (
					id VARCHAR(255),
					statuses JSON,
					logs JSON,
					primary key(id)
				);
				CALL DOLT_COMMIT('-Am', 'init core schema');
			`)

			kv, err := bolt.Open("db_test/chain.db", 0600, nil)
			if err != nil {
				panic(err)
			}

			// init key/value store
			err = kv.Update(func(tx *bolt.Tx) error {
				_, err := tx.CreateBucketIfNotExists([]byte("blocks"))

				if err != nil {
					return err
				}

				_, err = tx.CreateBucketIfNotExists([]byte("common"))

				if err != nil {
					return err
				}

				return nil
			})
			if err != nil {
				panic(err)
			}

			sChain.KV = kv
			return sChain
		}
	}
}
