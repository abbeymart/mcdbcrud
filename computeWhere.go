// @Author: abbeymart | Abi Akindele | @Created: 2020-12-08 | @Updated: 2020-12-08
// @Company: mConnect.biz | @License: MIT
// @Description: compute where-SQL script | TODO: review/resolve near WHERE error (re: fieldCount)

package mcdbcrud

import (
	"encoding/json"
	"fmt"
	"github.com/asaskevich/govalidator"
	"reflect"
	"time"
)

func whereErrMessage(errMsg string) WhereQueryResult {
	return WhereQueryResult{
		WhereQueryObject: WhereQueryObject{
			WhereQuery:  "",
			FieldValues: nil,
		},
		Ok:      false,
		Message: errMsg,
	}
}

// ComputeWhereQuery function computes the multi-cases where-conditions for crud-operations
func ComputeWhereQuery(queryParams QueryParamType, fieldLength int) WhereQueryResult {
	if len(queryParams) < 1 || fieldLength < 1 {
		return whereErrMessage("queryParams (where-conditions) and fieldLength (starting position for the where-condition-placeholder-values) are required.")
	}
	// compute queryParams script from queryParams
	whereQuery := "WHERE "
	var fieldValues []interface{}
	fieldCount := 0
	whereFieldLength := len(queryParams)
	for fieldName, fieldValue := range queryParams {
		// update fieldValues by fieldValue-type, for correct postgres-SQL-parsing
		var currentFieldValue interface{}
		// validate field-value type
		fieldType := fmt.Sprintf("%v", reflect.TypeOf(fieldValue).Kind())
		switch fieldType {
		case "slice":
			if fVal, ok := fieldValue.([]string); !ok {
				if fVal2, ok2 := fieldValue.([]interface{}); !ok2 {
					return whereErrMessage(fmt.Sprintf("field_name: %v [slice-type] | field_value: %v error: ", fieldName, fieldValue))
				} else {
					// compute IN clause for []interface{} (string and other values)
					idLen := len(fVal2)
					recIds := "("
					for i, val := range fVal2 {
						if valStr, valStrOk := val.(string); valStrOk {
							recIds += fmt.Sprintf("'%v'", valStr)
						} else {
							recIds += fmt.Sprintf("%v", val)
						}
						if idLen > 1 && i < idLen-1 {
							recIds += ", "
						}
					}
					recIds += ")"
					// fieldValues.push(`${recIds}`)
					fieldNameUnderscore := govalidator.CamelCaseToUnderscore(fieldName)
					whereQuery += fmt.Sprintf("%v IN %v", fieldNameUnderscore, recIds)
				}
			} else {
				// compute IN clause for []string
				idLen := len(fVal)
				recIds := "("
				for i, val := range fVal {
					recIds += "'" + val + "'"
					if i < idLen-1 {
						recIds += ", "
					}
				}
				recIds += ")"
				// fieldValues.push(`${recIds}`)
				fieldNameUnderscore := govalidator.CamelCaseToUnderscore(fieldName)
				whereQuery += fmt.Sprintf("%v IN %v", fieldNameUnderscore, recIds)
			}
		default:
			switch fieldValue.(type) {
			case time.Time:
				if fVal, ok := fieldValue.(time.Time); !ok {
					return whereErrMessage(fmt.Sprintf("field_name: %v [date-type] | field_value: %v error: ", fieldName, fieldValue))
				} else {
					currentFieldValue = "'" + fVal.Format("2006-01-02 15:04:05.000000") + "'"
					fieldValues = append(fieldValues, currentFieldValue)
					whereQuery += fmt.Sprintf("%v=$%v", govalidator.CamelCaseToUnderscore(fieldName), fieldLength)
				}
			case string:
				if fVal, ok := fieldValue.(string); !ok {
					return whereErrMessage(fmt.Sprintf("field_name: %v | field_value: %v error: ", fieldName, fieldValue))
				} else {
					if govalidator.IsJSON(fVal) {
						//fmt.Printf("string-toJson-value: %v\n\n", fVal)
						currentFieldValue = fVal
						fieldValues = append(fieldValues, currentFieldValue)
						whereQuery += fmt.Sprintf("%v=$%v", govalidator.CamelCaseToUnderscore(fieldName), fieldLength)
						//if fValue, jErr := govalidator.ToJSON(fieldValue); jErr != nil {
						//	return whereErrMessage(fmt.Sprintf("field_name: %v | field_value: %v error: ", fieldName, fieldValue))
						//} else {
						//}
					} else {
						//currentFieldValue = "'" + fVal + "'"
						currentFieldValue = fVal
						fieldValues = append(fieldValues, currentFieldValue)
						whereQuery += fmt.Sprintf("%v=$%v", govalidator.CamelCaseToUnderscore(fieldName), fieldLength)
					}
				}
			case int, uint, float32, float64, bool:
				currentFieldValue = fieldValue
				fieldValues = append(fieldValues, currentFieldValue)
				whereQuery += fmt.Sprintf("%v=$%v", govalidator.CamelCaseToUnderscore(fieldName), fieldLength)
			default:
				// json-stringify fieldValue
				if fVal, err := json.Marshal(fieldValue); err != nil {
					return whereErrMessage(fmt.Sprintf("Unknown or Unsupported field-value type: %v", err.Error()))
				} else {
					currentFieldValue = fVal
					fieldValues = append(fieldValues, currentFieldValue)
					whereQuery += fmt.Sprintf("%v=$%v", govalidator.CamelCaseToUnderscore(fieldName), fieldLength)
				}
			}
			// compute next fieldLength (where position), excluding []sting/interface{} case
			fieldLength += 1
		}
		// update fieldCount for all queryParams
		fieldCount += 1
		if whereFieldLength > 1 && fieldCount < whereFieldLength {
			whereQuery += " AND "
		}
	}

	// if all went well, return valid where-query-result
	return WhereQueryResult{
		WhereQueryObject: WhereQueryObject{
			WhereQuery:  whereQuery,
			FieldValues: fieldValues,
		},
		Ok:      true,
		Message: "success",
	}
}
