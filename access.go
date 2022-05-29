// @Author: abbeymart | Abi Akindele | @Created: 2020-12-08 | @Updated: 2020-12-08
// @Company: mConnect.biz | @License: MIT
// @Description: optional access methods, to be used as middleware, prior to CRUD operation

package mcdbcrud

import (
	"errors"
	"fmt"
	"github.com/abbeymart/mcresponse"
	"github.com/jmoiron/sqlx"
	"strings"
	"time"
)

// AccessInfoType for CheckUserAccess method value (interface{}) response,
// and to assert returned value
type AccessInfoType struct {
	UserId   string   `json:"userId"`
	RoleId   string   `json:"roleId"`
	RoleIds  []string `json:"roleIds"`
	IsAdmin  bool     `json:"isAdmin"`
	IsActive bool     `json:"isActive"`
}

// TaskPermissionType for TaskPermission method value (interface{}) response,
// and to assert returned value
type TaskPermissionType struct {
	Ok             bool     `json:"ok"`
	IsAdmin        bool     `json:"isAdmin"`
	IsActive       bool     `json:"isActive"`
	UserId         string   `json:"userId"`
	RoleId         string   `json:"roleId"`
	RoleIds        []string `json:"roleIds"`
	OwnerPermitted bool     `json:"ownerPermitted"`
}

// TODO: extend/refactor access-control to subscribers/apps/services groups/categories
// user can operate on owned records (crud)
// admin-user/System-Admin/User-Admin? can perform all tasks
// tasks may be performed by role-assignment
// subscriber-admin can grant role-assignment (CRUD - record/table) to owned app-data (i.e. by appId)

// RecordsCount returns the totalRecordsCount, ownerRecordsCount and error, if applicable
func (crud *Crud) RecordsCount() (totalRecords int, ownerRecords int, err error) {
	// totalRecordsCount from the table
	countQuery := fmt.Sprintf("SELECT COUNT(*) AS total_records FROM %v", crud.TableName)
	tRowErr := crud.AppDb.QueryRowx(countQuery).Scan(&totalRecords)
	if tRowErr != nil {
		return 0, 0, errors.New(fmt.Sprintf("Db query Error[total-records-count]: %v", tRowErr.Error()))
	}
	// count owner-records
	sqlScript := fmt.Sprintf("SELECT COUNT(*) AS owner_records FROM %v WHERE created_by = $1", crud.TableName)
	uRowErr := crud.AppDb.QueryRowx(sqlScript, crud.UserInfo.UserId).Scan(&ownerRecords)
	if uRowErr != nil {
		return 0, 0, errors.New(fmt.Sprintf("Db query Error[total-records-count]: %v", uRowErr.Error()))
	}
	return totalRecords, ownerRecords, nil
}

// CheckSaveTaskType determines the crud-task type based on actionParams[0] record (i.e. first record in the array)
func (crud *Crud) CheckSaveTaskType() string {
	taskType := ""
	if len(crud.ActionParams) > 0 {
		actParam := crud.ActionParams[0]
		_, ok := actParam["id"]
		if !ok {
			if len(crud.ActionParams) == 1 && (len(crud.RecordIds) > 0 || len(crud.QueryParams) > 0) {
				taskType = UpdateTask
			} else {
				taskType = CreateTask
			}
		} else {
			taskType = UpdateTask
		}
	}
	return taskType
}

