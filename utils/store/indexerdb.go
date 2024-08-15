package utils

import (
	"eastnode/indexer/repository/db"
	"eastnode/utils"
	"os"

	_ "github.com/dolthub/driver"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func (s *Store) InitIndexerDb() {
	if _, err := os.Stat(utils.Cwd() + "/db/indexer"); os.IsNotExist(err) {
		_, err = s.Instance.Exec("CREATE DATABASE indexer")
		if err != nil {
			panic(err)
		}
	}

	indexerDb, err := db.NewDB(mysql.New(mysql.Config{
		DriverName: "dolt",
		DSN:        "file://" + utils.Cwd() + "/db?commitname=root&commitemail=root@east&multistatements=true&database=indexer",
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
