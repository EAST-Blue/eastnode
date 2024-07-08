package utils

import (
	"context"
	"fmt"
	"log"

	_ "github.com/dolthub/driver"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
	"github.com/uptrace/bun/extra/bundebug"
)

func (s *Store) CreateTable(model interface{}, tableName string) {
	ctx := context.Background()

	if res, err := s.
		BunInstance.
		NewCreateTable().
		Model(model).
		ModelTableExpr(tableName).
		Exec(ctx); err != nil {
		log.Panicln(err)
	} else {
		fmt.Println("[+] Create table: ", res)
	}
}

func (s *Store) Insert(model interface{}, tableName string) {
	ctx := context.Background()

	if res, err := s.
		BunInstance.
		NewInsert().
		Model(model).
		ModelTableExpr(tableName).
		Exec(ctx); err != nil {
		log.Panicln(err)
	} else {
		fmt.Println("[+] Insert : ", res)
	}
}

func (s *Store) Update(model interface{}, tableName string, whereCondition map[string]interface{}) {
	ctx := context.Background()

	whereConditionStr := ""
	for k, v := range whereCondition {
		if len(whereConditionStr) > 1 {
			whereConditionStr = whereConditionStr + " and"
		}
		whereConditionStr = whereConditionStr + fmt.Sprintf("%s = \"%s\"", k, v)
	}

	if res, err := s.
		BunInstance.
		NewUpdate().
		Model(model).
		ModelTableExpr(tableName).
		Where(whereConditionStr).
		Exec(ctx); err != nil {
		log.Panicln(err)
	} else {
		fmt.Println("[+] Update : ", res)
	}
}

func (s *Store) Delete(tableName string, whereCondition map[string]interface{}) {
	ctx := context.Background()

	whereConditionStr := ""
	for k, v := range whereCondition {
		if len(whereConditionStr) > 1 {
			whereConditionStr = whereConditionStr + " and"
		}
		whereConditionStr = whereConditionStr + fmt.Sprintf("%s = \"%s\"", k, v)
	}

	if res, err := s.
		BunInstance.
		NewDelete().
		ModelTableExpr(tableName).
		Where(whereConditionStr).
		Exec(ctx); err != nil {
		log.Panicln(err)
	} else {
		fmt.Println("[+] Delete : ", res)
	}
}

func (s *Store) Select(tableName string, whereCondition map[string]interface{}) interface{} {
	ctx := context.Background()

	whereConditionStr := ""
	for k, v := range whereCondition {
		if len(whereConditionStr) > 1 {
			whereConditionStr = whereConditionStr + " and"
		}
		whereConditionStr = whereConditionStr + fmt.Sprintf("%s = \"%s\"", k, v)
	}

	var result map[string]interface{}

	if count, err := s.
		BunInstance.
		NewSelect().
		Model(&result).
		ModelTableExpr(tableName).
		Where(whereConditionStr).
		Limit(1).
		ScanAndCount(ctx); err != nil {
		return "not-found"
	} else {
		fmt.Println("[+] Select : ", count, result)
	}
	return result
}

func (s *Store) ShowTables(noPrint bool) {
	res, err := s.Instance.Query("show tables")

	if err != nil {
		log.Panicln(err)
	}

	var str string
	res.Scan(&str)

	for res.Next() {
		res.Scan(&str)
		if !noPrint {
			fmt.Println(str)

		}
	}

}

func (s *Store) InitWasmDB() {
	s.Instance.Exec("CREATE DATABASE states")
	s.Instance.Exec("Use states")

	bundb := bun.NewDB(s.Instance, mysqldialect.New())
	bundb.AddQueryHook(bundebug.NewQueryHook(
		bundebug.WithVerbose(true),
		bundebug.FromEnv("BUNDEBUG"),
	))

	s.BunInstance = bundb
}
