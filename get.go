// @Author: abbeymart | Abi Akindele | @Created: 2020-12-01 | @Updated: 2020-12-01
// @Company: mConnect.biz | @License: MIT
// @Description: get / query record(s)

package mcdbcrud

import (
	"encoding/json"
	"fmt"
	"github.com/abbeymart/mccache"
	"github.com/abbeymart/mcresponse"
	"github.com/jmoiron/sqlx"
)

// GetById method fetches/gets/reads record that met the specified record-id,
// constrained by optional skip and limit

func (crud *Crud) GetById(id string) mcresponse.ResponseMessage {
	// check cache
	if crud.CacheResult {
		getCacheRes := mccache.GetHashCache(crud.CacheKey, crud.TableName)
		val, ok := getCacheRes.Value.(GetResultType)
		if getCacheRes.Ok && ok && len(val.Records) > 0 {
			return mcresponse.GetResMessage("success", mcresponse.ResponseMessageOptions{
				Message: "records successfully retrieved from the cache",
				Value:   val,
			})
		}
	}
	logMessage := ""
	selectOptions := SelectQueryOptions{
		Skip:  crud.Skip,
		Limit: crud.Limit,
	}
	getQueryRes := ComputeSelectQueryById(crud.ModelRef, crud.TableName, id, selectOptions)
	if !getQueryRes.Ok {
		return mcresponse.GetResMessage("readError", mcresponse.ResponseMessageOptions{
			Message: getQueryRes.Message,
			Value:   nil,
		})
	}
	//fmt.Printf("Get-query-by-id: %v \n", getQueryRes.SelectQueryObject.SelectQuery )
	//fmt.Printf("Get-by-id-values: %#v\n", getQueryRes.SelectQueryObject.FieldValues)
	// totalRecordsCount from the table
	var totalRows int
	countQuery := fmt.Sprintf("SELECT COUNT(*) AS total_rows FROM %v", crud.TableName)
	tRowErr := crud.AppDb.QueryRowx(countQuery).Scan(&totalRows)
	if tRowErr != nil {
		return mcresponse.GetResMessage("readError", mcresponse.ResponseMessageOptions{
			Message: fmt.Sprintf("Db query Error: %v", tRowErr.Error()),
			Value:   nil,
		})
	}
	// perform crud-task action
	row := crud.AppDb.QueryRowx(getQueryRes.SelectQueryObject.SelectQuery, getQueryRes.SelectQueryObject.FieldValues...)
	//fmt.Printf("get-by-id-row: %#v \n", row)
	// check rows count
	//var rowCount = 0
	var getRecords []map[string]interface{}

	// cast model as struct
	scanRowErr := row.StructScan(crud.ModelPointer)
	if scanRowErr != nil {
		return mcresponse.GetResMessage("readError", mcresponse.ResponseMessageOptions{
			Message: fmt.Sprintf("Error reading/getting records[row-scan]: %v", scanRowErr.Error()),
			Value:   nil,
		})
	}
	// transform snapshot value from model-struct to map-value
	jByte, jErr := json.Marshal(crud.ModelPointer)
	if jErr != nil {
		return mcresponse.GetResMessage("paramsError", mcresponse.ResponseMessageOptions{
			Message: fmt.Sprintf("Error transforming result-value into json-value-format: %v", jErr.Error()),
			Value:   nil,
		})
	}
	mapValue := map[string]interface{}{}
	jErr = json.Unmarshal(jByte, &mapValue)
	if jErr != nil {
		return mcresponse.GetResMessage("paramsError", mcresponse.ResponseMessageOptions{
			Message: fmt.Sprintf("Error transforming result-value into json-value-format: %v", jErr.Error()),
			Value:   nil,
		})
	}

	getRecords = append(getRecords, mapValue)

	// handles not-found-error
	if len(getRecords) < 1 {
		return mcresponse.GetResMessage("notFound", mcresponse.ResponseMessageOptions{
			Message: "RECORDS NOT FOUND.",
			Value:   nil,
		})
	}

	//rowCount += len(getRecords)
	// perform audit-log
	logRes := mcresponse.ResponseMessage{}
	var logErr error
	if crud.LogRead || crud.LogCrud {
		logRecs := map[string]interface{}{"recordIds": []string{id}}
		//jVal, _ := json.Marshal(logRecs)
		auditInfo := AuditLogOptionsType{
			TableName:  crud.TableName,
			LogRecords: LogRecordsType{LogRecords: logRecs},
		}
		if logRes, logErr = crud.TransLog.AuditLog(ReadTask, crud.UserInfo.UserId, auditInfo); logErr != nil {
			logMessage = fmt.Sprintf("Audit-log-error: %v", logErr.Error())
		} else {
			logMessage = fmt.Sprintf("Audit-log-code: %v | Message: %v", logRes.Code, logRes.Message)
		}
	}
	// result
	getResult := GetResultType{
		Records: getRecords,
		Stats: GetStatType{
			Skip:              crud.Skip,
			Limit:             crud.Limit,
			RecordsCount:      len(getRecords),
			TotalRecordsCount: totalRows,
			QueryParam:        crud.QueryParams,
			RecordIds:         crud.RecordIds,
		},
		TaskType: crud.TaskType,
		LogRes:   logRes,
	}
	// update cache
	_ = mccache.SetHashCache(crud.CacheKey, crud.TableName, getResult, int64(crud.CacheExpire))
	// response
	return mcresponse.GetResMessage("success", mcresponse.ResponseMessageOptions{
		Message: logMessage,
		Value:   getResult,
	})
}

