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

func (s *Store) CreateTable(model interface{}, tableName string, indexes []string) {
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

	for _, column := range indexes {
		if res, err := s.
			BunInstance.
			NewCreateIndex().
			Model(model).
			ModelTableExpr(tableName).
			Index(fmt.Sprintf("%s_idx", column)).
			Column(column).
			Exec(ctx); err != nil {
			log.Panicln(err)
		} else {
			fmt.Println("[+] Create index: ", res)
		}
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
			whereConditionStr = whereConditionStr + " and "
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
			whereConditionStr = whereConditionStr + " and "
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

func (s *Store) Select(tableName string, whereCondition map[string]interface{}) (interface{}, error) {
	ctx := context.Background()

	whereConditionStr := ""
	for k, v := range whereCondition {
		if len(whereConditionStr) > 1 {
			whereConditionStr = whereConditionStr + " and "
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
		return nil, err
	} else {
		fmt.Println("[+] Select : ", count, result)
	}
	return result, nil
}

func (s *Store) SelectNative(statement string, args []string) (interface{}, error) {
	argsInput := make([]interface{}, len(args))
	for i := range args {
		argsInput[i] = args[i]
	}

	rows, err := s.BunInstance.Query(statement, argsInput...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		fmt.Println("Failed to get columns", err)
		return nil, err
	}

	dest := make([]interface{}, len(cols))
	rawResult := make([][]byte, len(cols))
	result := []map[string]interface{}{}

	for i, _ := range rawResult {
		dest[i] = &rawResult[i] // Put pointers to each string in the interface slice
	}

	for rows.Next() {
		err = rows.Scan(dest...)
		if err != nil {
			fmt.Println("Failed to scan row", err)
			return nil, err
		}

		resultTemp := make(map[string]interface{})
		for i, raw := range rawResult {
			if raw == nil {
				resultTemp[cols[i]] = "\\N"
			} else {
				resultTemp[cols[i]] = string(raw)
			}
		}

		result = append(result, resultTemp)
	}
	return result, nil
}

func (s *Store) InitWasmDB() {
	s.Instance.Exec("CREATE DATABASE states")
	s.Instance.Exec("USE states")

	bundb := bun.NewDB(s.Instance, mysqldialect.New())
	bundb.AddQueryHook(bundebug.NewQueryHook(
		bundebug.WithVerbose(true),
		bundebug.FromEnv("BUNDEBUG"),
	))

	s.BunInstance = bundb
}