// CheckTaskAccess method determines the access by role-assignment
func (crud *Crud) CheckTaskAccess() mcresponse.ResponseMessage {
	// validate current user active status: by token (API) and user/loggedIn-status
	accessRes := crud.CheckUserAccess()
	if accessRes.Code != "success" {
		return accessRes
	}
	// set current-user info for next steps
	var (
		userId   string
		roleId   string
		roleIds  []string
		isAdmin  bool
		isActive bool
	)
	val, ok := accessRes.Value.(AccessInfoType)
	if !ok {
		return mcresponse.GetResMessage("unAuthorized", mcresponse.ResponseMessageOptions{
			Message: "Error parsing user access information/value",
			Value:   nil,
		})
	}
	userId = val.UserId
	roleId = val.RoleId
	roleIds = val.RoleIds
	isAdmin = val.IsAdmin
	isActive = val.IsActive
	// determine records/documents ownership, for all records (atomic)
	ownerPermitted := false
	idLen := len(crud.RecordIds)
	if userId != "" && isActive {
		if idLen > 0 {
			// SQL script
			inValues := ""
			for idCount, id := range crud.RecordIds {
				inValues += "'" + id + "'"
				if idLen > 1 && idCount < idLen-1 {
					inValues += ", "
				}
			}
			var ownerRecords int
			sqlScript := fmt.Sprintf("SELECT COUNT(*) as ownerrecords FROM %v WHERE id IN (%v) AND created_by = $1", crud.TableName, inValues)
			rErr := crud.AppDb.QueryRowx(sqlScript, userId).Scan(&ownerRecords)
			if rErr != nil {
				ownerRecords = 0
			}
			// ensure complete records count, as requested
			if ownerRecords == len(crud.RecordIds) {
				ownerPermitted = true
			}
		} else {
			// totalRecordsCount from the table
			_, ownerRecords, recErr := crud.RecordsCount()
			if recErr != nil {
				ownerRecords = 0
			}
			// check rows count for
			switch crud.TaskType {
			case ReadTask:
				if ownerRecords > 0 {
					ownerPermitted = true
				}
			default:
				ownerPermitted = false
			}
		}
	}
	// TODO: for testing only createTask-permission, remove after-initial-testing (grant role via role/user-roles)
	if !val.IsAdmin && crud.TaskType == CreateTask && crud.TableName != crud.UserTable {
		ownerPermitted = true
	}
	// if all the above checks passed, check for role-services access by taskType
	// obtain table/collName id(_id) from serviceTable/Coll (repo for all resources)
	var (
		serviceId string
		category  string
	)
	serviceScript := fmt.Sprintf("SELECT id, category from %v WHERE name=$1", crud.ServiceTable)
	serviceRow := crud.AccessDb.QueryRow(serviceScript, crud.TableName)
	// check row-scan-error
	sErr := serviceRow.Scan(&serviceId, &category)
	if sErr != nil {
		serviceId = ""
		category = ""
	}
	// if permitted, include table/collId and recordIds in serviceIds
	tableId := ""
	serviceIds := crud.RecordIds
	catLowercase := strings.ToLower(category)
	if serviceId != "" && (catLowercase == "table" || catLowercase == "collection" || ArrayStringContains(crud.AppTables, catLowercase)) {
		tableId = serviceId
		serviceIds = append(serviceIds, serviceId)
	}
	// compute service-items/records
	var roleServices []RoleServiceType
	var rsErr error
	if len(serviceIds) > 0 {
		roleServices, rsErr = crud.GetRoleServices(crud.AccessDb, crud.RoleTable, roleId, serviceIds)
		if rsErr != nil {
			roleServices = []RoleServiceType{}
		}
	}

	permittedRes := CheckAccessType{
		UserId:         userId,
		RoleId:         roleId,
		RoleIds:        roleIds,
		IsActive:       isActive,
		IsAdmin:        isAdmin,
		RoleServices:   roleServices,
		TableId:        tableId,
		OwnerPermitted: ownerPermitted,
	}

	if permittedRes.IsActive && (permittedRes.IsAdmin || permittedRes.OwnerPermitted) {
		return mcresponse.GetResMessage("success", mcresponse.ResponseMessageOptions{
			Message: "Action authorised / permitted.",
			Value:   permittedRes,
		})
	}
	recLen := len(permittedRes.RoleServices)
	if permittedRes.IsActive && recLen > 0 && recLen >= len(crud.RecordIds) {
		return mcresponse.GetResMessage("success", mcresponse.ResponseMessageOptions{
			Message: fmt.Sprintf("Access permitted for %v of %v service-items/records", len(crud.RecordIds), recLen),
			Value:   permittedRes,
		})
	}
	return mcresponse.GetResMessage("unAuthorized", mcresponse.ResponseMessageOptions{
		Message: "Action authorised / permitted.",
		Value:   permittedRes,
	})
}