// GetByIds method fetches/gets/reads records that met the specified record-ids,
// constrained by optional skip and limit parameters
func (crud *Crud) GetByIds() mcresponse.ResponseMessage {
	// check cache
	if crud.CacheResult {
		getCacheRes := mccache.GetHashCache(crud.TableName, crud.CacheKey)
		val, ok := getCacheRes.Value.(GetResultType)
		if getCacheRes.Ok && ok && len(val.Records) > 0 {
			return mcresponse.GetResMessage("success", mcresponse.ResponseMessageOptions{
				Message: "records successfully retrieved from the cache",
				Value:   val,
			})
		}
	}
	if len(crud.RecordIds) < 1 {
		return mcresponse.GetResMessage("paramsError",
			mcresponse.ResponseMessageOptions{
				Message: "recordIds param is required to get-record-by-id",
				Value:   nil,
			})
	}
	logMessage := ""
	selectOptions := SelectQueryOptions{
		Skip:  crud.Skip,
		Limit: crud.Limit,
	}
	getQueryRes := ComputeSelectQueryByIds(crud.ModelRef, crud.TableName, crud.RecordIds, selectOptions)
	if !getQueryRes.Ok {
		return mcresponse.GetResMessage("readError", mcresponse.ResponseMessageOptions{
			Message: getQueryRes.Message,
			Value:   nil,
		})
	}
	//fmt.Printf("Get-query-by-ids: %#v \n", getQueryRes )
	// totalRecordsCount from the table
	var totalRows int
	countQuery := fmt.Sprintf("SELECT COUNT(*) AS total_rows FROM %v", crud.TableName)
	tRowErr := crud.AppDb.QueryRowx(countQuery).Scan(&totalRows)
	if tRowErr != nil {
		return mcresponse.GetResMessage("readError", mcresponse.ResponseMessageOptions{
			Message: fmt.Sprintf("Db query Error: %v", tRowErr.Error()),
			Value:   nil,
		})
	}
	// perform crud-task action
	rows, qRowErr := crud.AppDb.Queryx(getQueryRes.SelectQueryObject.SelectQuery, getQueryRes.SelectQueryObject.FieldValues...)
	//fmt.Printf("rows-result: %v \n", rows)
	if qRowErr != nil {
		return mcresponse.GetResMessage("readError", mcresponse.ResponseMessageOptions{
			Message: fmt.Sprintf("Db query Error: %v", qRowErr.Error()),
			Value:   nil,
		})
	}
	defer func(rows *sqlx.Rows) {
		err := rows.Close()
		if err != nil {

		}
	}(rows)
	// check rows count
	//var rowCount = 0
	var getRecords []map[string]interface{}
	for rows.Next() {
		// perform crud-task action
		// cast model as struct
		scanRowErr := rows.StructScan(crud.ModelPointer)
		if scanRowErr != nil {
			return mcresponse.GetResMessage("readError", mcresponse.ResponseMessageOptions{
				Message: fmt.Sprintf("Error reading/getting records[row-scan]: %v", scanRowErr.Error()),
				Value:   nil,
			})
		}
		// transform snapshot value from model-struct to map-value
		jByte, jErr := json.Marshal(crud.ModelPointer)
		if jErr != nil {
			return mcresponse.GetResMessage("paramsError", mcresponse.ResponseMessageOptions{
				Message: fmt.Sprintf("Error transforming result-value into json-value-format: %v", jErr.Error()),
				Value:   nil,
			})
		}
		mapValue := map[string]interface{}{}
		jErr = json.Unmarshal(jByte, &mapValue)
		if jErr != nil {
			return mcresponse.GetResMessage("paramsError", mcresponse.ResponseMessageOptions{
				Message: fmt.Sprintf("Error transforming result-value into json-value-format: %v", jErr.Error()),
				Value:   nil,
			})
		}

		getRecords = append(getRecords, mapValue)
	}

	// handles not-found-error
	if len(getRecords) < 1 {
		return mcresponse.GetResMessage("notFound", mcresponse.ResponseMessageOptions{
			Message: "RECORDS NOT FOUND.",
			Value:   nil,
		})
	}
	// check record-rows error
	if rowErr := rows.Err(); rowErr != nil {
		return mcresponse.GetResMessage("readError", mcresponse.ResponseMessageOptions{
			Message: fmt.Sprintf("Error reading/getting records: %v", rowErr.Error()),
			Value: GetResultType{
				Records:  nil,
				Stats:    GetStatType{},
				TaskType: crud.TaskType,
				LogRes:   mcresponse.ResponseMessage{},
			},
		})
	}
	// perform audit-log
	logRes := mcresponse.ResponseMessage{}
	var logErr error
	if crud.LogRead || crud.LogCrud {
		logRecs := map[string]interface{}{"recordIds": crud.RecordIds}
		//jVal, _ := json.Marshal(logRecs)
		auditInfo := AuditLogOptionsType{
			TableName:  crud.TableName,
			LogRecords: LogRecordsType{LogRecords: logRecs},
		}
		if logRes, logErr = crud.TransLog.AuditLog(ReadTask, crud.UserInfo.UserId, auditInfo); logErr != nil {
			logMessage = fmt.Sprintf("Audit-log-error: %v", logErr.Error())
		} else {
			logMessage = fmt.Sprintf("Audit-log-code: %v | Message: %v", logRes.Code, logRes.Message)
		}
	}
	// result
	getResult := GetResultType{
		Records: getRecords,
		Stats: GetStatType{
			Skip:              crud.Skip,
			Limit:             crud.Limit,
			RecordsCount:      len(getRecords),
			TotalRecordsCount: totalRows,
			QueryParam:        crud.QueryParams,
			RecordIds:         crud.RecordIds,
		},
		TaskType: crud.TaskType,
		LogRes:   logRes,
	}
	// update cache
	_ = mccache.SetHashCache(crud.CacheKey, crud.TableName, getResult, int64(crud.CacheExpire))
	// response
	return mcresponse.GetResMessage("success", mcresponse.ResponseMessageOptions{
		Message: logMessage,
		Value:   getResult,
	})
}

