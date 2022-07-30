// @Author: abbeymart | Abi Akindele | @Created: 2020-12-01 | @Updated: 2020-12-01
// @Company: mConnect.biz | @License: MIT
// @Description: Base type/method CRUD operations for PgDB

package mcdbcrud

import (
	"encoding/json"
	"fmt"
	"github.com/abbeymart/mcresponse"
	"time"
)

// Crud object / struct
type Crud struct {
	CrudParamsType
	CrudOptionsType
	CreateItems    ActionParamsType
	UpdateItems    ActionParamsType
	CurrentRecords []map[string]interface{}
	TransLog       LogParamX
	CacheKey       string // Unique for exactly the same query
}

// NewCrud constructor returns a new crud-instance
func NewCrud(params CrudParamsType, options CrudOptionsType) (crudInstance *Crud) {
	crudInstance = &Crud{}
	// compute crud params
	crudInstance.ModelRef = params.ModelRef
	crudInstance.ModelPointer = params.ModelPointer
	crudInstance.AppDb = params.AppDb
	crudInstance.TableName = params.TableName
	crudInstance.UserInfo = params.UserInfo
	crudInstance.ActionParams = params.ActionParams
	crudInstance.RecordIds = params.RecordIds
	crudInstance.QueryParams = params.QueryParams
	crudInstance.SortParams = params.SortParams
	crudInstance.ProjectParams = params.ProjectParams
	crudInstance.Token = params.Token
	crudInstance.TaskName = params.TaskName
	crudInstance.Skip = params.Skip
	crudInstance.Limit = params.Limit
	crudInstance.AppParams = params.AppParams

	// crud options
	crudInstance.MaxQueryLimit = options.MaxQueryLimit
	crudInstance.AuditTable = options.AuditTable
	crudInstance.AccessTable = options.AccessTable
	crudInstance.RoleTable = options.RoleTable
	crudInstance.UserTable = options.UserTable
	crudInstance.VerifyTable = options.VerifyTable
	crudInstance.ProfileTable = options.ProfileTable
	crudInstance.ServiceTable = options.ServiceTable
	crudInstance.UserRoleTable = options.UserRoleTable
	crudInstance.AuditDb = options.AuditDb
	crudInstance.AccessDb = options.AccessDb
	crudInstance.LogCrud = options.LogCrud
	crudInstance.LogRead = options.LogRead
	crudInstance.LogCreate = options.LogCreate
	crudInstance.LogUpdate = options.LogUpdate
	crudInstance.LogDelete = options.LogDelete
	crudInstance.CheckAccess = options.CheckAccess // Dec 09/2020: user to implement auth as a middleware
	crudInstance.CacheResult = options.CacheResult
	crudInstance.CacheExpire = options.CacheExpire // cache expire in secs
	crudInstance.BulkCreate = options.BulkCreate
	crudInstance.ModelOptions = options.ModelOptions
	crudInstance.FieldSeparator = options.FieldSeparator
	crudInstance.AppDbs = options.AppDbs
	crudInstance.AppTables = options.AppTables
	crudInstance.QueryFieldType = options.QueryFieldType

	// Default values
	if crudInstance.QueryFieldType == "" {
		crudInstance.QueryFieldType = CrudQueryFieldDefault
	}
	if crudInstance.AppDbs == nil {
		crudInstance.AppDbs = []string{"database", "database-mcpa", "database-mcpay", "database-mcship", "database-mctrade", "database-mcproperty",
			"database-mcinfo", "database-mcbc"}
	}
	if crudInstance.AppTables == nil {
		crudInstance.AppTables = []string{"table", "table-mcpa", "table-mcpay", "table-mcship", "table-mctrade", "table-mcproperty",
			"table-mcinfo", "table-mcbc"}
	}
	if crudInstance.FieldSeparator == "" {
		crudInstance.FieldSeparator = "_"
	}
	if crudInstance.AuditTable == "" {
		crudInstance.AuditTable = "audits"
	}
	if crudInstance.AccessTable == "" {
		crudInstance.AccessTable = "accesses"
	}
	if crudInstance.RoleTable == "" {
		crudInstance.RoleTable = "roles"
	}
	if crudInstance.UserTable == "" {
		crudInstance.UserTable = "users"
	}
	if crudInstance.VerifyTable == "" {
		crudInstance.VerifyTable = "verify_users"
	}
	if crudInstance.ProfileTable == "" {
		crudInstance.ProfileTable = "profiles"
	}
	if crudInstance.ServiceTable == "" {
		crudInstance.ServiceTable = "services"
	}
	if crudInstance.AuditDb == nil {
		crudInstance.AuditDb = crudInstance.AppDb
	}
	if crudInstance.AccessDb == nil {
		crudInstance.AccessDb = crudInstance.AppDb
	}
	if crudInstance.Skip < 0 {
		crudInstance.Skip = 0
	}

	if crudInstance.MaxQueryLimit <= 0 {
		crudInstance.MaxQueryLimit = 10000
	}

	if crudInstance.Limit <= 0 || crudInstance.Limit > crudInstance.MaxQueryLimit {
		crudInstance.Limit = crudInstance.MaxQueryLimit
	}

	if crudInstance.CacheExpire <= 0 {
		crudInstance.CacheExpire = 300 // 300 secs, 5 minutes
	}
	// Compute CacheKey from TableName, QueryParams, SortParams, ProjectParams and RecordIds
	qParam, _ := json.Marshal(params.QueryParams)
	sParam, _ := json.Marshal(params.SortParams)
	pParam, _ := json.Marshal(params.ProjectParams)
	dIds, _ := json.Marshal(params.RecordIds)
	//crudInstance.CacheKey = params.TableName + string(qParam) + string(sParam) + string(pParam) + string(dIds)
	crudInstance.CacheKey = fmt.Sprintf("%v-%v-%v-%v-%v-%v-%v", params.TableName, string(qParam), string(sParam), string(pParam), string(dIds), crudInstance.Skip, crudInstance.Limit)

	// Audit/TransLog instance
	crudInstance.TransLog = NewAuditLogx(crudInstance.AuditDb, crudInstance.AuditTable)

	return crudInstance
}

