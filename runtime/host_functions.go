package runtime

import (
	store "eastnode/utils/store"
	"encoding/json"
	"fmt"
	"log"

	"github.com/iancoleman/strcase"
	dynamicstruct "github.com/ompluscator/dynamic-struct"
)

var ContractTableSeparator string = "_"

type TableOption struct {
	PrimaryKey string
	Indexes    []string
}

func CreateTable(s store.Store, contractAddress string, tableName string, schema string, option string) {
	var ts map[string]interface{}
	if err := json.Unmarshal([]byte(schema), &ts); err != nil {
		log.Panicln(err)
	}

	var opt TableOption
	if err := json.Unmarshal([]byte(option), &opt); err != nil {
		log.Panicln(err)
	}

	// TODO: Validate schema input, e.g. table_schema keys must be exported
	instance := dynamicstruct.NewStruct()

	for k, v := range ts {
		var vType interface{}

		if v == "uint" || v == "int" {
			vType = (*int)(nil)
		} else {
			vType = (*string)(nil)
		}

		if opt.PrimaryKey == k {
			instance.AddField(strcase.ToCamel(k), vType, `bun:",pk"`)
		} else {
			instance.AddField(strcase.ToCamel(k), vType, `bun:"type:mediumtext"`)
		}
	}

	newInstance := instance.Build().New()

	s.CreateTable(newInstance, getStateTableName(contractAddress, tableName), opt.Indexes)
}

func Insert(s store.Store, contractAddress string, tableName string, values string) {
	var ts map[string]interface{}
	if err := json.Unmarshal([]byte(values), &ts); err != nil {
		log.Panicln(err)
	}

	s.Insert(&ts, getStateTableName(contractAddress, tableName))
}

func Update(s store.Store, contractAddress string, tableName string, whereCondition string, values string) {
	var valuesMap map[string]interface{}
	var whereConditionMap map[string]interface{}

	if err := json.Unmarshal([]byte(whereCondition), &whereConditionMap); err != nil {
		log.Panicln(err)
	}

	if err := json.Unmarshal([]byte(values), &valuesMap); err != nil {
		log.Panicln(err)
	}

	s.Update(&valuesMap, getStateTableName(contractAddress, tableName), whereConditionMap)
}

func Delete(s store.Store, contractAddress string, tableName string, whereCondition string) {
	var whereConditionMap map[string]interface{}

	if err := json.Unmarshal([]byte(whereCondition), &whereConditionMap); err != nil {
		log.Panicln(err)
	}

	s.Delete(getStateTableName(contractAddress, tableName), whereConditionMap)
}

func Select(s store.Store, contractAddress string, tableName string, whereCondition string) (string, error) {
	var whereConditionMap map[string]interface{}

	if err := json.Unmarshal([]byte(whereCondition), &whereConditionMap); err != nil {
		log.Panicln(err)
	}

	result, err := s.Select(getStateTableName(contractAddress, tableName), whereConditionMap)

	if err != nil {
		return "", fmt.Errorf(`{"error":"%s"}`, err)
	}

	resultMarshalled, err := json.Marshal(result)
	if err != nil {
		log.Panicln(err)
	}

	return string(resultMarshalled), nil
}

func SelectNative(s store.Store, statement string, args []string) (any, error) {
	return s.SelectNative(statement, args)
}

func getStateTableName(contractAddress string, tableName string) string {
	return fmt.Sprintf("%s%s%s", contractAddress, ContractTableSeparator, tableName)
}