// GetByParam method fetches/gets/reads records that met the specified query-params or where conditions,
// constrained by optional skip and limit parameters
func (crud *Crud) GetByParam() mcresponse.ResponseMessage {
	// check cache
	if crud.CacheResult {
		getCacheRes := mccache.GetHashCache(crud.TableName, crud.CacheKey)
		val, ok := getCacheRes.Value.(GetResultType)
		if getCacheRes.Ok && ok && len(val.Records) > 0 {
			return mcresponse.GetResMessage("success", mcresponse.ResponseMessageOptions{
				Message: "records successfully retrieved from the cache",
				Value:   val,
			})
		}
	}
	logMessage := ""
	selectOptions := SelectQueryOptions{
		Skip:  crud.Skip,
		Limit: crud.Limit,
	}
	getQueryRes := ComputeSelectQueryByParam(crud.ModelRef, crud.TableName, crud.QueryParams, selectOptions)
	if !getQueryRes.Ok {
		return mcresponse.GetResMessage("readError", mcresponse.ResponseMessageOptions{
			Message: getQueryRes.Message,
			Value:   nil,
		})
	}
	//fmt.Printf("\n Get-query-by-params: %#v \n\n", getQueryRes)
	// totalRecordsCount from the table
	var totalRows int
	countQuery := fmt.Sprintf("SELECT COUNT(*) AS total_rows FROM %v", crud.TableName)
	tRowErr := crud.AppDb.QueryRowx(countQuery).Scan(&totalRows)
	if tRowErr != nil {
		return mcresponse.GetResMessage("readError", mcresponse.ResponseMessageOptions{
			Message: fmt.Sprintf("Db query Error: %v", tRowErr.Error()),
			Value:   nil,
		})
	}
	// perform crud-task action
	rows, qRowErr := crud.AppDb.Queryx(getQueryRes.SelectQueryObject.SelectQuery, getQueryRes.SelectQueryObject.FieldValues...)
	if qRowErr != nil {
		return mcresponse.GetResMessage("readError", mcresponse.ResponseMessageOptions{
			Message: fmt.Sprintf("Db query Error: %v", qRowErr.Error()),
			Value:   nil,
		})
	}
	defer func(rows *sqlx.Rows) {
		err := rows.Close()
		if err != nil {

		}
	}(rows)
	// check rows count
	//var rowCount = 0
	var getRecords []map[string]interface{}
	for rows.Next() {
		// perform crud-task action
		scanRowErr := rows.StructScan(crud.ModelPointer)
		if scanRowErr != nil {
			return mcresponse.GetResMessage("readError", mcresponse.ResponseMessageOptions{
				Message: fmt.Sprintf("Error reading/getting records[row-scan]: %v", scanRowErr.Error()),
				Value:   nil,
			})
		}
		// transform snapshot value from model-struct to map-value
		jByte, jErr := json.Marshal(crud.ModelPointer)
		if jErr != nil {
			return mcresponse.GetResMessage("paramsError", mcresponse.ResponseMessageOptions{
				Message: fmt.Sprintf("Error transforming result-value into json-value-format: %v", jErr.Error()),
				Value:   nil,
			})
		}
		mapValue := map[string]interface{}{}
		jErr = json.Unmarshal(jByte, &mapValue)
		if jErr != nil {
			return mcresponse.GetResMessage("paramsError", mcresponse.ResponseMessageOptions{
				Message: fmt.Sprintf("Error transforming result-value into json-value-format: %v", jErr.Error()),
				Value:   nil,
			})
		}

		getRecords = append(getRecords, mapValue)
	}
	// handles not-found-error
	if len(getRecords) < 1 {
		return mcresponse.GetResMessage("notFound", mcresponse.ResponseMessageOptions{
			Message: "RECORDS NOT FOUND.",
			Value:   nil,
		})
	}
	// check record-rows error
	if rowErr := rows.Err(); rowErr != nil {
		return mcresponse.GetResMessage("readError", mcresponse.ResponseMessageOptions{
			Message: fmt.Sprintf("Error reading/getting records: %v", rowErr.Error()),
			Value: GetResultType{
				Records:  nil,
				Stats:    GetStatType{},
				TaskType: crud.TaskType,
				LogRes:   mcresponse.ResponseMessage{},
			},
		})
	}
	// perform audit-log
	logRes := mcresponse.ResponseMessage{}
	var logErr error
	if crud.LogRead || crud.LogCrud {
		logRecs := map[string]interface{}{"queryParams": crud.QueryParams}
		//jVal, _ := json.Marshal(logRecs)
		auditInfo := AuditLogOptionsType{
			TableName:  crud.TableName,
			LogRecords: LogRecordsType{LogRecords: logRecs},
		}
		if logRes, logErr = crud.TransLog.AuditLog(ReadTask, crud.UserInfo.UserId, auditInfo); logErr != nil {
			logMessage = fmt.Sprintf("Audit-log-error: %v", logErr.Error())
		} else {
			logMessage = fmt.Sprintf("Audit-log-code: %v | Message: %v", logRes.Code, logRes.Message)
		}
	}
	// result
	getResult := GetResultType{
		Records: getRecords,
		Stats: GetStatType{
			Skip:              crud.Skip,
			Limit:             crud.Limit,
			RecordsCount:      len(getRecords),
			TotalRecordsCount: totalRows,
			QueryParam:        crud.QueryParams,
			RecordIds:         crud.RecordIds,
		},
		TaskType: crud.TaskType,
		LogRes:   logRes,
	}
	// update cache
	_ = mccache.SetHashCache(crud.CacheKey, crud.TableName, getResult, int64(crud.CacheExpire))
	// response
	return mcresponse.GetResMessage("success", mcresponse.ResponseMessageOptions{
		Message: fmt.Sprintf("AmountRecord(s) read-query completed successfully [log-message: %v]", logMessage),
		Value:   getResult,
	})
}