// String() method implementation for crud instance/object
func (crud *Crud) String() string {
	return fmt.Sprintf("CRUD Instance Information: %#v \n\n", crud)
}

// Methods

// SaveRecord method creates new record(s) or updates existing record(s)
func (crud *Crud) SaveRecord() mcresponse.ResponseMessage {
	//  compute taskType-records from actionParams: create or update
	var (
		createRecs = ActionParamsType{} // records without id field-value
		updateRecs = ActionParamsType{} // records with id field-value
	)
	// cases - actionParams.length === 1 record OR > 1 records
	if len(crud.ActionParams) == 1 {
		rec := crud.ActionParams[0]
		// determine if record exists (update, cast id into string) or new (create)
		recIdStr := ""
		idOk := false
		recId, ok := rec["id"]
		if ok {
			recIdStr, idOk = recId.(string)
		}
		// exclude id from record, if present
		mapRec := ExcludeFieldFromMapRecord(rec, "id")
		if len(crud.RecordIds) > 0 || len(crud.QueryParams) > 0 {
			if crud.ModelOptions.ActorStamp {
				mapRec["updatedBy"] = crud.UserInfo.UserId
			}
			if crud.ModelOptions.TimeStamp {
				mapRec["updatedAt"] = time.Now()
			}
			updateRecs = append(updateRecs, mapRec)
		} else if idOk && recIdStr != "" {
			// reset recordIds and query-params for update-task
			crud.RecordIds = []string{}
			crud.QueryParams = QueryParamType{}
			if crud.ModelOptions.ActorStamp {
				mapRec["updatedBy"] = crud.UserInfo.UserId
			}
			if crud.ModelOptions.TimeStamp {
				mapRec["updatedAt"] = time.Now()
			}
			crud.RecordIds = append(crud.RecordIds, recIdStr)
			updateRecs = append(updateRecs, mapRec)
		} else {
			// reset recordIds and query-params for create-task
			crud.RecordIds = []string{}
			crud.QueryParams = QueryParamType{}
			if crud.ModelOptions.ActorStamp {
				mapRec["createdBy"] = crud.UserInfo.UserId
			}
			if crud.ModelOptions.TimeStamp {
				mapRec["createdAt"] = time.Now()
			}
			createRecs = append(createRecs, mapRec)
		}
		crud.CreateItems = createRecs
		crud.UpdateItems = updateRecs
	} else if len(crud.ActionParams) > 1 {
		// reset recordIds and query-params for multiple create-update-records task
		crud.RecordIds = []string{}
		crud.QueryParams = QueryParamType{}
		//var recIds []string // capture recordIds for separate/multiple updates
		for _, rec := range crud.ActionParams {
			// determine if record exists (update), cast id into string or new (create)
			recIdStr := ""
			idOk := false
			recId, ok := rec["id"]
			if ok {
				recIdStr, idOk = recId.(string)
			}
			if idOk && recIdStr != "" {
				if crud.ModelOptions.ActorStamp {
					rec["updatedBy"] = crud.UserInfo.UserId
				}
				if crud.ModelOptions.TimeStamp {
					rec["updatedAt"] = time.Now()
				}
				crud.RecordIds = append(crud.RecordIds, recIdStr)
				updateRecs = append(updateRecs, rec)
			} else {
				// exclude id from record, if present
				mapRec := ExcludeFieldFromMapRecord(rec, "id")
				if crud.ModelOptions.ActorStamp {
					mapRec["createdBy"] = crud.UserInfo.UserId
				}
				if crud.ModelOptions.TimeStamp {
					mapRec["createdAt"] = time.Now()
				}
				createRecs = append(createRecs, mapRec)
			}
		}
		//crud.RecordIds = recIds
		crud.CreateItems = createRecs
		crud.UpdateItems = updateRecs
	}
	// validate and set task-type, create or update
	if len(createRecs) > 0 && len(updateRecs) > 0 {
		return mcresponse.GetResMessage("paramsError", mcresponse.ResponseMessageOptions{
			Message: "You may only create or update record(s), not both at the same time",
			Value:   nil,
		})
	}
	// set task-type
	if len(createRecs) > 0 {
		crud.TaskType = CreateTask
	} else if len(updateRecs) > 0 {
		crud.TaskType = UpdateTask
	} else {
		return mcresponse.GetResMessage("paramsError", mcresponse.ResponseMessageOptions{
			Message: "Inputs errors: actionParams required to complete create or update task.",
			Value:   nil,
		})
	}

	// create/insert new record(s)
	if crud.TaskType == CreateTask && len(createRecs) > 0 {
		// check task-permission
		if crud.CheckAccess {
			accessRes := crud.CheckTaskAccess()
			if accessRes.Code != "success" {
				return accessRes
			}
		}
		return crud.Create(createRecs)
	}

	// update existing record(s), by record-id(s) or queryParams | or perform multiple updates
	if crud.TaskType == UpdateTask {
		if len(updateRecs) == 1 {
			if len(crud.RecordIds) == 1 {
				// check task-permission
				if crud.CheckAccess {
					accessRes := crud.TaskPermissionById(crud.TaskType)
					if accessRes.Code != "success" {
						return accessRes
					}
				}
				return crud.UpdateById(updateRecs[0], crud.RecordIds[0])
			}
			if len(crud.RecordIds) > 1 {
				// check task-permission
				if crud.CheckAccess {
					accessRes := crud.TaskPermissionById(crud.TaskType)
					if accessRes.Code != "success" {
						return accessRes
					}
				}
				return crud.UpdateByIds(updateRecs[0])
			}
			if len(crud.QueryParams) > 0 {
				// check task-permission
				if crud.CheckAccess {
					accessRes := crud.TaskPermissionByParam(crud.TaskType)
					if accessRes.Code != "success" {
						return accessRes
					}
				}
				return crud.UpdateByParam(updateRecs[0])
			}
		}
		// check task-permission
		if crud.CheckAccess {
			accessRes := crud.TaskPermissionById(crud.TaskType)
			if accessRes.Code != "success" {
				return accessRes
			}
		}
		return crud.Update(updateRecs)
	}
	// otherwise, return saveError
	return mcresponse.GetResMessage("saveError", mcresponse.ResponseMessageOptions{
		Message: "Save error: incomplete or invalid parameters (action/query-params/record-ids) provided",
		Value:   nil,
	})
}

