package chain

import (
	"database/sql"
	"eastnode/utils"
	"fmt"
	"log"
	"os"

	_ "github.com/dolthub/driver"
	bolt "go.etcd.io/bbolt"
)

type Store struct {
	KV     *bolt.DB
	Engine *sql.DB
}

func (s *Store) Init() {
	engine, err := sql.Open("dolt", "file://"+utils.Cwd()+"/db?commitname=root&commitemail=root@east&multistatements=true")

	if err != nil {
		panic(err)
	}

	// init core schema
	if _, err := os.Stat(utils.Cwd() + "/db/core"); os.IsNotExist(err) {
		fmt.Println("Runtime is initializing...")
		_, err := engine.Exec(`
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
			CALL DOLT_COMMIT('-Am', 'init core schema');
		`)

		if err != nil {
			panic(err)
		}
	} else {
		engine.Exec("USE core")
		_, err = engine.Exec("CALL DOLT_RESET('--hard')")

		if err != nil {
			panic(err)
		}
	}

	log.Println("storage engine is running")

	kv, err := bolt.Open("db/chain.db", 0600, nil)
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
	log.Println("storage kv is running")

	s.Engine = engine
	s.KV = kv
}

func (s *Store) Close() error {
	if err := s.Engine.Close(); err != nil {
		return err
	}
	if err := s.KV.Close(); err != nil {
		return err
	}
	return nil
}