// GetAll method fetches/gets/reads all record(s), constrained by optional skip and limit parameters
func (crud *Crud) GetAll() mcresponse.ResponseMessage {
	// compute select-query
	selectOptions := SelectQueryOptions{
		Skip:  crud.Skip,
		Limit: crud.Limit,
	}
	getQueryRes := ComputeSelectQueryAll(crud.ModelRef, crud.TableName, selectOptions)
	if !getQueryRes.Ok {
		return mcresponse.GetResMessage("readError", mcresponse.ResponseMessageOptions{
			Message: getQueryRes.Message,
			Value:   nil,
		})
	}
	// totalRecordsCount from the table
	var totalRows int
	countQuery := fmt.Sprintf("SELECT COUNT(*) AS total_rows FROM %v", crud.TableName)
	tRowErr := crud.AppDb.QueryRowx(countQuery).Scan(&totalRows)
	if tRowErr != nil {
		return mcresponse.GetResMessage("readError", mcresponse.ResponseMessageOptions{
			Message: fmt.Sprintf("Db query Error[scan-total-records-count]: %v", tRowErr.Error()),
			Value:   nil,
		})
	}
	// perform crud-task action
	rows, qRowErr := crud.AppDb.Queryx(getQueryRes.SelectQueryObject.SelectQuery, getQueryRes.SelectQueryObject.FieldValues...)
	if qRowErr != nil {
		return mcresponse.GetResMessage("readError", mcresponse.ResponseMessageOptions{
			Message: fmt.Sprintf("Db query Error: %v", qRowErr.Error()),
			Value:   nil,
		})
	}
	//fmt.Printf("row-records: %#v", rows)
	defer func(rows *sqlx.Rows) {
		err := rows.Close()
		if err != nil {

		}
	}(rows)
	// check rows count
	//var rowCount = 0
	var getRecords []map[string]interface{}
	for rows.Next() {
		// perform crud-task action
		// cast model as struct
		scanRowErr := rows.StructScan(crud.ModelPointer)
		if scanRowErr != nil {
			return mcresponse.GetResMessage("readError", mcresponse.ResponseMessageOptions{
				Message: fmt.Sprintf("Error reading/getting records[row-scan]: %v", scanRowErr.Error()),
				Value:   nil,
			})
		}
		// transform snapshot value from model-struct to map-value
		jByte, jErr := json.Marshal(crud.ModelPointer)
		if jErr != nil {
			return mcresponse.GetResMessage("paramsError", mcresponse.ResponseMessageOptions{
				Message: fmt.Sprintf("Error transforming result-value into json-value-format: %v", jErr.Error()),
				Value:   nil,
			})
		}
		mapValue := map[string]interface{}{}
		jErr = json.Unmarshal(jByte, &mapValue)
		if jErr != nil {
			return mcresponse.GetResMessage("paramsError", mcresponse.ResponseMessageOptions{
				Message: fmt.Sprintf("Error transforming result-value into json-value-format: %v", jErr.Error()),
				Value:   nil,
			})
		}
		//fmt.Printf("query-record: %#v\n\n", mapValue)
		getRecords = append(getRecords, mapValue)
	}
	// handles not-found-error
	if len(getRecords) < 1 {
		return mcresponse.GetResMessage("notFound", mcresponse.ResponseMessageOptions{
			Message: "RECORDS NOT FOUND.",
			Value:   nil,
		})
	}
	if rowErr := rows.Err(); rowErr != nil {
		return mcresponse.GetResMessage("readError", mcresponse.ResponseMessageOptions{
			Message: fmt.Sprintf("Error reading/getting records: %v", rowErr.Error()),
			Value: GetResultType{
				Records:  nil,
				Stats:    GetStatType{},
				TaskType: crud.TaskType,
				LogRes:   mcresponse.ResponseMessage{},
			},
		})
	}
	// perform audit-log | initialize log-variables
	logMessage := ""
	logRes := mcresponse.ResponseMessage{}
	var logErr error
	if crud.LogRead || crud.LogCrud {
		logRecs := map[string]interface{}{"query": "all"}
		//jVal, _ := json.Marshal(logRecs)
		auditInfo := AuditLogOptionsType{
			TableName:  crud.TableName,
			LogRecords: LogRecordsType{LogRecords: logRecs},
		}
		if logRes, logErr = crud.TransLog.AuditLog(ReadTask, crud.UserInfo.UserId, auditInfo); logErr != nil {
			logMessage = fmt.Sprintf("Audit-log-error: %v", logErr.Error())
		} else {
			logMessage = fmt.Sprintf("Audit-log-code: %v | Message: %v", logRes.Code, logRes.Message)
		}
	}
	// result
	getResult := GetResultType{
		Records: getRecords,
		Stats: GetStatType{
			Skip:              crud.Skip,
			Limit:             crud.Limit,
			RecordsCount:      len(getRecords),
			TotalRecordsCount: totalRows,
			QueryParam:        crud.QueryParams,
			RecordIds:         crud.RecordIds,
		},
		TaskType: crud.TaskType,
		LogRes:   logRes,
	}
	// update cache | *****don't cache all-table-records, due to large/unknown size*****
	//_ = mccache.SetHashCache(crud.CacheKey, crud.TableName, getRecords, uint(crud.CacheExpire))

	return mcresponse.GetResMessage("success", mcresponse.ResponseMessageOptions{
		Message: fmt.Sprintf("AmountRecord(s) read-query completed successfully [log-message: %v]", logMessage),
		Value:   getResult,
	})
}

// CustomSelectQuery method obtain the query result for the specified selectQuery, tableName and modelPointer and optional fieldPositionalValues.
func (crud *Crud) CustomSelectQuery(params CustomSelectQueryParamsType) mcresponse.ResponseMessage {
	//  validate required parameters
	if params.SelectQuery == "" || params.TableName == "" || params.ModelPointer == nil {
		return mcresponse.ResponseMessage{
			Code:    "paramsError",
			Message: fmt.Sprintf("Valid selectQuery, tableName and modelPointer are required"),
			Value:   nil,
		}
	}
	// totalRecordsCount, for the query-condition, from the table
	var totalRows int
	countQuery := params.CountQuery
	if countQuery == "" {
		countQuery = fmt.Sprintf("SELECT COUNT(*) AS total_rows FROM %v", params.TableName)
	}
	tRowErr := crud.AppDb.QueryRowx(countQuery).Scan(&totalRows)
	if tRowErr != nil {
		return mcresponse.GetResMessage("readError", mcresponse.ResponseMessageOptions{
			Message: fmt.Sprintf("Db query Error[scan-total-records-count]: %v", tRowErr.Error()),
			Value:   nil,
		})
	}
	// perform crud-task action
	rows, qRowErr := crud.AppDb.Queryx(params.SelectQuery, params.QueryPositionalFieldValues...)
	if qRowErr != nil {
		return mcresponse.GetResMessage("readError", mcresponse.ResponseMessageOptions{
			Message: fmt.Sprintf("Db query Error: %v", qRowErr.Error()),
			Value:   nil,
		})
	}
	defer func(rows *sqlx.Rows) {
		err := rows.Close()
		if err != nil {

		}
	}(rows)
	// check rows count
	modelPointer := params.ModelPointer
	var getRecords []map[string]interface{}
	for rows.Next() {
		// perform crud-task action
		// cast model as struct
		scanRowErr := rows.StructScan(modelPointer)
		if scanRowErr != nil {
			return mcresponse.GetResMessage("readError", mcresponse.ResponseMessageOptions{
				Message: fmt.Sprintf("Error reading/getting records[row-scan]: %v", scanRowErr.Error()),
				Value:   nil,
			})
		}
		// transform snapshot value from model-struct to map-value
		jByte, jErr := json.Marshal(modelPointer)
		if jErr != nil {
			return mcresponse.GetResMessage("paramsError", mcresponse.ResponseMessageOptions{
				Message: fmt.Sprintf("Error transforming result-value into json-value-format: %v", jErr.Error()),
				Value:   nil,
			})
		}
		mapValue := map[string]interface{}{}
		jErr = json.Unmarshal(jByte, &mapValue)
		if jErr != nil {
			return mcresponse.GetResMessage("paramsError", mcresponse.ResponseMessageOptions{
				Message: fmt.Sprintf("Error transforming result-value into json-value-format: %v", jErr.Error()),
				Value:   nil,
			})
		}
		//fmt.Printf("query-record: %#v\n\n", mapValue)
		getRecords = append(getRecords, mapValue)
	}
	// handles not-found-error
	if len(getRecords) < 1 {
		return mcresponse.GetResMessage("notFound", mcresponse.ResponseMessageOptions{
			Message: "RECORDS NOT FOUND.",
			Value:   nil,
		})
	}
	if rowErr := rows.Err(); rowErr != nil {
		return mcresponse.GetResMessage("readError", mcresponse.ResponseMessageOptions{
			Message: fmt.Sprintf("Error reading/getting records: %v", rowErr.Error()),
			Value: GetResultType{
				Records:  nil,
				Stats:    GetStatType{},
				TaskType: crud.TaskType,
				LogRes:   mcresponse.ResponseMessage{},
			},
		})
	}
	// result
	getResult := GetResultType{
		Records: getRecords,
		Stats: GetStatType{
			Skip:              0,
			Limit:             crud.MaxQueryLimit,
			RecordsCount:      len(getRecords),
			TotalRecordsCount: totalRows,
			QueryParam:        QueryParamType{},
			RecordIds:         []string{},
		},
		TaskType: crud.TaskType,
	}

	return mcresponse.GetResMessage("success", mcresponse.ResponseMessageOptions{
		Message: fmt.Sprintf("Record(s) read-query completed successfully"),
		Value:   getResult,
	})
}

