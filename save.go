// @Author: abbeymart | Abi Akindele | @Created: 2020-12-01 | @Updated: 2020-12-01
// @Company: mConnect.biz | @License: MIT
// @Description: save (create / update) record(s)

package mcdbcrud

import (
	"fmt"
	"github.com/abbeymart/mccache"
	"github.com/abbeymart/mcresponse"
	"log"
)

// Create method creates new record(s)
func (crud *Crud) Create(recs ActionParamsType) mcresponse.ResponseMessage {
	// compute query
	createQueryRes := ComputeCreateQuery(crud.TableName, recs)
	if !createQueryRes.Ok {
		return mcresponse.GetResMessage("insertError", mcresponse.ResponseMessageOptions{
			Message: createQueryRes.Message,
			Value:   nil,
		})
	}
	//fmt.Printf("Query-info: %v \n", createQueryRes.CreateQueryObject.CreateQuery)
	//fmt.Printf("query-values: %v\n", createQueryRes.CreateQueryObject.FieldValues)
	// perform create/insert action, via transaction/copy-protocol:
	tx, txErr := crud.AppDb.Beginx()
	if txErr != nil {
		return mcresponse.GetResMessage("insertError", mcresponse.ResponseMessageOptions{
			Message: fmt.Sprintf("Error creating new record(s): %v", txErr.Error()),
			Value:   nil,
		})
	}
	// perform records' creation
	insertCount := 0
	var insertIds []string
	var insertId string
	// create new records by fieldValues
	for _, fValues := range createQueryRes.CreateQueryObject.FieldValues {
		insertErr := tx.QueryRowx(createQueryRes.CreateQueryObject.CreateQuery, fValues...).Scan(&insertId)
		if insertErr != nil {
			if rErr := tx.Rollback(); rErr != nil {
				log.Fatalf("Unable to Rollback: Check DB-driver: %v", rErr.Error())
			}
			return mcresponse.GetResMessage("insertError", mcresponse.ResponseMessageOptions{
				Message: fmt.Sprintf("Error creating new record(s): %v", insertErr.Error()),
				Value:   nil,
			})
		}
		insertCount += 1
		insertIds = append(insertIds, insertId)
	}
	// commit
	txcErr := tx.Commit()
	if txcErr != nil {
		if rErr := tx.Rollback(); rErr != nil {
			log.Fatalf("Unable to Rollback: Check DB-driver: %v", rErr.Error())
		}
		return mcresponse.GetResMessage("insertError", mcresponse.ResponseMessageOptions{
			Message: fmt.Sprintf("Error creating new record(s): %v", txcErr.Error()),
			Value:   nil,
		})
	}
	// delete cache
	_ = mccache.DeleteHashCache(crud.TableName, crud.CacheKey, "key")
	// perform audit-log
	logMessage := ""
	logRes := mcresponse.ResponseMessage{}
	var logErr error
	if crud.LogCreate {
		auditInfo := AuditLogOptionsType{
			TableName:  crud.TableName,
			LogRecords: LogRecordsType{LogRecords: crud.ActionParams},
		}
		if logRes, logErr = crud.TransLog.AuditLog(CreateTask, crud.UserInfo.UserId, auditInfo); logErr != nil {
			logMessage = fmt.Sprintf("Audit-log-error: %v", logErr.Error())
		} else {
			logMessage = fmt.Sprintf("Audit-log-code: %v | Message: %v", logRes.Code, logRes.Message)
		}
	}
	// response
	return mcresponse.GetResMessage("success", mcresponse.ResponseMessageOptions{
		Message: fmt.Sprintf("Record(s) creation completed successfully [log-message: %v]", logMessage),
		Value: CrudResultType{
			RecordIds:    insertIds,
			RecordsCount: insertCount,
			TaskType:     crud.TaskType,
			LogRes:       logRes,
		},
	})
}