// GetRoleServices method process and returns the permission to user / user-group/roleId for the specified service items
func (crud *Crud) GetRoleServices(accessDb *sqlx.DB, roleTable string, userRoleId string, serviceIds []string) ([]RoleServiceType, error) {
	var roleServices []RoleServiceType
	// where-in-values
	inValues := ""
	idLen := len(serviceIds)
	for idCount, id := range serviceIds {
		inValues += "'" + id + "'"
		if idLen > 1 && idCount < idLen-1 {
			inValues += ", "
		}
	}
	roleScript := fmt.Sprintf("SELECT role_id, service_id, service_category, can_read, can_create, can_delete, can_update, can_crud from %v WHERE service_id IN (%v) AND role_id=$1 AND is_active=$2", roleTable, inValues)
	rows, err := accessDb.Queryx(roleScript, userRoleId, true)
	if err != nil {
		//errMsg := fmt.Sprintf("Db query Error: %v", err.Error())
		return roleServices, errors.New(fmt.Sprintf("%v", err.Error()))
	}
	defer func(rows *sqlx.Rows) {
		err := rows.Close()
		if err != nil {

		}
	}(rows)
	var (
		roleId, serviceId, serviceCategory                string
		canRead, canCreate, canDelete, canUpdate, canCrud bool
	)
	for rows.Next() {
		if err := rows.Scan(&roleId, &serviceId, &serviceCategory, &canRead, &canCreate, &canDelete, &canUpdate, &canCrud); err == nil {
			roleServices = append(roleServices, RoleServiceType{
				ServiceId:       serviceId,
				RoleId:          roleId,
				ServiceCategory: serviceCategory,
				CanRead:         canRead,
				CanCreate:       canCreate,
				CanUpdate:       canUpdate,
				CanDelete:       canDelete,
				CanCrud:         canCrud,
			})
		}
	}

	return roleServices, nil
}