// get-scan-to-map

func (crud *Crud) GetById1(id string) mcresponse.ResponseMessage {
	// check cache
	if crud.CacheResult {
		getCacheRes := mccache.GetHashCache(crud.CacheKey, crud.TableName)
		val, ok := getCacheRes.Value.(GetResultType)
		if getCacheRes.Ok && ok && len(val.Records) > 0 {
			return mcresponse.GetResMessage("success", mcresponse.ResponseMessageOptions{
				Message: "records successfully retrieved from the cache",
				Value:   val,
			})
		}
	}
	logMessage := ""
	selectOptions := SelectQueryOptions{
		Skip:  crud.Skip,
		Limit: crud.Limit,
	}
	getQueryRes := ComputeSelectQueryById(crud.ModelRef, crud.TableName, id, selectOptions)
	if !getQueryRes.Ok {
		return mcresponse.GetResMessage("readError", mcresponse.ResponseMessageOptions{
			Message: getQueryRes.Message,
			Value:   nil,
		})
	}
	//fmt.Printf("Get-query-by-id: %v \n", getQueryRes.SelectQueryObject.SelectQuery )
	// totalRecordsCount from the table
	var totalRows int
	countQuery := fmt.Sprintf("SELECT COUNT(*) AS total_rows FROM %v", crud.TableName)
	tRowErr := crud.AppDb.QueryRowx(countQuery).Scan(&totalRows)
	if tRowErr != nil {
		return mcresponse.GetResMessage("readError", mcresponse.ResponseMessageOptions{
			Message: fmt.Sprintf("Db query Error: %v", tRowErr.Error()),
			Value:   nil,
		})
	}
	// perform crud-task action

	mapRes := make(map[string]interface{})
	row := crud.AppDb.QueryRowx(getQueryRes.SelectQueryObject.SelectQuery, getQueryRes.SelectQueryObject.FieldValues...)
	//fmt.Printf("get-by-id-row: %v \n", row)
	qRowErr := row.MapScan(mapRes)
	if qRowErr != nil {
		return mcresponse.GetResMessage("readError", mcresponse.ResponseMessageOptions{
			Message: fmt.Sprintf("Error reading/getting records[row-scan]: %v", qRowErr.Error()),
			Value:   nil,
		})
	}
	// check rows count
	//var rowCount = 0
	var getRecords []map[string]interface{}

	// transform snapshot value from model-struct to map-value
	jByte, jErr := json.Marshal(mapRes)
	if jErr != nil {
		return mcresponse.GetResMessage("paramsError", mcresponse.ResponseMessageOptions{
			Message: fmt.Sprintf("Error transforming result-value into json-value-format: %v", jErr.Error()),
			Value:   nil,
		})
	}
	//fmt.Printf("map-scanned-result: %v \n", mapRes)
	mapValue := map[string]interface{}{}
	jErr = json.Unmarshal(jByte, &mapValue)
	if jErr != nil {
		return mcresponse.GetResMessage("paramsError", mcresponse.ResponseMessageOptions{
			Message: fmt.Sprintf("Error transforming result-value into json-value-format: %v", jErr.Error()),
			Value:   nil,
		})
	}

	//fmt.Printf("map-scanned-result: %v \n", mapValue)
	mapVal, mapErr := MapToMapCamelCase(mapValue, crud.FieldSeparator)
	if mapErr != nil {
		return mcresponse.GetResMessage("paramsError", mcresponse.ResponseMessageOptions{
			Message: fmt.Sprintf("map-underscore-to-camelCase-error: %v", mapErr.Error()),
			Value:   nil,
		})
	}

	getRecords = append(getRecords, mapVal)

	// handles not-found-error
	if len(getRecords) < 1 {
		return mcresponse.GetResMessage("notFound", mcresponse.ResponseMessageOptions{
			Message: "RECORDS NOT FOUND.",
			Value:   nil,
		})
	}

	//rowCount += len(getRecords)
	// perform audit-log
	logRes := mcresponse.ResponseMessage{}
	var logErr error
	if crud.LogRead || crud.LogCrud {
		logRecs := map[string]interface{}{"recordIds": []string{id}}
		//jVal, _ := json.Marshal(logRecs)
		auditInfo := AuditLogOptionsType{
			TableName:  crud.TableName,
			LogRecords: LogRecordsType{LogRecords: logRecs},
		}
		if logRes, logErr = crud.TransLog.AuditLog(ReadTask, crud.UserInfo.UserId, auditInfo); logErr != nil {
			logMessage = fmt.Sprintf("Audit-log-error: %v", logErr.Error())
		} else {
			logMessage = fmt.Sprintf("Audit-log-code: %v | Message: %v", logRes.Code, logRes.Message)
		}
	}
	// result
	getResult := GetResultType{
		Records: getRecords,
		Stats: GetStatType{
			Skip:              crud.Skip,
			Limit:             crud.Limit,
			RecordsCount:      len(getRecords),
			TotalRecordsCount: totalRows,
			QueryParam:        crud.QueryParams,
			RecordIds:         crud.RecordIds,
		},
		TaskType: crud.TaskType,
		LogRes:   logRes,
	}
	// update cache
	_ = mccache.SetHashCache(crud.CacheKey, crud.TableName, getResult, int64(crud.CacheExpire))
	// response
	return mcresponse.GetResMessage("success", mcresponse.ResponseMessageOptions{
		Message: fmt.Sprintf("AmountRecord(s) read-query completed successfully [log-message: %v]", logMessage),
		Value:   getResult,
	})
}