// Update method updates existing record(s)
func (crud *Crud) Update(recs ActionParamsType) mcresponse.ResponseMessage {
	// include audit-log feature
	if crud.LogUpdate || crud.LogCrud {
		getRes := crud.GetByIds()
		if getRes.Code == "success" {
			value, _ := getRes.Value.(GetResultType)
			crud.CurrentRecords = value.Records
		}
	}
	// create from updatedRecs (actionParams)
	updateQueryRes := ComputeUpdateQuery(crud.TableName, recs)
	if !updateQueryRes.Ok {
		return mcresponse.GetResMessage("updateError", mcresponse.ResponseMessageOptions{
			Message: updateQueryRes.Message,
			Value:   nil,
		})
	}
	// perform update action, via transaction:
	tx, txErr := crud.AppDb.Begin()
	if txErr != nil {
		return mcresponse.GetResMessage("updateError", mcresponse.ResponseMessageOptions{
			Message: fmt.Sprintf("Error updating record(s): %v", txErr.Error()),
			Value:   nil,
		})
	}
	// perform records' updates
	updateCount := 0
	for _, upQuery := range updateQueryRes.UpdateQueryObjects {
		_, updateErr := tx.Exec(upQuery.UpdateQuery, upQuery.FieldValues...)
		if updateErr != nil {
			if rErr := tx.Rollback(); rErr != nil {
				log.Fatalf("Unable to Rollback: Check DB-driver: %v", rErr.Error())
			}
			return mcresponse.GetResMessage("updateError", mcresponse.ResponseMessageOptions{
				Message: fmt.Sprintf("Error updating record(s): %v", updateErr.Error()),
				Value:   nil,
			})
		}
		updateCount += 1
	}
	// commit
	txcErr := tx.Commit()
	if txcErr != nil {
		if rErr := tx.Rollback(); rErr != nil {
			log.Fatalf("Unable to Rollback: Check DB-driver: %v", rErr.Error())
		}
		return mcresponse.GetResMessage("updateError", mcresponse.ResponseMessageOptions{
			Message: fmt.Sprintf("Error updating record(s): %v", txcErr.Error()),
			Value:   nil,
		})
	}
	// delete cache
	_ = mccache.DeleteHashCache(crud.TableName, crud.CacheKey, "key")
	// perform audit-log
	logMessage := ""
	logRes := mcresponse.ResponseMessage{}
	var logErr error
	if crud.LogUpdate || crud.LogCrud {
		auditInfo := AuditLogOptionsType{
			TableName:     crud.TableName,
			LogRecords:    LogRecordsType{LogRecords: crud.CurrentRecords},
			NewLogRecords: LogRecordsType{LogRecords: crud.ActionParams},
		}
		if logRes, logErr = crud.TransLog.AuditLog(UpdateTask, crud.UserInfo.UserId, auditInfo); logErr != nil {
			logMessage = fmt.Sprintf("Audit-log-error: %v", logErr.Error())
		} else {
			logMessage = fmt.Sprintf("Audit-log-code: %v | Message: %v", logRes.Code, logRes.Message)
		}
	}
	// response
	return mcresponse.GetResMessage("success", mcresponse.ResponseMessageOptions{
		Message: fmt.Sprintf("Record(s) update completed successfully [log-message: %v]", logMessage),
		Value: CrudResultType{
			QueryParam:   crud.QueryParams,
			RecordIds:    crud.RecordIds,
			RecordsCount: updateCount,
			TaskType:     crud.TaskType,
			LogRes:       logRes,
		},
	})
}

// UpdateById method updates existing records (in batch) that met the specified record-id(s)
func (crud *Crud) UpdateById(rec ActionParamType, id string) mcresponse.ResponseMessage {
	// include audit-log feature
	if crud.LogUpdate || crud.LogCrud {
		getRes := crud.GetById(id)
		if getRes.Code == "success" {
			value, _ := getRes.Value.(GetResultType)
			crud.CurrentRecords = value.Records
		}
	}
	// create from updatedRecs (actionParams)
	updateQueryRes := ComputeUpdateQueryById(crud.TableName, rec, id)
	if !updateQueryRes.Ok {
		return mcresponse.GetResMessage("updateError", mcresponse.ResponseMessageOptions{
			Message: updateQueryRes.Message,
			Value:   nil,
		})
	}
	// perform update action, via transaction:
	tx, txErr := crud.AppDb.Begin()
	if txErr != nil {
		return mcresponse.GetResMessage("updateError", mcresponse.ResponseMessageOptions{
			Message: fmt.Sprintf("Error updating record(s): %v", txErr.Error()),
			Value:   nil,
		})
	}
	_, updateErr := tx.Exec(updateQueryRes.UpdateQueryObject.UpdateQuery, updateQueryRes.UpdateQueryObject.FieldValues...)
	if updateErr != nil {
		if rErr := tx.Rollback(); rErr != nil {
			log.Fatalf("Unable to Rollback: Check DB-driver: %v", rErr.Error())
		}
		return mcresponse.GetResMessage("updateError", mcresponse.ResponseMessageOptions{
			Message: fmt.Sprintf("Error updating record(s): %v", updateErr.Error()),
			Value:   nil,
		})
	}
	// commit
	txcErr := tx.Commit()
	if txcErr != nil {
		if rErr := tx.Rollback(); rErr != nil {
			log.Fatalf("Unable to Rollback: Check DB-driver: %v", rErr.Error())
		}
		return mcresponse.GetResMessage("updateError", mcresponse.ResponseMessageOptions{
			Message: fmt.Sprintf("Error updating record(s): %v", txcErr.Error()),
			Value:   nil,
		})
	}
	// delete cache
	_ = mccache.DeleteHashCache(crud.TableName, crud.CacheKey, "key")
	// perform audit-log
	logMessage := ""
	logRes := mcresponse.ResponseMessage{}
	var logErr error
	if crud.LogUpdate || crud.LogCrud {
		auditInfo := AuditLogOptionsType{
			TableName:     crud.TableName,
			LogRecords:    LogRecordsType{LogRecords: crud.CurrentRecords},
			NewLogRecords: LogRecordsType{LogRecords: crud.ActionParams, RecordIds: []string{id}},
		}
		if logRes, logErr = crud.TransLog.AuditLog(UpdateTask, crud.UserInfo.UserId, auditInfo); logErr != nil {
			logMessage = fmt.Sprintf("Audit-log-error: %v", logErr.Error())
		} else {
			logMessage = fmt.Sprintf("Audit-log-code: %v | Message: %v", logRes.Code, logRes.Message)
		}
	}
	// response
	rowsCount := 1
	return mcresponse.GetResMessage("success", mcresponse.ResponseMessageOptions{
		Message: fmt.Sprintf("Record(s) update completed successfully [log-message: %v]", logMessage),
		Value: CrudResultType{
			QueryParam:   crud.QueryParams,
			RecordIds:    crud.RecordIds,
			RecordsCount: rowsCount,
			TaskType:     crud.TaskType,
			LogRes:       logRes,
		},
	})
}

