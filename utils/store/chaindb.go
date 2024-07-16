package utils

import (
	"eastnode/utils"
	"fmt"
	"log"
	"os"

	_ "github.com/dolthub/driver"
	bolt "go.etcd.io/bbolt"
)

func (s *Store) InitChainDb() {
	// init core schema
	if _, err := os.Stat(utils.Cwd() + "/db/core"); os.IsNotExist(err) {
		fmt.Println("Runtime is initializing...")
		_, err := s.Instance.Exec(`
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
				actions BLOB,
				created_at BIGINT,
				primary key(id)
			);
			CREATE TABLE smart_index (
				smart_index_address VARCHAR(255),
				owner_address VARCHAR(255),
				wasm_blob BLOB,
				primary key(smart_index_address)
			);
			CREATE TABLE transaction_logs (
				id VARCHAR(255),
				statuses JSON,
				logs JSON
			);
			CALL DOLT_COMMIT('-Am', 'init core schema');
		`)

		if err != nil {
			panic(err)
		}
	} else {
		s.Instance.Exec("USE core")
		_, err = s.Instance.Exec("CALL DOLT_RESET('--hard')")

		if err != nil {
			panic(err)
		}
	}

	log.Println("[+] Chain database instance is running")

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

	s.KV = kv
}

func (s *Store) Close() error {
	if err := s.Instance.Close(); err != nil {
		return err
	}
	if err := s.KV.Close(); err != nil {
		return err
	}
	return nil
}