func (crud *Crud) GetByIds1() mcresponse.ResponseMessage {
	// check cache
	if crud.CacheResult {
		getCacheRes := mccache.GetHashCache(crud.TableName, crud.CacheKey)
		val, ok := getCacheRes.Value.(GetResultType)
		if getCacheRes.Ok && ok && len(val.Records) > 0 {
			return mcresponse.GetResMessage("success", mcresponse.ResponseMessageOptions{
				Message: "records successfully retrieved from the cache",
				Value:   val,
			})
		}
	}
	if len(crud.RecordIds) < 1 {
		return mcresponse.GetResMessage("paramsError",
			mcresponse.ResponseMessageOptions{
				Message: "recordIds param is required to get-record-by-id",
				Value:   nil,
			})
	}
	logMessage := ""
	selectOptions := SelectQueryOptions{
		Skip:  crud.Skip,
		Limit: crud.Limit,
	}
	getQueryRes := ComputeSelectQueryByIds(crud.ModelRef, crud.TableName, crud.RecordIds, selectOptions)
	if !getQueryRes.Ok {
		return mcresponse.GetResMessage("readError", mcresponse.ResponseMessageOptions{
			Message: getQueryRes.Message,
			Value:   nil,
		})
	}
	//fmt.Printf("Get-query-by-ids: %#v \n", getQueryRes )
	// totalRecordsCount from the table
	var totalRows int
	countQuery := fmt.Sprintf("SELECT COUNT(*) AS total_rows FROM %v", crud.TableName)
	tRowErr := crud.AppDb.QueryRowx(countQuery).Scan(&totalRows)
	if tRowErr != nil {
		return mcresponse.GetResMessage("readError", mcresponse.ResponseMessageOptions{
			Message: fmt.Sprintf("Db query Error: %v", tRowErr.Error()),
			Value:   nil,
		})
	}
	// perform crud-task action
	rows, qRowErr := crud.AppDb.Queryx(getQueryRes.SelectQueryObject.SelectQuery, getQueryRes.SelectQueryObject.FieldValues...)
	//fmt.Printf("rows-result: %v \n", rows)
	if qRowErr != nil {
		return mcresponse.GetResMessage("readError", mcresponse.ResponseMessageOptions{
			Message: fmt.Sprintf("Db query Error: %v", qRowErr.Error()),
			Value:   nil,
		})
	}
	defer func(rows *sqlx.Rows) {
		err := rows.Close()
		if err != nil {

		}
	}(rows)
	// check rows count
	//var rowCount = 0
	var getRecords []map[string]interface{}
	for rows.Next() {
		mapRes := make(map[string]interface{})
		if rowScanErr := rows.MapScan(mapRes); rowScanErr != nil {
			return mcresponse.GetResMessage("readError", mcresponse.ResponseMessageOptions{
				Message: fmt.Sprintf("Error reading/getting records[row-scan]: %v", rowScanErr.Error()),
				Value:   nil,
			})
		} else {
			// transform snapshot value from model-struct to map-value
			jByte, jErr := json.Marshal(mapRes)
			if jErr != nil {
				return mcresponse.GetResMessage("paramsError", mcresponse.ResponseMessageOptions{
					Message: fmt.Sprintf("Error transforming result-value into json-value-format: %v", jErr.Error()),
					Value:   nil,
				})
			}
			mapValue := map[string]interface{}{}
			jErr = json.Unmarshal(jByte, &mapValue)
			if jErr != nil {
				return mcresponse.GetResMessage("paramsError", mcresponse.ResponseMessageOptions{
					Message: fmt.Sprintf("Error transforming result-value into json-value-format: %v", jErr.Error()),
					Value:   nil,
				})
			}

			//fmt.Printf("map-scanned-result: %v \n", mapValue)
			mapVal, mapErr := MapToMapCamelCase(mapValue, crud.FieldSeparator)
			if mapErr != nil {
				return mcresponse.GetResMessage("paramsError", mcresponse.ResponseMessageOptions{
					Message: fmt.Sprintf("map-underscore-to-camelCase-error: %v", mapErr.Error()),
					Value:   nil,
				})
			}

			//fmt.Printf("map-transformed-result: %v \n", mapVal)
			getRecords = append(getRecords, mapVal)
			//rowCount += 1
			//fmt.Printf("Get-query-result: %v", mapValue)
		}
	}
	// handles not-found-error
	if len(getRecords) < 1 {
		return mcresponse.GetResMessage("notFound", mcresponse.ResponseMessageOptions{
			Message: "RECORDS NOT FOUND.",
			Value:   nil,
		})
	}
	// check record-rows error
	if rowErr := rows.Err(); rowErr != nil {
		return mcresponse.GetResMessage("readError", mcresponse.ResponseMessageOptions{
			Message: fmt.Sprintf("Error reading/getting records: %v", rowErr.Error()),
			Value: GetResultType{
				Records:  nil,
				Stats:    GetStatType{},
				TaskType: crud.TaskType,
				LogRes:   mcresponse.ResponseMessage{},
			},
		})
	}
	// perform audit-log
	logRes := mcresponse.ResponseMessage{}
	var logErr error
	if crud.LogRead || crud.LogCrud {
		logRecs := map[string]interface{}{"recordIds": crud.RecordIds}
		//jVal, _ := json.Marshal(logRecs)
		auditInfo := AuditLogOptionsType{
			TableName:  crud.TableName,
			LogRecords: LogRecordsType{LogRecords: logRecs},
		}
		if logRes, logErr = crud.TransLog.AuditLog(ReadTask, crud.UserInfo.UserId, auditInfo); logErr != nil {
			logMessage = fmt.Sprintf("Audit-log-error: %v", logErr.Error())
		} else {
			logMessage = fmt.Sprintf("Audit-log-code: %v | Message: %v", logRes.Code, logRes.Message)
		}
	}
	// result
	getResult := GetResultType{
		Records: getRecords,
		Stats: GetStatType{
			Skip:              crud.Skip,
			Limit:             crud.Limit,
			RecordsCount:      len(getRecords),
			TotalRecordsCount: totalRows,
			QueryParam:        crud.QueryParams,
			RecordIds:         crud.RecordIds,
		},
		TaskType: crud.TaskType,
		LogRes:   logRes,
	}
	// update cache
	_ = mccache.SetHashCache(crud.CacheKey, crud.TableName, getResult, int64(crud.CacheExpire))
	// response
	return mcresponse.GetResMessage("success", mcresponse.ResponseMessageOptions{
		Message: fmt.Sprintf("AmountRecord(s) read-query completed successfully [log-message: %v]", logMessage),
		Value:   getResult,
	})
}