// DeleteRecord method deletes/removes record(s) by recordIds or queryParams
func (crud *Crud) DeleteRecord() mcresponse.ResponseMessage {
	if len(crud.RecordIds) == 1 {
		if crud.CheckAccess {
			accessRes := crud.TaskPermissionById(DeleteTask)
			if accessRes.Code != "success" {
				return accessRes
			}
		}
		return crud.DeleteById(crud.RecordIds[0])
	}
	if len(crud.RecordIds) > 1 {
		if crud.CheckAccess {
			accessRes := crud.TaskPermissionById(DeleteTask)
			if accessRes.Code != "success" {
				return accessRes
			}
		}
		return crud.DeleteByIds()
	}
	if crud.QueryParams != nil && len(crud.QueryParams) > 0 {
		if crud.CheckAccess {
			accessRes := crud.TaskPermissionByParam(DeleteTask)
			if accessRes.Code != "success" {
				return accessRes
			}
		}
		return crud.DeleteByParam()
	}
	// delete-all ***RESTRICTED***
	// otherwise return error
	return mcresponse.GetResMessage("removeError", mcresponse.ResponseMessageOptions{
		Message: "You may delete records by recordIds or queryParams only.",
		Value:   nil,
	})
}

// GetRecord method fetches records by recordIds, queryParams or all
func (crud *Crud) GetRecord() mcresponse.ResponseMessage {
	if len(crud.RecordIds) == 1 {
		if crud.CheckAccess {
			accessRes := crud.TaskPermissionById(ReadTask)
			if accessRes.Code != "success" {
				return accessRes
			}
		}
		return crud.GetById(crud.RecordIds[0])
	}
	if len(crud.RecordIds) > 1 {
		if crud.CheckAccess {
			accessRes := crud.TaskPermissionById(ReadTask)
			if accessRes.Code != "success" {
				return accessRes
			}
		}
		return crud.GetByIds()
	}
	if crud.QueryParams != nil && len(crud.QueryParams) > 0 {
		if crud.CheckAccess {
			accessRes := crud.TaskPermissionByParam(ReadTask)
			if accessRes.Code != "success" {
				return accessRes
			}
		}
		return crud.GetByParam()
	}
	// get-task for admin or owner
	accessRes := crud.CheckUserAccess()
	if accessRes.Code != "success" {
		return accessRes
	}
	accessRec, ok := accessRes.Value.(AccessInfoType)
	// admin
	if ok && accessRec.IsActive && accessRec.IsAdmin {
		return crud.GetAll()
	}
	// owner
	if ok && accessRec.IsActive && accessRec.UserId != "" {
		crud.QueryParams = map[string]interface{}{"createdBy": accessRec.UserId}
		return crud.GetByParam()
	}
	// not-found-error-message
	return mcresponse.GetResMessage("notFound", mcresponse.ResponseMessageOptions{
		Message: "Records not found - ensure you have provided the correct query-parameters",
		Value:   nil,
	})
}

// GetRecords method fetches records by recordIds, queryParams or all - lookup-items (no-access-constraint)
func (crud *Crud) GetRecords() mcresponse.ResponseMessage {
	if len(crud.RecordIds) == 1 {
		return crud.GetById(crud.RecordIds[0])
	}
	if len(crud.RecordIds) > 1 {
		return crud.GetByIds()
	}
	if crud.QueryParams != nil && len(crud.QueryParams) > 0 {
		return crud.GetByParam()
	}
	return crud.GetAll()
}