// TaskPermissionById method determines the access permission by owner, role/group (on coll/table or doc/record(s)) or admin
// for various : create/insert, update, delete/remove, read
func (crud *Crud) TaskPermissionById(taskType string) mcresponse.ResponseMessage {
	// permit crud : by owner, role (on table or record(s)) or admin
	// task permission access variables
	var (
		taskPermitted   = false
		ownerPermitted  = false
		recordPermitted = false
		tablePermitted  = false
		isAdmin         = false
		isActive        = false
		userId          = ""
		tableId         = ""
		roleId          = ""
		roleIds         []string
		roleServices    []RoleServiceType
	)
	// check role-based access
	accessRes := crud.CheckTaskAccess()
	// capture roleServices value
	if accessRes.Code != "success" {
		return accessRes
	}
	// get access-record
	accessRec, ok := accessRes.Value.(CheckAccessType)
	if !ok {
		return mcresponse.GetResMessage("unAuthorized", mcresponse.ResponseMessageOptions{
			Message: "Error parsing task access information/value",
			Value:   nil,
		})
	}
	// set access status variables
	ownerPermitted = accessRec.OwnerPermitted
	isAdmin = accessRec.IsAdmin
	isActive = accessRec.IsActive
	roleServices = accessRec.RoleServices
	userId = accessRec.UserId
	roleId = accessRec.RoleId
	roleIds = accessRec.RoleIds
	tableId = accessRec.TableId
	// validate active status
	if !isActive {
		return mcresponse.GetResMessage("unAuthorized", mcresponse.ResponseMessageOptions{
			Message: "Account is not active. Validate active status",
			Value:   nil,
		})
	}
	// validate roleServices permission, for non-admin/non-owner users
	if !isAdmin && !ownerPermitted && len(roleServices) < 1 {
		return mcresponse.GetResMessage("unAuthorized", mcresponse.ResponseMessageOptions{
			Message: "You are not authorized to perform the requested action/task",
			Value:   nil,
		})
	}

	// filter the roleServices by categories ("collection | table" or "record | document")
	recordIds := crud.RecordIds
	collTabFunc := func(item RoleServiceType) bool {
		return item.ServiceCategory == tableId
	}
	recordFunc := func(item RoleServiceType) bool {
		return ArrayStringContains(recordIds, item.ServiceCategory)
	}

	var (
		roleTables, roleRecords []RoleServiceType
	)
	if len(roleServices) > 0 {
		for _, v := range roleServices {
			if collTabFunc(v) {
				roleTables = append(roleTables, v)
			}
		}
		for _, v := range roleServices {
			if recordFunc(v) {
				roleRecords = append(roleRecords, v)
			}
		}
	}
	// helper functions
	canCreateFunc := func(item RoleServiceType) bool {
		return item.CanCreate
	}
	canUpdateFunc := func(item RoleServiceType) bool {
		return item.CanUpdate
	}
	canDeleteFunc := func(item RoleServiceType) bool {
		return item.CanDelete
	}
	canReadFunc := func(item RoleServiceType) bool {
		return item.CanRead
	}

	roleUpdateFunc := func(it1 string, it2 RoleServiceType) bool {
		return it2.ServiceId == it1 && it2.CanUpdate
	}
	roleDeleteFunc := func(it1 string, it2 RoleServiceType) bool {
		return it2.ServiceId == it1 && it2.CanDelete
	}
	roleReadFunc := func(it1 string, it2 RoleServiceType) bool {
		return it2.ServiceId == it1 && it2.CanRead
	}

	roleRecFunc := func(it1 string, roleRecs []RoleServiceType, roleFunc RoleFuncType) bool {
		// test if any or some of the roleRecords it1/it2 met the access condition
		for _, it2 := range roleRecs {
			if roleFunc(it1, it2) {
				return true
			}
		}
		return false
	}
	// taskType specific permission(s)
	if !isAdmin && len(roleServices) > 0 {
		switch taskType {
		case CreateTask, InsertTask:
			// collection/table level access | only tableId was included in serviceIds
			// must be able to perform create on the specified tableId(s)
			if len(roleTables) > 0 {
				tablePermitted = func() bool {
					for _, v := range roleTables {
						if !canCreateFunc(v) {
							return false
						}
					}
					return true
				}()
			}
		case UpdateTask:
			// collection/table level access
			if len(roleTables) > 0 {
				tablePermitted = func() bool {
					for _, v := range roleTables {
						if !canUpdateFunc(v) {
							return false
						}
					}
					return true
				}()
			}
			// document/record level access: all recordIds must have at least a match in the roleRecords
			if len(recordIds) > 0 {
				recordPermitted = func() bool {
					for _, v := range recordIds {
						if !roleRecFunc(v, roleRecords, roleUpdateFunc) {
							return false
						}
					}
					return true
				}()
			}
		case DeleteTask, RemoveTask:
			// collection/table level access
			if len(roleTables) > 0 {
				tablePermitted = func() bool {
					for _, v := range roleTables {
						if !canDeleteFunc(v) {
							return false
						}
					}
					return true
				}()
			}
			// document/record level access: all recordIds must have at least a match in the roleRecords
			if len(recordIds) > 0 {
				recordPermitted = func() bool {
					for _, v := range recordIds {
						if !roleRecFunc(v, roleRecords, roleDeleteFunc) {
							return false
						}
					}
					return true
				}()
			}
		case ReadTask:
			// collection/table level access
			if len(roleTables) > 0 {
				tablePermitted = func() bool {
					for _, v := range roleTables {
						if !canReadFunc(v) {
							return false
						}
					}
					return true
				}()
			}
			// document/record level access: all recordIds must have at least a match in the roleRecords
			if len(recordIds) > 0 {
				recordPermitted = func() bool {
					for _, v := range recordIds {
						if !roleRecFunc(v, roleRecords, roleReadFunc) {
							return false
						}
					}
					return true
				}()
			}
		default:
			return mcresponse.GetResMessage("unAuthorized", mcresponse.ResponseMessageOptions{
				Message: "Unknown access type or access type not specified.",
				Value:   nil,
			})
		}
	}

	// overall access permitted
	taskPermitted = recordPermitted || tablePermitted || ownerPermitted || isAdmin

	if !taskPermitted {
		return mcresponse.GetResMessage("unAuthorized", mcresponse.ResponseMessageOptions{
			Message: "You are not authorized to perform the requested action/task.",
			Value: TaskPermissionType{
				Ok: taskPermitted,
			},
		})
	}
	// if all went well
	return mcresponse.GetResMessage("success", mcresponse.ResponseMessageOptions{
		Message: "Action authorised / permitted.",
		Value: TaskPermissionType{
			Ok:             taskPermitted,
			IsAdmin:        isAdmin,
			IsActive:       isActive,
			UserId:         userId,
			RoleId:         roleId,
			RoleIds:        roleIds,
			OwnerPermitted: ownerPermitted,
		},
	})
}

