// @Author: abbeymart | Abi Akindele | @Created: 2020-12-08 | @Updated: 2020-12-08
// @Company: mConnect.biz | @License: MIT
// @Description: compute select-SQL script

package mcdbcrud

import (
	"fmt"
)

func selectErrMessage(errMsg string) SelectQueryResult {
	return SelectQueryResult{
		SelectQueryObject: SelectQueryObject{
			SelectQuery: "",
			FieldValues: nil,
			WhereQuery:  WhereQueryObject{},
		},
		Ok:      false,
		Message: errMsg,
	}
}

// ComputeSelectQueryAll compose select SQL script to retrieve all table-records.
// The query may be constraint by skip(offset) and limit options
func ComputeSelectQueryAll(modelRef interface{}, tableName string, options SelectQueryOptions) SelectQueryResult {
	if tableName == "" || modelRef == nil {
		return selectErrMessage("tableName and modelRef(type-struct) are required.")
	}
	// compute map[string]interface (underscore_fields) from the modelRef (struct)
	mapMod, mapErr := StructToMapUnderscore(modelRef)
	if mapErr != nil {
		return selectErrMessage(mapErr.Error())
	}
	// compute table-fields
	var fieldNames []string
	for fieldName := range mapMod {
		fieldNames = append(fieldNames, fieldName)
	}
	fieldLen := len(fieldNames)
	fieldText := ""
	for i, fieldName := range fieldNames {
		//fieldText += "'" + fieldName + "'"
		fieldText += fieldName
		if i < fieldLen-1 {
			fieldText += ", "
		}
	}
	// get records for the model-defined fields/columns
	selectQuery := fmt.Sprintf("SELECT %v FROM %v ", fieldText, tableName)

	// adjust selectQuery for skip and limit options
	if options.Limit > 0 {
		selectQuery += fmt.Sprintf(" LIMIT %v", options.Limit)
	}
	if options.Skip > 0 {
		selectQuery += fmt.Sprintf(" OFFSET %v", options.Skip)
	}

	return SelectQueryResult{
		SelectQueryObject: SelectQueryObject{
			SelectQuery: selectQuery,
			FieldValues: nil,
		},
		Ok:      true,
		Message: "success",
	}
}

// ComputeSelectQueryById compose select SQL scripts by id
func ComputeSelectQueryById(modelRef interface{}, tableName string, recordId string, options SelectQueryOptions) SelectQueryResult {
	if tableName == "" || recordId == "" || modelRef == nil {
		return selectErrMessage("tableName, modelRef(type-struct) and record-id are required.")
	}
	// compute map[string]interface (underscore_fields) from the modelRef (struct)
	mapMod, mapErr := StructToMapUnderscore(modelRef)
	if mapErr != nil {
		return selectErrMessage(mapErr.Error())
	}
	// compute table-fields
	var fieldNames []string
	for fieldName := range mapMod {
		fieldNames = append(fieldNames, fieldName)
	}
	fieldLen := len(fieldNames)
	fieldText := ""
	for i, fieldName := range fieldNames {
		//fieldText += "'" + fieldName + "'"
		fieldText += fieldName
		if i < fieldLen-1 {
			fieldText += ", "
		}
	}
	// get record(s) based on projected/provided field names ([]string)
	selectQuery := fmt.Sprintf("SELECT %v FROM %v ", fieldText, tableName)
	// from / where condition (where-in-values)
	selectQuery += fmt.Sprintf("WHERE id=$1")
	// adjust selectQuery for skip and limit options
	if options.Limit > 0 {
		selectQuery += fmt.Sprintf(" LIMIT %v", options.Limit)
	}
	if options.Skip > 0 {
		selectQuery += fmt.Sprintf(" OFFSET %v", options.Skip)
	}

	return SelectQueryResult{
		SelectQueryObject: SelectQueryObject{
			SelectQuery: selectQuery,
			FieldValues: []interface{}{recordId},
		},
		Ok:      true,
		Message: "success",
	}
}

