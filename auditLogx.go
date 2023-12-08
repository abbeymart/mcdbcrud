// @Author: abbeymart | Abi Akindele | @Created: 2020-12-04 | @Updated: 2020-12-04
// @Company: mConnect.biz | @License: MIT
// @Description: go: mConnect

package mcdbcrud

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/abbeymart/mcresponse"
	"github.com/jmoiron/sqlx"
	"strings"
	"time"
)

// LogParamX interfaces / types
type LogParamX struct {
	AuditDb    *sqlx.DB
	AuditTable string
}

type AuditLogOptionsXType struct {
	AuditTable    string
	TableName     string
	LogRecords    interface{}
	NewLogRecords interface{}
	QueryParams   interface{}
}

type AuditLoggerx interface {
	AuditLog(logType, userId string, options AuditLogOptionsType) (mcresponse.ResponseMessage, error)
}
type CreateLoggerx interface {
	CreateLog(table string, logRecords interface{}, userId string) (mcresponse.ResponseMessage, error)
}
type UpdateLoggerx interface {
	UpdateLog(tableName string, logRecords interface{}, newLogRecords interface{}, userId string) (mcresponse.ResponseMessage, error)
}
type ReadLoggerx interface {
	ReadLog(tableName string, logRecords interface{}, userId string) (mcresponse.ResponseMessage, error)
}
type DeleteLoggerx interface {
	DeleteLog(tableName string, logRecords interface{}, userId string) (mcresponse.ResponseMessage, error)
}
type AccessLoggerx interface {
	LoginLog(logRecords interface{}, userId string, tableName string) (mcresponse.ResponseMessage, error)
	LogoutLog(logRecords interface{}, userId string, tableName string) (mcresponse.ResponseMessage, error)
}

//type AuditCrudLogger interface {
//	CreateLog(table string, LogRecords interface{}, userId string) (mcresponse.ResponseMessage, error)
//	UpdateLog(TableName string, LogRecords interface{}, NewLogRecords interface{}, userId string) (mcresponse.ResponseMessage, error)
//	ReadLog(TableName string, LogRecords interface{}, userId string) (mcresponse.ResponseMessage, error)
//	DeleteLog(TableName string, LogRecords interface{}, userId string) (mcresponse.ResponseMessage, error)
//	LoginLog(LogRecords interface{}, userId string, TableName string) (mcresponse.ResponseMessage, error)
//	LogoutLog(LogRecords interface{}, userId string, TableName string) (mcresponse.ResponseMessage, error)
//	AuditLog(LogType, userId string, options AuditLogOptionsType) (mcresponse.ResponseMessage, error)
//}

// constants

func NewAuditLogx(auditDb *sqlx.DB, auditTable string) LogParamX {
	result := LogParamX{}
	result.AuditDb = auditDb
	result.AuditTable = auditTable
	// default value
	if result.AuditTable == "" {
		result.AuditTable = "audits"
	}
	return result
}

// String() function implementation
func (log LogParamX) String() string {
	return fmt.Sprintf(`
	AuditLog DB: %v \n AudiLog Table ItemName: %v \n
	`,
		log.AuditDb,
		log.AuditTable)
}