func (crud *Crud) TaskPermissionByParam(taskType string) mcresponse.ResponseMessage {
	// ids of records, from queryParams
	var recordIds []string
	if len(crud.CurrentRecords) < 1 {
		currentRecRes := crud.GetByParam()
		if currentRecRes.Code != "success" {
			return currentRecRes
		}
		result, ok := currentRecRes.Value.(GetResultType)
		if !ok {
			return mcresponse.GetResMessage("notFound", mcresponse.ResponseMessageOptions{
				Message: "Missing or Invalid record(s) for task-permission-by-queryParams",
				Value:   result,
			})
		}
		crud.CurrentRecords = result.Records
	}
	for _, rec := range crud.CurrentRecords {
		//val, _ := rec.(ActionParamType)
		id, ok := rec["id"].(string)
		if !ok {
			return mcresponse.GetResMessage("notFound", mcresponse.ResponseMessageOptions{
				Message: "Missing record(s) for task-permission-by-queryParams",
				Value:   rec,
			})
		}
		recordIds = append(recordIds, id)
	}
	crud.RecordIds = recordIds
	return crud.TaskPermissionById(taskType)
}

// CheckUserAccess method determines the user access status: active, valid login and admin
func (crud *Crud) CheckUserAccess() mcresponse.ResponseMessage {
	// validate current user active status: by token (API) and user/loggedIn-status
	// get the accessKey information for the user
	accessScript := fmt.Sprintf("SELECT expire from %v WHERE user_id=$1 AND token=$2 AND login_name=$3", crud.AccessTable)
	rowAccess := crud.AccessDb.QueryRow(accessScript, crud.UserInfo.UserId, crud.UserInfo.Token, crud.UserInfo.LoginName)
	// check login-status/expiration
	var accessExpire int64
	if aErr := rowAccess.Scan(&accessExpire); aErr != nil {
		return mcresponse.GetResMessage("unAuthorized", mcresponse.ResponseMessageOptions{
			Message: fmt.Sprintf("UnAuthorized: please ensure that you are logged-in: %v", aErr.Error()),
			Value:   nil,
		})
	} else {
		if (time.Now().Unix() * 1000) > accessExpire {
			return mcresponse.GetResMessage("tokenExpired", mcresponse.ResponseMessageOptions{
				Message: "Access expired: please login to continue",
				Value:   nil,
			})
		}
	}
	// check the current-user status/info
	var (
		userId   string
		isAdmin  bool
		roleIds  []string
		isActive bool
	)
	userScript := fmt.Sprintf("SELECT id, is_admin, is_active from %v WHERE id=$1 AND is_active=$2", crud.UserTable)
	uRow := crud.AccessDb.QueryRow(userScript, crud.UserInfo.UserId, true)
	if uErr := uRow.Scan(&userId, &isAdmin, &isActive); uErr != nil {
		return mcresponse.GetResMessage("unAuthorized", mcresponse.ResponseMessageOptions{
			Message: fmt.Sprintf("UnAuthorized: user information not found or is inactive: %v", uErr.Error()),
			Value:   nil,
		})
	}
	// get user-role/roleIds and profile/roleId information
	urScript := fmt.Sprintf("SELECT id from %v WHERE user_id=$1 AND is_active=$2", crud.UserRoleTable)
	urRows, urErr := crud.AccessDb.Queryx(urScript, crud.UserInfo.UserId, true)

	//if urErr != nil {
	//	roleIds = []string{}
	//	return mcresponse.GetResMessage("notFound", mcresponse.ResponseMessageOptions{
	//		Message: fmt.Sprintf("User-role record not found or could not be processed: %v", urErr.Error()),
	//		Value:   nil,
	//	})
	//}

	if urErr == nil {
		defer func(rows *sqlx.Rows) {
			err := rows.Close()
			if err != nil {

			}
		}(urRows)
		// process user-role-records
		for urRows.Next() {
			roleId := ""
			scanRowErr := urRows.Scan(&roleId)
			if scanRowErr == nil {
				roleIds = append(roleIds, roleId)
			}
		}
	}
	// user-profile
	var roleId string
	upScript := fmt.Sprintf("SELECT id from %v WHERE user_id=$1 AND is_active=$2", crud.ProfileTable)
	upErr := crud.AccessDb.QueryRowx(upScript, crud.UserInfo.UserId, true).Scan(&roleId)
	if upErr != nil {
		roleId = ""
	}
	// if all went well
	return mcresponse.GetResMessage("success", mcresponse.ResponseMessageOptions{
		Message: "Action authorised / permitted.",
		Value: AccessInfoType{
			UserId:   userId,
			RoleId:   roleId,
			RoleIds:  roleIds,
			IsAdmin:  isAdmin,
			IsActive: isActive,
		},
	})
}

