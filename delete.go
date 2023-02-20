// @Author: abbeymart | Abi Akindele | @Created: 2020-12-01 | @Updated: 2020-12-01
// @Company: mConnect.biz | @License: MIT
// @Description: delete or remove record(s)

package mcdbcrud

import (
	"fmt"
	"github.com/abbeymart/mccache"
	"github.com/abbeymart/mcresponse"
)

// DeleteById method deletes or removes record(s) by record-id(s)
func (crud *Crud) DeleteById(id string) mcresponse.ResponseMessage {
	// current record(s)
	getRes := crud.GetById(id)
	if getRes.Code == "success" {
		value, _ := getRes.Value.(GetResultType)
		crud.CurrentRecords = value.Records
	} else {
		return mcresponse.GetResMessage("notFound", mcresponse.ResponseMessageOptions{
			Message: "Record not found",
			Value:   nil,
		})
	}
	// compute delete query by record-id
	deleteQueryRes := ComputeDeleteQueryById(crud.TableName, id)
	if !deleteQueryRes.Ok {
		return mcresponse.GetResMessage("deleteError", mcresponse.ResponseMessageOptions{
			Message: deleteQueryRes.Message,
			Value:   nil,
		})
	}
	//fmt.Printf("Delete-query: %v", deleteQueryRes.DeleteQueryObject.DeleteQuery )
	res, delErr := crud.AppDb.Exec(deleteQueryRes.DeleteQueryObject.DeleteQuery, deleteQueryRes.DeleteQueryObject.FieldValues...)
	if delErr != nil {
		return mcresponse.GetResMessage("deleteError", mcresponse.ResponseMessageOptions{
			Message: fmt.Sprintf("Error deleting record(s): %v", delErr.Error()),
			Value:   nil,
		})
	}
	// delete cache
	_ = mccache.DeleteHashCache(crud.CacheKey, crud.TableName, "hash")
	// perform audit-log
	logMessage := ""
	logRes := mcresponse.ResponseMessage{}
	var logErr error
	if crud.LogDelete || crud.LogCrud {
		auditInfo := AuditLogOptionsType{
			TableName:  crud.TableName,
			LogRecords: LogRecordsType{LogRecords: crud.CurrentRecords, RecordIds: []string{id}},
		}
		if logRes, logErr = crud.TransLog.AuditLog(DeleteTask, crud.UserInfo.UserId, auditInfo); logErr != nil {
			logMessage = fmt.Sprintf("Audit-log-error: %v", logErr.Error())
		} else {
			logMessage = fmt.Sprintf("Audit-log-code: %v | Message: %v", logRes.Code, logRes.Message)
		}
	}
	rowsCount, rcErr := res.RowsAffected()
	if rcErr != nil {
		rowsCount = 0
	}
	return mcresponse.GetResMessage("success", mcresponse.ResponseMessageOptions{
		Message: fmt.Sprintf("Record(s) deleted successfully: [log-message: %v]", logMessage),
		Value: CrudResultType{
			QueryParam:   crud.QueryParams,
			RecordsCount: int(rowsCount),
			TaskType:     crud.TaskType,
			LogRes:       logRes,
		},
	})
}

// DeleteByIds method deletes or removes record(s) by record-id(s)
func (crud *Crud) DeleteByIds() mcresponse.ResponseMessage {
	// current record(s)
	getRes := crud.GetByIds()
	if getRes.Code == "success" {
		value, _ := getRes.Value.(GetResultType)
		crud.CurrentRecords = value.Records
	} else {
		return mcresponse.GetResMessage("notFound", mcresponse.ResponseMessageOptions{
			Message: "Record(s) not found",
			Value:   nil,
		})
	}
	// compute delete query by record-ids
	deleteQueryRes := ComputeDeleteQueryByIds(crud.TableName, crud.RecordIds)
	if !deleteQueryRes.Ok {
		return mcresponse.GetResMessage("deleteError", mcresponse.ResponseMessageOptions{
			Message: deleteQueryRes.Message,
			Value:   nil,
		})
	}
	res, delErr := crud.AppDb.Exec(deleteQueryRes.DeleteQueryObject.DeleteQuery, deleteQueryRes.DeleteQueryObject.FieldValues...)
	if delErr != nil {
		return mcresponse.GetResMessage("deleteError", mcresponse.ResponseMessageOptions{
			Message: fmt.Sprintf("Error deleting record(s): %v", delErr.Error()),
			Value:   nil,
		})
	}
	// delete cache
	_ = mccache.DeleteHashCache(crud.CacheKey, crud.TableName, "hash")
	// perform audit-log
	logMessage := ""
	logRes := mcresponse.ResponseMessage{}
	var logErr error
	if crud.LogDelete || crud.LogCrud {
		auditInfo := AuditLogOptionsType{
			TableName:  crud.TableName,
			LogRecords: LogRecordsType{LogRecords: crud.CurrentRecords, RecordIds: crud.RecordIds},
		}
		if logRes, logErr = crud.TransLog.AuditLog(DeleteTask, crud.UserInfo.UserId, auditInfo); logErr != nil {
			logMessage = fmt.Sprintf("Audit-log-error: %v", logErr.Error())
		} else {
			logMessage = fmt.Sprintf("Audit-log-code: %v | Message: %v", logRes.Code, logRes.Message)
		}
	}
	rowsCount, rcErr := res.RowsAffected()
	if rcErr != nil {
		rowsCount = 0
	}
	return mcresponse.GetResMessage("success", mcresponse.ResponseMessageOptions{
		Message: fmt.Sprintf("Record(s) deleted successfully: [log-message: %v]", logMessage),
		Value: CrudResultType{
			QueryParam:   crud.QueryParams,
			RecordsCount: int(rowsCount),
			TaskType:     crud.TaskType,
			LogRes:       logRes,
		},
	})
}