// UpdateByIds method updates existing records (in batch) that met the specified record-id(s)
func (crud *Crud) UpdateByIds(rec ActionParamType) mcresponse.ResponseMessage {
	// include audit-log feature
	if crud.LogUpdate || crud.LogCrud {
		getRes := crud.GetByIds()
		if getRes.Code == "success" {
			value, _ := getRes.Value.(GetResultType)
			crud.CurrentRecords = value.Records
		}
	}
	// create from updatedRecs (actionParams)
	updateQueryRes := ComputeUpdateQueryByIds(crud.TableName, rec, crud.RecordIds)
	if !updateQueryRes.Ok {
		return mcresponse.GetResMessage("updateError", mcresponse.ResponseMessageOptions{
			Message: updateQueryRes.Message,
			Value:   nil,
		})
	}
	// perform update action, via transaction:
	tx, txErr := crud.AppDb.Begin()
	if txErr != nil {
		return mcresponse.GetResMessage("updateError", mcresponse.ResponseMessageOptions{
			Message: fmt.Sprintf("Error updating record(s): %v", txErr.Error()),
			Value:   nil,
		})
	}
	updateCount := 0
	_, updateErr := tx.Exec(updateQueryRes.UpdateQueryObject.UpdateQuery, updateQueryRes.UpdateQueryObject.FieldValues...)
	if updateErr != nil {
		if rErr := tx.Rollback(); rErr != nil {
			log.Fatalf("Unable to Rollback: Check DB-driver: %v", rErr.Error())
		}
		return mcresponse.GetResMessage("updateError", mcresponse.ResponseMessageOptions{
			Message: fmt.Sprintf("Error updating record(s): %v", updateErr.Error()),
			Value:   nil,
		})
	}
	// commit
	txcErr := tx.Commit()
	if txcErr != nil {
		if rErr := tx.Rollback(); rErr != nil {
			log.Fatalf("Unable to Rollback: Check DB-driver: %v", rErr.Error())
		}
		return mcresponse.GetResMessage("updateError", mcresponse.ResponseMessageOptions{
			Message: fmt.Sprintf("Error updating record(s): %v", txcErr.Error()),
			Value:   nil,
		})
	}
	updateCount += len(crud.RecordIds)
	// TODO: review the RowsAffected option
	// rowsCount, rcErr := res.RowsAffected()
	//	if rcErr != nil {
	//		updateCount += len(crud.RecordIds)
	//	}
	// delete cache
	_ = mccache.DeleteHashCache(crud.TableName, crud.CacheKey, "key")
	// perform audit-log
	logMessage := ""
	logRes := mcresponse.ResponseMessage{}
	var logErr error
	if crud.LogUpdate || crud.LogCrud {
		auditInfo := AuditLogOptionsType{
			TableName:     crud.TableName,
			LogRecords:    LogRecordsType{LogRecords: crud.CurrentRecords},
			NewLogRecords: LogRecordsType{LogRecords: crud.ActionParams, RecordIds: crud.RecordIds},
		}
		if logRes, logErr = crud.TransLog.AuditLog(UpdateTask, crud.UserInfo.UserId, auditInfo); logErr != nil {
			logMessage = fmt.Sprintf("Audit-log-error: %v", logErr.Error())
		} else {
			logMessage = fmt.Sprintf("Audit-log-code: %v | Message: %v", logRes.Code, logRes.Message)
		}
	}
	// response
	return mcresponse.GetResMessage("success", mcresponse.ResponseMessageOptions{
		Message: fmt.Sprintf("Record(s) update completed successfully [log-message: %v]", logMessage),
		Value: CrudResultType{
			QueryParam:   crud.QueryParams,
			RecordIds:    crud.RecordIds,
			RecordsCount: updateCount,
			TaskType:     crud.TaskType,
			LogRes:       logRes,
		},
	})
}