// CheckLoginStatus method checks if the user exists and has active login status/token
func (crud *Crud) CheckLoginStatus() mcresponse.ResponseMessage {
	params := crud.UserInfo
	// check if user exists, from users table
	var userId string
	userQuery := fmt.Sprintf("SELECT id from %v WHERE id=$1 AND (email=$2 OR username=$3)", crud.UserTable)
	uRow := crud.AccessDb.QueryRow(userQuery, params.UserId, params.LoginName, params.LoginName)
	uErr := uRow.Scan(&userId)
	if uErr != nil {
		return mcresponse.GetResMessage("unAuthorized", mcresponse.ResponseMessageOptions{
			Message: fmt.Sprintf("Record not found for %v. Register a new account: %v", params.LoginName, uErr.Error()),
			Value:   nil,
		})
	}

	// check loginName, userId and token validity... from access_keys table
	var expire int64
	accessQuery := fmt.Sprintf("SELECT expire from %v WHERE user_id=$1 AND login_name=$2 AND token=$3", crud.AccessTable)
	aRow := crud.AccessDb.QueryRow(accessQuery, params.UserId, params.LoginName, params.Token)
	err := aRow.Scan(&expire)
	if err != nil {
		return mcresponse.GetResMessage("unAuthorized", mcresponse.ResponseMessageOptions{
			Message: fmt.Sprintf("Access information for %v not found. Login first, or contact system administrator: %v", params.LoginName, err.Error()),
			Value:   nil,
		})
	}
	if (time.Now().Unix() * 1000) > expire {
		// Delete the expired access_keys | remove access-info from access_keys table
		delQuery := fmt.Sprintf("DELETE FROM %v WHERE user_id=$1 AND token=$2", crud.AccessTable)
		_, _ = crud.AppDb.Exec(delQuery, params.UserId, params.Token)
		return mcresponse.GetResMessage("tokenExpired", mcresponse.ResponseMessageOptions{
			Message: "Access expired: please login to continue",
			Value:   nil,
		})
	}
	// if all went well
	return mcresponse.GetResMessage("success", mcresponse.ResponseMessageOptions{
		Message: "Action authorised / Access permitted.",
		Value:   userId,
	})
}
