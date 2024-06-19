package runtime

import (
	store "eastnode/utils/store"
	"encoding/json"
	"fmt"
	"log"

	"github.com/iancoleman/strcase"
	dynamicstruct "github.com/ompluscator/dynamic-struct"
)

// Host function to DB
// CreateTable
// Insert
// Select
// Update

var ContractTableSeparator string = "_"

func CreateTable(s store.Store, contractAddress string, tableName string, primaryKey string, schema string) {
	var ts map[string]interface{}
	if err := json.Unmarshal([]byte(schema), &ts); err != nil {
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

		if primaryKey == k {
			instance.AddField(strcase.ToCamel(k), vType, `bun:",pk"`)
		} else {
			instance.AddField(strcase.ToCamel(k), vType, "")
		}
	}

	newInstance := instance.Build().New()

	s.CreateTable(newInstance, fmt.Sprintf("%s%s%s", contractAddress, ContractTableSeparator, tableName))
}

func Insert(s store.Store, contractAddress string, tableName string, values string) {
	var ts map[string]interface{}
	if err := json.Unmarshal([]byte(values), &ts); err != nil {
		log.Panicln(err)
	}

	s.Insert(&ts, fmt.Sprintf("%s%s%s", contractAddress, ContractTableSeparator, tableName))
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

	s.Update(&valuesMap, fmt.Sprintf("%s%s%s", contractAddress, ContractTableSeparator, tableName), whereConditionMap)
}

func Delete(s store.Store, contractAddress string, tableName string, whereCondition string) {
	var whereConditionMap map[string]interface{}

	if err := json.Unmarshal([]byte(whereCondition), &whereConditionMap); err != nil {
		log.Panicln(err)
	}

	s.Delete(fmt.Sprintf("%s%s%s", contractAddress, ContractTableSeparator, tableName), whereConditionMap)
}

func Select(s store.Store, contractAddress string, tableName string, whereCondition string) string {
	var whereConditionMap map[string]interface{}

	if err := json.Unmarshal([]byte(whereCondition), &whereConditionMap); err != nil {
		log.Panicln(err)
	}

	result := s.Select(fmt.Sprintf("%s%s%s", contractAddress, ContractTableSeparator, tableName), whereConditionMap)

	resultMarshalled, err := json.Marshal(result)
	if err != nil {
		log.Panicln(err)
	}

	return string(resultMarshalled)
}

// Host function to Bitcoin node
// Get block id
// Get transaction id
