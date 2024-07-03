package utils

import (
	"eastnode/indexer/model"
	"eastnode/utils"
	"os"

	_ "github.com/dolthub/driver"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	// "gorm.io/gorm/logger"
)

func (s *Store) InitIndexerDb() {
	if _, err := os.Stat(utils.Cwd() + "/db/indexer"); os.IsNotExist(err) {
		_, err = s.Instance.Exec("CREATE DATABASE indexer")
		if err != nil {
			panic(err)
		}
	}

	db, err := model.NewDB(mysql.New(mysql.Config{
		DriverName: "dolt",
		DSN:        "file://" + utils.Cwd() + "/db?commitname=root&commitemail=root@east&multistatements=true&database=indexer",
	}), &gorm.Config{
		// Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		panic(err)
	}

	s.Gorm = db
}