// GetByParam1 method fetches/gets/reads records that met the specified query-params or where conditions,
// constrained by optional skip and limit parameters
func (crud *Crud) GetByParam1() mcresponse.ResponseMessage {
	// check cache
	if crud.CacheResult {
		getCacheRes := mccache.GetHashCache(crud.TableName, crud.CacheKey)
		val, ok := getCacheRes.Value.(GetResultType)
		if getCacheRes.Ok && ok && len(val.Records) > 0 {
			return mcresponse.GetResMessage("success", mcresponse.ResponseMessageOptions{
				Message: "records successfully retrieved from the cache",
				Value:   val,
			})
		}
	}
	logMessage := ""
	selectOptions := SelectQueryOptions{
		Skip:  crud.Skip,
		Limit: crud.Limit,
	}
	getQueryRes := ComputeSelectQueryByParam(crud.ModelRef, crud.TableName, crud.QueryParams, selectOptions)
	if !getQueryRes.Ok {
		return mcresponse.GetResMessage("readError", mcresponse.ResponseMessageOptions{
			Message: getQueryRes.Message,
			Value:   nil,
		})
	}
	//fmt.Printf("Get-query-by-params: %#v \n\n", getQueryRes )
	// totalRecordsCount from the table
	var totalRows int
	countQuery := fmt.Sprintf("SELECT COUNT(*) AS total_rows FROM %v", crud.TableName)
	tRowErr := crud.AppDb.QueryRowx(countQuery).Scan(&totalRows)
	if tRowErr != nil {
		return mcresponse.GetResMessage("readError", mcresponse.ResponseMessageOptions{
			Message: fmt.Sprintf("Db query Error: %v", tRowErr.Error()),
			Value:   nil,
		})
	}
	// perform crud-task action
	rows, qRowErr := crud.AppDb.Queryx(getQueryRes.SelectQueryObject.SelectQuery, getQueryRes.SelectQueryObject.FieldValues...)
	if qRowErr != nil {
		return mcresponse.GetResMessage("readError", mcresponse.ResponseMessageOptions{
			Message: fmt.Sprintf("Db query Error: %v", qRowErr.Error()),
			Value:   nil,
		})
	}
	defer func(rows *sqlx.Rows) {
		err := rows.Close()
		if err != nil {

		}
	}(rows)
	// check rows count
	//var rowCount = 0
	var getRecords []map[string]interface{}
	for rows.Next() {
		mapRes := map[string]interface{}{}
		rowScanErr := rows.MapScan(mapRes)
		if rowScanErr != nil {
			return mcresponse.GetResMessage("readError", mcresponse.ResponseMessageOptions{
				Message: fmt.Sprintf("Error reading/getting records[row-scan-map]: %v", rowScanErr.Error()),
				Value:   nil,
			})
		}
		fmt.Printf("row-map-res: %#v", mapRes)
		// transform snapshot value from model-struct to map-value
		jByte, jErr := json.Marshal(mapRes)
		if jErr != nil {
			return mcresponse.GetResMessage("paramsError", mcresponse.ResponseMessageOptions{
				Message: fmt.Sprintf("Error transforming result-value into json-value-format: %v", jErr.Error()),
				Value:   nil,
			})
		}
		mapValue := map[string]interface{}{}
		jErr = json.Unmarshal(jByte, &mapValue)
		if jErr != nil {
			return mcresponse.GetResMessage("paramsError", mcresponse.ResponseMessageOptions{
				Message: fmt.Sprintf("Error transforming result-value into json-value-format: %v", jErr.Error()),
				Value:   nil,
			})
		}
		fmt.Printf("map-scanned-result: %v \n", mapValue)
		mapVal, mapErr := MapToMapCamelCase(mapValue, crud.FieldSeparator)
		if mapErr != nil {
			return mcresponse.GetResMessage("paramsError", mcresponse.ResponseMessageOptions{
				Message: fmt.Sprintf("map-underscore-to-camelCase-error: %v", mapErr.Error()),
				Value:   nil,
			})
		}
		fmt.Printf("map-transformed-result(camelCase): %v \n", mapVal)
		getRecords = append(getRecords, mapVal)
		//rowCount += 1
	}
	// handles not-found-error
	if len(getRecords) < 1 {
		return mcresponse.GetResMessage("notFound", mcresponse.ResponseMessageOptions{
			Message: "RECORDS NOT FOUND.",
			Value:   nil,
		})
	}
	// check record-rows error
	if rowErr := rows.Err(); rowErr != nil {
		return mcresponse.GetResMessage("readError", mcresponse.ResponseMessageOptions{
			Message: fmt.Sprintf("Error reading/getting records: %v", rowErr.Error()),
			Value: GetResultType{
				Records:  nil,
				Stats:    GetStatType{},
				TaskType: crud.TaskType,
				LogRes:   mcresponse.ResponseMessage{},
			},
		})
	}
	// perform audit-log
	logRes := mcresponse.ResponseMessage{}
	var logErr error
	if crud.LogRead || crud.LogCrud {
		logRecs := map[string]interface{}{"queryParams": crud.QueryParams}
		//jVal, _ := json.Marshal(logRecs)
		auditInfo := AuditLogOptionsType{
			TableName:  crud.TableName,
			LogRecords: LogRecordsType{LogRecords: logRecs},
		}
		if logRes, logErr = crud.TransLog.AuditLog(ReadTask, crud.UserInfo.UserId, auditInfo); logErr != nil {
			logMessage = fmt.Sprintf("Audit-log-error: %v", logErr.Error())
		} else {
			logMessage = fmt.Sprintf("Audit-log-code: %v | Message: %v", logRes.Code, logRes.Message)
		}
	}
	// result
	getResult := GetResultType{
		Records: getRecords,
		Stats: GetStatType{
			Skip:              crud.Skip,
			Limit:             crud.Limit,
			RecordsCount:      len(getRecords),
			TotalRecordsCount: totalRows,
			QueryParam:        crud.QueryParams,
			RecordIds:         crud.RecordIds,
		},
		TaskType: crud.TaskType,
		LogRes:   logRes,
	}
	// update cache
	_ = mccache.SetHashCache(crud.CacheKey, crud.TableName, getResult, int64(crud.CacheExpire))
	// response
	return mcresponse.GetResMessage("success", mcresponse.ResponseMessageOptions{
		Message: logMessage,
		Value:   getResult,
	})
}