func (log LogParamX) AuditLog(logType, userId string, options AuditLogOptionsType) (mcresponse.ResponseMessage, error) {
	// variables
	logType = strings.ToLower(logType)
	logBy := userId

	var (
		tableName     = options.TableName
		sqlScript     = ""
		logRecords    interface{}
		newLogRecords interface{}
		logAt         = time.Now()
		dbResult      sql.Result
		err           error
	)
	// log-cases
	logRecs, _ := json.Marshal(options.LogRecords)
	newLogRecs, _ := json.Marshal(options.NewLogRecords)
	logRecords = string(logRecs)
	newLogRecords = string(newLogRecs)
	switch logType {
	case CreateLog:
		// validate params
		var errorMessage = ""
		if tableName == "" {
			errorMessage = "Table or Collection name is required."
		}
		if logBy == "" {
			if errorMessage != "" {
				errorMessage = errorMessage + " | userId is required."
			} else {
				errorMessage = "userId is required."
			}
		}
		if logRecords == nil {
			if errorMessage != "" {
				errorMessage = errorMessage + " | Created record(s) information is required."
			} else {
				errorMessage = "Created record(s) information is required."
			}
		}
		if errorMessage != "" {
			return mcresponse.GetResMessage("paramsError",
				mcresponse.ResponseMessageOptions{
					Message: errorMessage,
					Value:   nil,
				}), errors.New(errorMessage)
		}
		// compose SQL-script
		sqlScript = fmt.Sprintf("INSERT INTO %v(table_name, log_records, log_type, log_by, log_at ) VALUES ($1, $2, $3, $4, $5)", log.AuditTable)
		// perform db-log-insert action
		dbResult, err = log.AuditDb.Exec(sqlScript, tableName, logRecords, logType, logBy, logAt)
	case UpdateLog:
		// validate params
		var errorMessage = ""
		if tableName == "" {
			errorMessage = "Table or Collection name is required."
		}
		if logBy == "" {
			if errorMessage != "" {
				errorMessage = errorMessage + " | userId is required."
			} else {
				errorMessage = "userId is required."
			}
		}
		if logRecords == nil {
			if errorMessage != "" {
				errorMessage = errorMessage + " | Updated record(s) information is required."
			} else {
				errorMessage = "Updated record(s) information is required."
			}
		}
		if newLogRecords == nil {
			if errorMessage != "" {
				errorMessage = errorMessage + " | New/Update record(s) information is required."
			} else {
				errorMessage = "New/Update record(s) information is required."
			}
		}
		if errorMessage != "" {
			return mcresponse.GetResMessage("paramsError",
				mcresponse.ResponseMessageOptions{
					Message: errorMessage,
					Value:   nil,
				}), errors.New(errorMessage)
		}
		// compose SQL-script
		sqlScript = fmt.Sprintf("INSERT INTO %v(table_name, log_records, new_log_records, log_type, log_by, log_at ) VALUES ($1, $2, $3, $4, $5, $6)", log.AuditTable)
		// perform db-log-insert action
		dbResult, err = log.AuditDb.Exec(sqlScript, tableName, logRecords, newLogRecords, logType, logBy, logAt)
	case GetLog, ReadLog:
		// validate params
		var errorMessage = ""
		if tableName == "" {
			errorMessage = "Table or Collection name is required."
		}
		if logBy == "" {
			if errorMessage != "" {
				errorMessage = errorMessage + " | userId is required."
			} else {
				errorMessage = "userId is required."
			}
		}
		if logRecords == nil {
			if errorMessage != "" {
				errorMessage = errorMessage + " | Read/Get Params/Keywords information is required."
			} else {
				errorMessage = "Read/Get Params/Keywords information is required."
			}
		}
		if errorMessage != "" {
			return mcresponse.GetResMessage("paramsError",
				mcresponse.ResponseMessageOptions{
					Message: errorMessage,
					Value:   nil,
				}), errors.New(errorMessage)
		}
		// compose SQL-script
		sqlScript = fmt.Sprintf("INSERT INTO %v(table_name, log_records, log_type, log_by, log_at ) VALUES ($1, $2, $3, $4, $5)", log.AuditTable)
		// perform db-log-insert action
		dbResult, err = log.AuditDb.Exec(sqlScript, tableName, logRecords, logType, logBy, logAt)
	case DeleteLog, RemoveLog:
		// validate params
		var errorMessage = ""
		if tableName == "" {
			errorMessage = "Table or Collection name is required."
		}
		if logBy == "" {
			if errorMessage != "" {
				errorMessage = errorMessage + " | userId is required."
			} else {
				errorMessage = "userId is required."
			}
		}
		if logRecords == nil {
			if errorMessage != "" {
				errorMessage = errorMessage + " | Deleted record(s) information is required."
			} else {
				errorMessage = "Deleted record(s) information is required."
			}
		}
		if errorMessage != "" {
			return mcresponse.GetResMessage("paramsError",
				mcresponse.ResponseMessageOptions{
					Message: errorMessage,
					Value:   nil,
				}), errors.New(errorMessage)
		}
		// compose SQL-script
		sqlScript = fmt.Sprintf("INSERT INTO %v(table_name, log_records, log_type, log_by, log_at ) VALUES ($1, $2, $3, $4, $5)", log.AuditTable)
		// perform db-log-insert action
		dbResult, err = log.AuditDb.Exec(sqlScript, tableName, logRecords, logType, logBy, logAt)
	case LoginLog:
		// validate params
		var errorMessage = ""
		if tableName == "" {
			errorMessage = "Table or Collection name is required."
		}
		if logBy == "" {
			if errorMessage != "" {
				errorMessage = errorMessage + " | userId is required."
			} else {
				errorMessage = "userId is required."
			}
		}
		if logRecords == nil {
			if errorMessage != "" {
				errorMessage = errorMessage + " | Login record(s) information is required."
			} else {
				errorMessage = "Login record(s) information is required."
			}
		}
		if errorMessage != "" {
			return mcresponse.GetResMessage("paramsError",
				mcresponse.ResponseMessageOptions{
					Message: errorMessage,
					Value:   nil,
				}), errors.New(errorMessage)
		}
		// compose SQL-script
		sqlScript = fmt.Sprintf("INSERT INTO %v(table_name, log_records, log_type, log_by, log_at ) VALUES ($1, $2, $3, $4, $5)", log.AuditTable)
		// perform db-log-insert action
		dbResult, err = log.AuditDb.Exec(sqlScript, tableName, logRecords, logType, logBy, logAt)
	case LogoutLog:
		// validate params
		var errorMessage = ""
		if tableName == "" {
			errorMessage = "Table or Collection name is required."
		}
		if logBy == "" {
			if errorMessage != "" {
				errorMessage = errorMessage + " | userId is required."
			} else {
				errorMessage = "userId is required."
			}
		}
		if logRecords == nil {
			if errorMessage != "" {
				errorMessage = errorMessage + " | Logout record(s) information is required."
			} else {
				errorMessage = "Logout record(s) information is required."
			}
		}
		if errorMessage != "" {
			return mcresponse.GetResMessage("paramsError",
				mcresponse.ResponseMessageOptions{
					Message: errorMessage,
					Value:   nil,
				}), errors.New(errorMessage)
		}
		// compose SQL-script
		sqlScript = fmt.Sprintf("INSERT INTO %v(table_name, log_records, log_type, log_by, log_at ) VALUES ($1, $2, $3, $4, $5)", log.AuditTable)
		// perform db-log-insert action
		dbResult, err = log.AuditDb.Exec(sqlScript, tableName, logRecords, logType, logBy, logAt)
	default:
		return mcresponse.GetResMessage("logError",
			mcresponse.ResponseMessageOptions{
				Message: "Unknown log type and/or incomplete log information",
				Value:   nil,
			}), errors.New("unknown log type and/or incomplete log information")
	}

	// Handle error
	if err != nil {
		errMsg := fmt.Sprintf("%v", err.Error())
		return mcresponse.GetResMessage("logError",
			mcresponse.ResponseMessageOptions{
				Message: errMsg,
				Value:   nil,
			}), errors.New(errMsg)
	}
	lastInsertId, _ := dbResult.LastInsertId()
	rowsAffected, _ := dbResult.RowsAffected()
	return mcresponse.GetResMessage("success",
		mcresponse.ResponseMessageOptions{
			Message: "successful audit-log action",
			Value:   map[string]interface{}{"lastInsertId": lastInsertId, "rowsAffected": rowsAffected},
		}), nil
}

