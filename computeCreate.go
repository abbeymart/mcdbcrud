// @Author: abbeymart | Abi Akindele | @Created: 2020-12-08 | @Updated: 2020-12-08
// @Company: mConnect.biz | @License: MIT
// @Description: compute create-SQL script, for bulk/copy insert operation | updated field-type

package mcdbcrud

import (
	"encoding/json"
	"fmt"
	"github.com/asaskevich/govalidator"
	"time"
)

func errMessage(errMsg string) CreateQueryResult {
	return CreateQueryResult{
		CreateQueryObject: CreateQueryObject{
			CreateQuery: "",
			FieldNames:  nil,
			FieldValues: nil,
		},
		Ok:      false,
		Message: errMsg,
	}
}

// ComputeCreateQuery function computes insert SQL scripts. It returns createScripts []string and err error
func ComputeCreateQuery(tableName string, actionParams ActionParamsType) CreateQueryResult {
	if tableName == "" || len(actionParams) < 1 {
		return errMessage("table-name is required for the create operation")
	}

	// declare slice variable for create/insert queries
	var createQuery string
	var fieldNames []string
	var fieldNamesUnderscore []string
	var fieldValues [][]interface{}

	// compute create script and associated values () for all the records in actionParams
	// compute create-query from the first actionParams
	itemQuery := fmt.Sprintf("INSERT INTO %v(", tableName)
	itemValuePlaceholder := " VALUES("
	fieldsLength := len(actionParams[0])
	fieldCount := 0
	for fieldName := range actionParams[0] {
		fieldCount += 1
		fieldNameUnderScore := govalidator.CamelCaseToUnderscore(fieldName)
		fieldNames = append(fieldNames, fieldName)
		fieldNamesUnderscore = append(fieldNamesUnderscore, fieldNameUnderScore)
		itemQuery += fmt.Sprintf("%v", fieldNameUnderScore)
		itemValuePlaceholder += fmt.Sprintf("$%v", fieldCount)
		if fieldsLength > 1 && fieldCount < fieldsLength {
			itemQuery += ", "
			itemValuePlaceholder += ", "
		}
	}
	// close item-script/value-placeholder
	itemQuery += ")"
	itemValuePlaceholder += ")"
	// add/append item-script & value-placeholder to the createScript
	createQuery = itemQuery + itemValuePlaceholder
	createQuery += " RETURNING id"
	// compute create-record-values from actionParams/records, in order of the fields-sequence
	// value-computation for each of the actionParams / records must match the record-fields
	for recIndex, rec := range actionParams {
		// item-values-computation variable
		var recFieldValues []interface{}
		for _, fieldName := range fieldNames {
			fieldValue, ok := rec[fieldName]
			// check for required field in each record
			if !ok {
				return errMessage(fmt.Sprintf("Record #%v [%#v]: required field_name[%v] has field_value of %v ", recIndex, rec, fieldName, fieldValue))
			}
			// update recFieldValues by fieldValue-type, for correct postgres-SQL-parsing
			var currentFieldValue interface{}
			switch fieldValue.(type) {
			case time.Time:
				if fVal, ok := fieldValue.(time.Time); !ok {
					return errMessage(fmt.Sprintf("field_name: %v [date-type] | field_value: %v error: ", fieldName, fieldValue))
				} else {
					currentFieldValue = "'" + fVal.Format("2006-01-02 15:04:05.000000") + "'"
				}
			case map[string]interface{}:
				if fVal, ok := fieldValue.(map[string]interface{}); !ok {
					return errMessage(fmt.Sprintf("field_name: %v [date-type] | field_value: %v error: ", fieldName, fieldValue))
				} else {
					itemValue, _ := json.Marshal(fVal)
					currentFieldValue = string(itemValue)
				}
			case string:
				if fVal, ok := fieldValue.(string); !ok {
					return errMessage(fmt.Sprintf("field_name: %v | field_value: %v error: ", fieldName, fieldValue))
				} else {
					if govalidator.IsUUID(fVal) {
						currentFieldValue = fVal
					} else if govalidator.IsJSON(fVal) {
						if fValue, err := govalidator.ToJSON(fieldValue); err != nil {
							return errMessage(fmt.Sprintf("field_name: %v | field_value: %v error: ", fieldName, fieldValue))
						} else {
							currentFieldValue = fValue
						}
					} else {
						currentFieldValue = fVal
					}
				}
			default:
				currentFieldValue = fieldValue
			}
			// add itemValue
			recFieldValues = append(recFieldValues, currentFieldValue)
		}
		// update fieldValues
		fieldValues = append(fieldValues, recFieldValues)
		// re-initialise recFieldValues, for next update
		recFieldValues = []interface{}{}
	}

	// result
	return CreateQueryResult{
		CreateQueryObject: CreateQueryObject{
			CreateQuery: createQuery,
			FieldNames:  fieldNamesUnderscore,
			FieldValues: fieldValues,
		},
		Ok:      true,
		Message: "success",
	}
}