// UpdateByParam method updates existing records (in batch) that met the specified query-params or where conditions
func (crud *Crud) UpdateByParam(rec ActionParamType) mcresponse.ResponseMessage {
	// include audit-log feature
	if crud.LogUpdate || crud.LogCrud {
		getRes := crud.GetByParam()
		if getRes.Code == "success" {
			value, _ := getRes.Value.(GetResultType)
			crud.CurrentRecords = value.Records
		}
	}
	// create from updatedRecs (actionParams)
	updateQueryRes := ComputeUpdateQueryByParam(crud.TableName, rec, crud.QueryParams)
	if !updateQueryRes.Ok {
		return mcresponse.GetResMessage("updateError", mcresponse.ResponseMessageOptions{
			Message: updateQueryRes.Message,
			Value:   nil,
		})
	}
	// perform update action, via transaction:
	tx, txErr := crud.AppDb.Begin()
	if txErr != nil {
		return mcresponse.GetResMessage("updateError", mcresponse.ResponseMessageOptions{
			Message: fmt.Sprintf("Error updating record(s): %v", txErr.Error()),
			Value:   nil,
		})
	}
	updateFieldValues := updateQueryRes.UpdateQueryObject.FieldValues
	res, updateErr := tx.Exec(updateQueryRes.UpdateQueryObject.UpdateQuery, updateFieldValues...)
	if updateErr != nil {
		if rErr := tx.Rollback(); rErr != nil {
			log.Fatalf("Unable to Rollback: Check DB-driver: %v", rErr.Error())
		}
		return mcresponse.GetResMessage("updateError", mcresponse.ResponseMessageOptions{
			Message: fmt.Sprintf("Error updating record(s): %v", updateErr.Error()),
			Value:   nil,
		})
	}
	// commit
	txcErr := tx.Commit()
	if txcErr != nil {
		if rErr := tx.Rollback(); rErr != nil {
			log.Fatalf("Unable to Rollback: Check DB-driver: %v", rErr.Error())
		}
		return mcresponse.GetResMessage("updateError", mcresponse.ResponseMessageOptions{
			Message: fmt.Sprintf("Error updating record(s): %v", txcErr.Error()),
			Value:   nil,
		})
	}
	// delete cache
	_ = mccache.DeleteHashCache(crud.TableName, crud.CacheKey, "key")
	// perform audit-log
	logMessage := ""
	logRes := mcresponse.ResponseMessage{}
	var logErr error
	if crud.LogUpdate || crud.LogCrud {
		auditInfo := AuditLogOptionsType{
			TableName:     crud.TableName,
			LogRecords:    LogRecordsType{LogRecords: crud.CurrentRecords},
			NewLogRecords: LogRecordsType{LogRecords: crud.ActionParams, QueryParam: crud.QueryParams},
		}
		if logRes, logErr = crud.TransLog.AuditLog(UpdateTask, crud.UserInfo.UserId, auditInfo); logErr != nil {
			logMessage = fmt.Sprintf("Audit-log-error: %v", logErr.Error())
		} else {
			logMessage = fmt.Sprintf("Audit-log-code: %v | Message: %v", logRes.Code, logRes.Message)
		}
	}
	// response | TODO: review res.RowsAffected impact
	rowsCount, rcErr := res.RowsAffected()
	if rcErr != nil {
		rowsCount = 0
	}
	return mcresponse.GetResMessage("success", mcresponse.ResponseMessageOptions{
		Message: fmt.Sprintf("Record(s) update completed successfully [log-message: %v]", logMessage),
		Value: CrudResultType{
			QueryParam:   crud.QueryParams,
			RecordIds:    crud.RecordIds,
			RecordsCount: int(rowsCount),
			TaskType:     crud.TaskType,
			LogRes:       logRes,
		},
	})
}