func (log LogParamX) CustomLog(params AuditParamsType) (mcresponse.ResponseMessage, error) {
	// validate params
	var errorMessage = ""
	if params.LogBy == "" {
		errorMessage = "Log userId/name or owner required."
	}
	if params.LogRecords == nil {
		if errorMessage != "" {
			errorMessage = errorMessage + " | Data / information to be logged is required."
		} else {
			errorMessage = "Data / information to be logged is required."
		}
	}
	if errorMessage != "" {
		return mcresponse.GetResMessage("paramsError",
			mcresponse.ResponseMessageOptions{
				Message: errorMessage,
				Value:   nil,
			}), errors.New(errorMessage)
	}
	// log-cases
	logBy := params.LogBy
	tableName := "not-specified"
	if params.TableName != "" {
		tableName = params.TableName
	}
	logRecs, _ := json.Marshal(params.LogRecords)
	logRecords := string(logRecs)
	var newLogRecords interface{}
	if params.NewLogRecords != nil {
		newLogRecs, newLogRecErr := json.Marshal(params.NewLogRecords)
		if newLogRecErr == nil {
			newLogRecords = string(newLogRecs)
		}
	}
	logAt := time.Now()
	logType := params.LogType
	if logType == "" {
		logType = CreateLog
	}
	// compose SQL-script
	sqlScript := fmt.Sprintf("INSERT INTO %v(table_name, log_records, new_log_records, log_type, log_by, log_at ) VALUES ($1, $2, $3, $4, $5, $6)", log.AuditTable)
	// perform db-log-insert action
	dbResult, dbErr := log.AuditDb.Exec(sqlScript, tableName, logRecords, newLogRecords, logType, logBy, logAt)
	// Handle error
	if dbErr != nil {
		errMsg := fmt.Sprintf("%v", dbErr.Error())
		return mcresponse.GetResMessage("logError",
			mcresponse.ResponseMessageOptions{
				Message: errMsg,
				Value:   nil,
			}), errors.New(errMsg)
	}
	lastInsertId, _ := dbResult.LastInsertId()
	rowsAffected, _ := dbResult.RowsAffected()
	return mcresponse.GetResMessage("success",
		mcresponse.ResponseMessageOptions{
			Message: "successful audit-log action",
			Value:   map[string]interface{}{"lastInsertId": lastInsertId, "rowsAffected": rowsAffected},
		}), nil
}