// DeleteByParam method deletes or removes record(s) by query-parameters or where conditions
func (crud *Crud) DeleteByParam() mcresponse.ResponseMessage {
	// current record(s)
	getRes := crud.GetByParam()
	if getRes.Code == "success" {
		value, _ := getRes.Value.(GetResultType)
		crud.CurrentRecords = value.Records
	} else {
		return mcresponse.GetResMessage("notFound", mcresponse.ResponseMessageOptions{
			Message: "Record(s) not found",
			Value:   nil,
		})
	}
	// compute delete query by query-params
	deleteQueryRes := ComputeDeleteQueryByParam(crud.TableName, crud.QueryParams)
	//fmt.Printf("delete-by-param-query: %v \n", deleteQueryRes.DeleteQueryObject.DeleteQuery)
	if !deleteQueryRes.Ok {
		return mcresponse.GetResMessage("deleteError", mcresponse.ResponseMessageOptions{
			Message: deleteQueryRes.Message,
			Value:   nil,
		})
	}
	res, delErr := crud.AppDb.Exec(deleteQueryRes.DeleteQueryObject.DeleteQuery, deleteQueryRes.DeleteQueryObject.FieldValues...)
	if delErr != nil {
		return mcresponse.GetResMessage("deleteError", mcresponse.ResponseMessageOptions{
			Message: fmt.Sprintf("Error deleting record(s): %v", delErr.Error()),
			Value:   nil,
		})
	}
	// delete cache
	_ = mccache.DeleteHashCache(crud.CacheKey, crud.TableName, "hash")
	// perform audit-log
	logMessage := ""
	logRes := mcresponse.ResponseMessage{}
	var logErr error
	if crud.LogDelete || crud.LogCrud {
		auditInfo := AuditLogOptionsType{
			TableName:  crud.TableName,
			LogRecords: LogRecordsType{LogRecords: crud.CurrentRecords, QueryParam: crud.QueryParams},
		}
		if logRes, logErr = crud.TransLog.AuditLog(DeleteTask, crud.UserInfo.UserId, auditInfo); logErr != nil {
			logMessage = fmt.Sprintf("Audit-log-error: %v", logErr.Error())
		} else {
			logMessage = fmt.Sprintf("Audit-log-code: %v | Message: %v", logRes.Code, logRes.Message)
		}
	}
	rowsCount, rcErr := res.RowsAffected()
	if rcErr != nil {
		rowsCount = 0
	}
	return mcresponse.GetResMessage("success", mcresponse.ResponseMessageOptions{
		Message: fmt.Sprintf("Record(s) deleted successfully: [log-message: %v]", logMessage),
		Value: CrudResultType{
			QueryParam:   crud.QueryParams,
			RecordsCount: int(rowsCount),
			TaskType:     DeleteTask,
			LogRes:       logRes,
		},
	})
}

// DeleteAll method deletes or removes all records in the tables. Recommended for admin-users only
// Use if and only if you know what you are doing
func (crud *Crud) DeleteAll() mcresponse.ResponseMessage {
	// ***** perform DELETE-ALL-RECORDS FROM A TABLE, IF RELATIONS/CONSTRAINTS PERMIT *****
	// ***** && IF-AND-ONLY-IF-YOU-KNOW-WHAT-YOU-ARE-DOING && AT-YOUR-OWN-RISK *****
	// compute delete query
	delQuery := fmt.Sprintf("DELETE FROM %v", crud.TableName)
	res, delErr := crud.AppDb.Exec(delQuery)
	if delErr != nil {
		return mcresponse.GetResMessage("deleteError", mcresponse.ResponseMessageOptions{
			Message: fmt.Sprintf("Error deleting record(s): %v", delErr.Error()),
			Value:   nil,
		})
	}
	// delete cache, by key (TableName)
	_ = mccache.DeleteHashCache(crud.CacheKey, crud.TableName, "hash")
	// perform audit-log
	logMessage := ""
	logRes := mcresponse.ResponseMessage{}
	var logErr error
	if crud.LogDelete || crud.LogCrud {
		currentRecs := map[string]interface{}{"query": "all"}
		auditInfo := AuditLogOptionsType{
			TableName:  crud.TableName,
			LogRecords: LogRecordsType{QueryParam: currentRecs},
		}
		if logRes, logErr = crud.TransLog.AuditLog(DeleteTask, crud.UserInfo.UserId, auditInfo); logErr != nil {
			logMessage = fmt.Sprintf("Audit-log-error: %v", logErr.Error())
		} else {
			logMessage = fmt.Sprintf("Audit-log-code: %v | Message: %v", logRes.Code, logRes.Message)
		}
	}
	// response
	rowsCount, rcErr := res.RowsAffected()
	if rcErr != nil {
		rowsCount = 0
	}
	return mcresponse.GetResMessage("success", mcresponse.ResponseMessageOptions{
		Message: fmt.Sprintf("Record(s) deleted successfully [log-message: %v] ", logMessage),
		Value: CrudResultType{
			RecordsCount: int(rowsCount),
			TaskType:     DeleteTask,
			LogRes:       logRes,
		},
	})
}
