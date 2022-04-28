// @Author: abbeymart | Abi Akindele | @Created: 2020-12-08 | @Updated: 2020-12-08
// @Company: mConnect.biz | @License: MIT
// @Description: compute delete-SQL scripts

package mcdbcrud

import (
	"fmt"
)

func deleteErrMessage(errMsg string) DeleteQueryResult {
	return DeleteQueryResult{
		DeleteQueryObject: DeleteQueryObject{
			DeleteQuery: "",
			FieldValues: nil,
			WhereQuery:  WhereQueryObject{},
		},
		Ok:      false,
		Message: errMsg,
	}
}

// ComputeDeleteQueryById function computes delete SQL scripts by id(s)
func ComputeDeleteQueryById(tableName string, recordId string) DeleteQueryResult {
	if tableName == "" || recordId == "" {
		return deleteErrMessage("tableName and recordId are required for the delete-by-id operation.")
	}
	// validated recordIds, strictly contains string/UUID values, to avoid SQL-injection
	deleteQuery := fmt.Sprintf("DELETE FROM %v WHERE id=$1", tableName)
	return DeleteQueryResult{
		DeleteQueryObject: DeleteQueryObject{
			DeleteQuery: deleteQuery,
			FieldValues: []interface{}{recordId},
		},
		Ok:      true,
		Message: "success",
	}
}

// ComputeDeleteQueryByIds function computes delete SQL scripts by id(s)
func ComputeDeleteQueryByIds(tableName string, recordIds []string) DeleteQueryResult {
	if tableName == "" || len(recordIds) < 1 {
		return deleteErrMessage("tableName and recordIds are required for the delete-by-ids operation.")
	}
	// validated recordIds, strictly contains string/UUID values, to avoid SQL-injection
	// from / where condition (where-in-values)
	whereIds := ""
	idLen := len(recordIds)
	for idCount, id := range recordIds {
		whereIds += "'" + id + "'"
		if idLen > 1 && idCount < idLen-1 {
			whereIds += ", "
		}
	}
	deleteQuery := fmt.Sprintf("DELETE FROM %v WHERE id IN (%v)", tableName, whereIds)
	return DeleteQueryResult{
		DeleteQueryObject: DeleteQueryObject{
			DeleteQuery: deleteQuery,
			FieldValues: nil,
		},
		Ok:      true,
		Message: "success",
	}
}

// ComputeDeleteQueryByParam function computes delete SQL scripts by parameter specifications
func ComputeDeleteQueryByParam(tableName string, queryParam QueryParamType) DeleteQueryResult {
	if tableName == "" || len(queryParam) < 1 {
		return deleteErrMessage("tableName and queryParam (where-conditions) are required for the delete-by-param operation.")
	}
	whereRes := ComputeWhereQuery(queryParam, 1)
	if whereRes.Ok {
		deleteScript := fmt.Sprintf("DELETE FROM %v %v", tableName, whereRes.WhereQueryObject.WhereQuery)
		return DeleteQueryResult{
			DeleteQueryObject: DeleteQueryObject{
				DeleteQuery: deleteScript,
				FieldValues: whereRes.WhereQueryObject.FieldValues,
			},
			Ok:      true,
			Message: "success",
		}
	} else {
		return deleteErrMessage(fmt.Sprintf("error computing where-query condition(s): %v", whereRes.Message))
	}
}