func (log LogParamX) CreateLogx(table string, logRecords interface{}, userId string) (mcresponse.ResponseMessage, error) {

	return mcresponse.GetResMessage("success", mcresponse.ResponseMessageOptions{}), nil
}

func (log LogParamX) UpdateLogx(tableName string, logRecords interface{}, newLogRecords interface{}, userId string) (mcresponse.ResponseMessage, error) {

	return mcresponse.GetResMessage("success", mcresponse.ResponseMessageOptions{}), nil
}

func (log LogParamX) ReadLogx(tableName string, logRecords interface{}, userId string) (mcresponse.ResponseMessage, error) {

	return mcresponse.GetResMessage("success", mcresponse.ResponseMessageOptions{}), nil
}

func (log LogParamX) DeleteLogx(tableName string, logRecords interface{}, userId string) (mcresponse.ResponseMessage, error) {

	return mcresponse.GetResMessage("success", mcresponse.ResponseMessageOptions{}), nil
}

func (log LogParamX) LoginLogx(logRecords interface{}, userId string, tableName string) (mcresponse.ResponseMessage, error) {
	// default-values
	if tableName == "" {
		tableName = "users"
	}

	return mcresponse.GetResMessage("success", mcresponse.ResponseMessageOptions{}), nil
}

func (log LogParamX) LogoutLogx(logRecords interface{}, userId string, tableName string) (mcresponse.ResponseMessage, error) {
	// default-values
	if tableName == "" {
		tableName = "users"
	}

	return mcresponse.GetResMessage("success", mcresponse.ResponseMessageOptions{}), nil
}
