// @Author: abbeymart | Abi Akindele | @Created: 2020-12-28 | @Updated: 2020-12-28
// @Company: mConnect.biz | @License: MIT
// @Description: test-cases data: for get, delete and save record(s)

package mcdbcrud

import (
	"encoding/json"
	"time"
)

// Models

type Audit struct {
	Id            string      `json:"id" db:"id"`
	TableName     string      `json:"tableName" db:"table_name"`
	LogRecords    interface{} `json:"logRecords" db:"log_records"`
	NewLogRecords interface{} `json:"newLogRecords" db:"new_log_records"`
	LogType       string      `json:"logType" db:"log_type"`
	LogBy         string      `json:"logBy" db:"log_by"`
	LogAt         time.Time   `json:"logAt" db:"log_at"`
}

type AuditPtr struct {
	Id            string      `json:"id" db:"id"`
	TableName     string      `json:"tableName" db:"table_name"`
	LogRecords    interface{} `json:"logRecords" db:"log_records"`
	NewLogRecords interface{} `json:"newLogRecords" db:"new_log_records"`
	LogType       string      `json:"logType" db:"log_type"`
	LogBy         *string     `json:"logBy" db:"log_by"`
	LogAt         time.Time   `json:"logAt" db:"log_at"`
}

const AuditTable = "audits"
const GetTable = "audits_get"
const DeleteTable = "audits_delete"
const DeleteAllTable = "audits_delete_all"
const UpdateTable = "audits_update"

const UserId = "c85509ac-7373-464d-b667-425bb59b5738" // TODO: review/update

var TestUserInfo = UserInfoType{
	UserId:    "c85509ac-7373-464d-b667-425bb59b5738",
	LoginName: "abbeymart",
	Email:     "abbeya1@yahoo.com",
	Language:  "en-US",
	Firstname: "Abi",
	Lastname:  "Akindele",
	Token:     "",
	Expire:    0,
	RoleId:    "",
}

var CrudParamOptions = CrudOptionsType{
	CheckAccess:   false,
	AuditTable:    "audits",
	UserTable:     "users",
	ProfileTable:  "profiles",
	ServiceTable:  "services",
	AccessTable:   "accesses",
	VerifyTable:   "verify_users",
	RoleTable:     "roles",
	LogCrud:       false,
	LogCreate:     false,
	LogUpdate:     false,
	LogDelete:     false,
	LogRead:       false,
	LogLogin:      false,
	LogLogout:     false,
	MaxQueryLimit: 100000,
	MsgFrom:       "support@mconnect.biz",
}

// TODO: create/update, get & delete records for groups & categories tables

var LogRecords = ActionParamType{
	"name":     "Abi",
	"desc":     "Testing only",
	"url":      "localhost:9000",
	"priority": 100,
	"cost":     1000.00,
}

var NewLogRecords = ActionParamType{
	"name":     "Abi Akindele",
	"desc":     "Testing only - updated",
	"url":      "localhost:9900",
	"priority": 1,
	"cost":     2000.00,
}

var LogRecords2 = ActionParamType{
	"name":     "Ola",
	"desc":     "Testing only - 2",
	"url":      "localhost:9000",
	"priority": 1,
	"cost":     10000.00,
}

var NewLogRecords2 = ActionParamType{
	"name":     "Ola",
	"desc":     "Testing only - 2 - updated",
	"url":      "localhost:9000",
	"priority": 1,
	"cost":     20000.00,
}

// create record(s)

var LogRecs, _ = json.Marshal(LogRecords)
var NewLogRecs, _ = json.Marshal(NewLogRecords)
var LogRecs2, _ = json.Marshal(LogRecords2)
var NewLogRecs2, _ = json.Marshal(NewLogRecords2)

var AuditCreateRec1 = ActionParamType{
	"tableName":  "audits",
	"logAt":      time.Now(),
	"logBy":      UserId,
	"logRecords": string(LogRecs),
	"logType":    CreateTask,
}
var AuditCreateRec2 = ActionParamType{
	"tableName":  "audits",
	"logAt":      time.Now(),
	"logBy":      UserId,
	"logRecords": string(LogRecs2),
	"logType":    CreateTask,
}
var AuditUpdateRec1 = ActionParamType{
	"id":            "c1c3f614-b10d-40a4-9269-4e03f5fcf55e",
	"tableName":     "todos",
	"logAt":         time.Now(),
	"logBy":         UserId,
	"logRecords":    string(LogRecs),
	"newLogRecords": string(NewLogRecs),
	"logType":       UpdateTask,
}

var AuditUpdateRec2 = ActionParamType{
	"id":            "003c1422-c7cb-476f-b96f-9c8028e04a14",
	"tableName":     "todos",
	"logAt":         time.Now(),
	"logBy":         UserId,
	"logRecords":    string(LogRecs2),
	"newLogRecords": string(NewLogRecs2),
	"logType":       UpdateTask,
}

var AuditCreateActionParams = ActionParamsType{
	AuditCreateRec1,
	AuditCreateRec2,
}
var AuditUpdateActionParams = ActionParamsType{
	AuditUpdateRec1,
	AuditUpdateRec2,
}

// TODO: update and delete params, by ids / queryParams

var AuditUpdateRecordById = ActionParamType{
	"id":            "b126f4c0-9bad-4242-bec1-4c4ab74ae481",
	"tableName":     "groups",
	"logAt":         time.Now(),
	"logBy":         UserId,
	"logRecords":    string(LogRecs),
	"newLogRecords": string(NewLogRecs),
	"logType":       DeleteTask,
}
var AuditUpdateRecordByParam = ActionParamType{
	"id":            "f380f132-422f-4cd4-82c1-07b4caf35da0",
	"tableName":     "contacts",
	"logAt":         time.Now(),
	"logBy":         UserId,
	"logRecords":    string(LogRecs),
	"newLogRecords": string(NewLogRecs),
	"logType":       UpdateTask,
}

// GetIds: for get-records by ids & params | TODO: update ids after create

var GetAuditById = "7461ae6c-96e0-4b4f-974b-9a0a7f91e016"
var GetAuditByIds = []string{"7461ae6c-96e0-4b4f-974b-9a0a7f91e016", "aa9ba999-b138-414b-be66-9f0264e50f4a"}
var GetAuditByParams = QueryParamType{
	"logType": "create",
}
var DeleteAuditById = "99f0f869-3c84-4a5e-83ac-3b9f893dcd60"
var DeleteAuditByIds = []string{
	"9e9f7733-7653-4069-9f42-dc157768a960",
	"35304003-567f-4e25-9f1d-6483760db621",
	"d0a1445e-f12f-4d45-98e5-22689dec48e5",
	"39774322-9be5-4b43-9d6e-e2ba514e0f43",
}
var DeleteAuditByParams = QueryParamType{
	"logType": "read",
}
var UpdateAuditById = "98bb024e-2b22-42b4-b379-7099166ad1c9"
var UpdateAuditByIds = []string{
	"c158c19f-e396-4625-96ee-d054ef4f40a1",
	"e34b10f9-6320-4573-96cc-2cd8c69c9a89",
	"9b9acf43-9008-4261-9528-39f47f261adf",
}
var UpdateAuditByParams = QueryParamType{
	"logType": "read",
}
