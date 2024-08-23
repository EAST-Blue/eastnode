package utils

import (
	"eastnode/indexer/repository/db"

	_ "github.com/dolthub/driver"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func (s *Store) InitIndexerDb() {
	_, err := s.Instance.Exec("CREATE DATABASE IF NOT EXISTS indexer")
	if err != nil {
		panic(err)
	}

	indexerDb, err := db.NewDB(mysql.New(mysql.Config{
		DriverName: "mysql",
		DSN:        "user:password@tcp(127.0.0.1)/db",
	}), &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Error),
		PrepareStmt:            true,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		panic(err)
	}

	s.Gorm = indexerDb
}