// GetAll1 method fetches/gets/reads all record(s), constrained by optional skip and limit parameters
func (crud *Crud) GetAll1() mcresponse.ResponseMessage {
	// compute select-query
	selectOptions := SelectQueryOptions{
		Skip:  crud.Skip,
		Limit: crud.Limit,
	}
	getQueryRes := ComputeSelectQueryAll(crud.ModelRef, crud.TableName, selectOptions)
	if !getQueryRes.Ok {
		return mcresponse.GetResMessage("readError", mcresponse.ResponseMessageOptions{
			Message: getQueryRes.Message,
			Value:   nil,
		})
	}
	//fmt.Printf("Get-query-by-all: %#v", getQueryRes )
	// totalRecordsCount from the table
	var totalRows int
	countQuery := fmt.Sprintf("SELECT COUNT(*) AS total_rows FROM %v", crud.TableName)
	tRowErr := crud.AppDb.QueryRowx(countQuery).Scan(&totalRows)
	if tRowErr != nil {
		return mcresponse.GetResMessage("readError", mcresponse.ResponseMessageOptions{
			Message: fmt.Sprintf("Db query Error: %v", tRowErr.Error()),
			Value:   nil,
		})
	}
	// perform crud-task action
	rows, qRowErr := crud.AppDb.Queryx(getQueryRes.SelectQueryObject.SelectQuery, getQueryRes.SelectQueryObject.FieldValues...)
	if qRowErr != nil {
		return mcresponse.GetResMessage("readError", mcresponse.ResponseMessageOptions{
			Message: fmt.Sprintf("Db query Error: %v", qRowErr.Error()),
			Value:   nil,
		})
	}
	defer func(rows *sqlx.Rows) {
		err := rows.Close()
		if err != nil {

		}
	}(rows)
	// check rows count
	//var rowCount = 0
	var getRecords []map[string]interface{}
	for rows.Next() {
		mapRes := map[string]interface{}{}
		rowScanErr := rows.MapScan(mapRes)
		if rowScanErr != nil {
			return mcresponse.GetResMessage("readError", mcresponse.ResponseMessageOptions{
				Message: fmt.Sprintf("Error reading/getting records[row-scan-map]: %v", rowScanErr.Error()),
				Value:   nil,
			})
		}
		fmt.Printf("row-map-res: %#v", mapRes)
		// transform snapshot value from model-struct to map-value
		jByte, jErr := json.Marshal(mapRes)
		if jErr != nil {
			return mcresponse.GetResMessage("paramsError", mcresponse.ResponseMessageOptions{
				Message: fmt.Sprintf("Error transforming result-value into json-value-format: %v", jErr.Error()),
				Value:   nil,
			})
		}
		mapValue := map[string]interface{}{}
		jErr = json.Unmarshal(jByte, &mapValue)
		if jErr != nil {
			return mcresponse.GetResMessage("paramsError", mcresponse.ResponseMessageOptions{
				Message: fmt.Sprintf("Error transforming result-value into json-value-format: %v", jErr.Error()),
				Value:   nil,
			})
		}
		//fmt.Printf("map-scanned-result: %v \n", mapValue)
		mapVal, mapErr := MapToMapCamelCase(mapValue, crud.FieldSeparator)
		if mapErr != nil {
			return mcresponse.GetResMessage("paramsError", mcresponse.ResponseMessageOptions{
				Message: fmt.Sprintf("map-underscore-to-camelCase-error: %v", mapErr.Error()),
				Value:   nil,
			})
		}
		fmt.Printf("map-transformed-result(camelCase): %v \n", mapVal)
		getRecords = append(getRecords, mapVal)
		//rowCount += 1
	}
	// handles not-found-error
	if len(getRecords) < 1 {
		return mcresponse.GetResMessage("notFound", mcresponse.ResponseMessageOptions{
			Message: "RECORDS NOT FOUND.",
			Value:   nil,
		})
	}
	if rowErr := rows.Err(); rowErr != nil {
		return mcresponse.GetResMessage("readError", mcresponse.ResponseMessageOptions{
			Message: fmt.Sprintf("Error reading/getting records: %v", rowErr.Error()),
			Value: GetResultType{
				Records:  nil,
				Stats:    GetStatType{},
				TaskType: crud.TaskType,
				LogRes:   mcresponse.ResponseMessage{},
			},
		})
	}
	// perform audit-log | initialize log-variables
	logMessage := ""
	logRes := mcresponse.ResponseMessage{}
	var logErr error
	if crud.LogRead || crud.LogCrud {
		logRecs := map[string]interface{}{"query": "all"}
		//jVal, _ := json.Marshal(logRecs)
		auditInfo := AuditLogOptionsType{
			TableName:  crud.TableName,
			LogRecords: LogRecordsType{LogRecords: logRecs},
		}
		if logRes, logErr = crud.TransLog.AuditLog(ReadTask, crud.UserInfo.UserId, auditInfo); logErr != nil {
			logMessage = fmt.Sprintf("Audit-log-error: %v", logErr.Error())
		} else {
			logMessage = fmt.Sprintf("Audit-log-code: %v | Message: %v", logRes.Code, logRes.Message)
		}
	}
	// result
	getResult := GetResultType{
		Records: getRecords,
		Stats: GetStatType{
			Skip:              crud.Skip,
			Limit:             crud.Limit,
			RecordsCount:      len(getRecords),
			TotalRecordsCount: totalRows,
			QueryParam:        crud.QueryParams,
			RecordIds:         crud.RecordIds,
		},
		TaskType: crud.TaskType,
		LogRes:   logRes,
	}
	// update cache | *****don't cache all-table-records, due to large/unknown size*****
	//_ = mccache.SetHashCache(crud.CacheKey, crud.TableName, getRecords, uint(crud.CacheExpire))

	return mcresponse.GetResMessage("success", mcresponse.ResponseMessageOptions{
		Message: fmt.Sprintf("AmountRecord(s) read-query completed successfully [log-message: %v]", logMessage),
		Value:   getResult,
	})
}