// ComputeSelectQueryByIds compose select SQL scripts by ids
func ComputeSelectQueryByIds(modelRef interface{}, tableName string, recordIds []string, options SelectQueryOptions) SelectQueryResult {
	if tableName == "" || len(recordIds) < 1 || modelRef == nil {
		return selectErrMessage("tableName, modelRef(type-struct) and record-ids are required.")
	}
	// compute map[string]interface (underscore_fields) from the modelRef (struct)
	mapMod, mapErr := StructToMapUnderscore(modelRef)
	if mapErr != nil {
		return selectErrMessage(mapErr.Error())
	}
	// compute table-fields
	var fieldNames []string
	for fieldName := range mapMod {
		fieldNames = append(fieldNames, fieldName)
	}
	fieldLen := len(fieldNames)
	fieldText := ""
	for i, fieldName := range fieldNames {
		//fieldText += "'" + fieldName + "'"
		fieldText += fieldName
		if i < fieldLen-1 {
			fieldText += ", "
		}
	}
	// get record(s) based on projected/provided field names ([]string)
	selectQuery := fmt.Sprintf("SELECT %v FROM %v ", fieldText, tableName)
	// from / where condition (where-in-values)
	whereIds := ""
	idLen := len(recordIds)
	for idCount, id := range recordIds {
		whereIds += "'" + id + "'"
		if idLen > 1 && idCount < idLen-1 {
			whereIds += ", "
		}
	}
	selectQuery += fmt.Sprintf("WHERE id IN (%v)", whereIds)
	// adjust selectQuery for skip and limit options
	if options.Limit > 0 {
		selectQuery += fmt.Sprintf(" LIMIT %v", options.Limit)
	}
	if options.Skip > 0 {
		selectQuery += fmt.Sprintf(" OFFSET %v", options.Skip)
	}

	return SelectQueryResult{
		SelectQueryObject: SelectQueryObject{
			SelectQuery: selectQuery,
			FieldValues: nil,
		},
		Ok:      true,
		Message: "success",
	}
}

// ComputeSelectQueryByParam compose SELECT query from the where-parameters
func ComputeSelectQueryByParam(modelRef interface{}, tableName string, queryParam QueryParamType, options SelectQueryOptions) SelectQueryResult {
	if tableName == "" || len(queryParam) < 1 || modelRef == nil {
		return selectErrMessage("tableName, modelRef(type-struct) and queryParam are required.")
	}
	// compute map[string]interface (underscore_fields) from the modelRef (struct)
	mapMod, mapErr := StructToMapUnderscore(modelRef)
	if mapErr != nil {
		return selectErrMessage(mapErr.Error())
	}
	// compute table-fields
	var fieldNames []string
	for fieldName := range mapMod {
		fieldNames = append(fieldNames, fieldName)
	}
	fieldLen := len(fieldNames)
	fieldText := ""
	for i, fieldName := range fieldNames {
		//fieldText += "'" + fieldName + "'"
		fieldText += fieldName
		if i < fieldLen-1 {
			fieldText += ", "
		}
	}

	// get record(s) based on projected/provided field names ([]string)
	selectQuery := fmt.Sprintf("SELECT %v FROM %v ", fieldText, tableName)
	// add queryParam-params condition
	whereRes := ComputeWhereQuery(queryParam, 1)
	if whereRes.Ok {
		selectQuery += whereRes.WhereQueryObject.WhereQuery
		// adjust selectQuery for skip and limit options
		if options.Limit > 0 {
			selectQuery += fmt.Sprintf(" LIMIT %v", options.Limit)
		}
		if options.Skip > 0 {
			selectQuery += fmt.Sprintf(" OFFSET %v", options.Skip)
		}
		return SelectQueryResult{
			SelectQueryObject: SelectQueryObject{
				SelectQuery: selectQuery,
				FieldValues: whereRes.WhereQueryObject.FieldValues,
			},
			Ok:      true,
			Message: "success",
		}
	} else {
		return selectErrMessage(fmt.Sprintf("error computing where-query condition(s): %v", whereRes.Message))
	}
}

// TODO: select-query functions for relational tables (eager & lazy queries) and data aggregation
