// @Author: abbeymart | Abi Akindele | @Created: 2020-12-05 | @Updated: 2020-12-05
// @Company: mConnect.biz | @License: MIT
// @Description: go: mConnect

package mcdbcrud

import (
	"fmt"
	"strings"
	"testing"
)
import (
	"github.com/abbeymart/mctest"
)

func TestAuditLogx(t *testing.T) {
	// test-data: db-configuration settings

	tableName := "services"
	userId := "085f48c5-8763-4e22-a1c6-ac1a68ba07de"
	recs := TestParam{Name: "Abi", Desc: "Testing only", Url: "localhost:9000", Priority: 1, Cost: 1000.00}
	//recs, _ := json.Marshal(recs)
	//fmt.Println("table-records-json", LogRecordsType{LogRecords: recs})
	newRecs := TestParam{Name: "Abi Akindele", Desc: "Testing only - updated", Url: "localhost:9900", Priority: 1, Cost: 2000.00}
	//newRecs, _ := json.Marshal(newRecs)
	//fmt.Println("new-table-records-json", LogRecordsType{LogRecords: newRecs})
	readP := map[string][]string{"keywords": {"lagos", "nigeria", "ghana", "accra"}}
	//readP, _ := json.Marshal(readP)

	myDb := MyDb
	myDb.Options = DbConnectOptions{}

	// db-connection
	dbc, err := myDb.OpenDb()
	//fmt.Printf("*****dbc-info: %v\n", dbc)
	// defer dbClose
	defer myDb.CloseDb()
	// check db-connection-error
	if err != nil {
		fmt.Printf("*****db-connection-error: %v\n", err.Error())
		return
	}
	// expected db-connection result
	mcLogResult := LogParam{AuditDb: dbc, AuditTable: "audits"}
	// audit-log instance
	mcLog := NewAuditLog(dbc, "audits")

	mctest.McTest(mctest.OptionValue{
		Name: "[Sqlx]should connect to the DB and return an instance object:",
		TestFunc: func() {
			mctest.AssertEquals(t, err, nil, "error-response should be: nil")
			mctest.AssertEquals(t, mcLog, mcLogResult, "db-connection instance should be: "+mcLogResult.String())
		},
	})
	mctest.McTest(mctest.OptionValue{
		Name: "[Sqlx]should store create-transaction log and return success:",
		TestFunc: func() {
			res, err := mcLog.AuditLog(CreateLog, userId, AuditLogOptionsType{
				TableName:  tableName,
				LogRecords: LogRecordsType{LogRecords: recs},
			})
			//fmt.Printf("create-log: %v", res)
			mctest.AssertEquals(t, err, nil, "error-response should be: nil")
			mctest.AssertEquals(t, res.Code, "success", "log-action response-code should be: success")
		},
	})
	mctest.McTest(mctest.OptionValue{
		Name: "[Sqlx]should store update-transaction log and return success:",
		TestFunc: func() {
			res, err := mcLog.AuditLog(UpdateLog, userId, AuditLogOptionsType{
				TableName:     tableName,
				LogRecords:    LogRecordsType{LogRecords: recs},
				NewLogRecords: LogRecordsType{LogRecords: newRecs},
			})
			mctest.AssertEquals(t, err, nil, "error-response should be: nil")
			mctest.AssertEquals(t, res.Code, "success", "log-action response-code should be: success")
		},
	})
	mctest.McTest(mctest.OptionValue{
		Name: "[Sqlx]should store read-transaction log and return success:",
		TestFunc: func() {
			res, err := mcLog.AuditLog(ReadLog, userId, AuditLogOptionsType{
				TableName:  tableName,
				LogRecords: LogRecordsType{LogRecords: readP},
			})
			mctest.AssertEquals(t, err, nil, "error-response should be: nil")
			mctest.AssertEquals(t, res.Code, "success", "log-action response-code should be: success")
		},
	})
	mctest.McTest(mctest.OptionValue{
		Name: "[Sqlx]should store delete-transaction log and return success:",
		TestFunc: func() {
			res, err := mcLog.AuditLog(DeleteLog, userId, AuditLogOptionsType{
				TableName:  tableName,
				LogRecords: LogRecordsType{LogRecords: recs},
			})
			mctest.AssertEquals(t, err, nil, "error-response should be: nil")
			mctest.AssertEquals(t, res.Code, "success", "log-action response-code should be: success")
		},
	})
	mctest.McTest(mctest.OptionValue{
		Name: "[Sqlx]should store login-transaction log and return success:",
		TestFunc: func() {
			res, err := mcLog.AuditLog(LoginLog, userId, AuditLogOptionsType{
				TableName:  tableName,
				LogRecords: LogRecordsType{LogRecords: recs},
			})
			mctest.AssertEquals(t, err, nil, "error-response should be: nil")
			mctest.AssertEquals(t, res.Code, "success", "log-action response-code should be: success")
		},
	})
	mctest.McTest(mctest.OptionValue{
		Name: "[Sqlx]should store logout-transaction log and return success:",
		TestFunc: func() {
			res, err := mcLog.AuditLog(LogoutLog, userId, AuditLogOptionsType{
				TableName:  tableName,
				LogRecords: LogRecordsType{LogRecords: recs},
			})
			mctest.AssertEquals(t, err, nil, "error-response should be: nil")
			mctest.AssertEquals(t, res.Code, "success", "log-action response-code should be: success")
		},
	})
	mctest.McTest(mctest.OptionValue{
		Name: "[Sqlx]should return paramsError for incomplete/undefined inputs:",
		TestFunc: func() {
			res, err := mcLog.AuditLog(CreateLog, "", AuditLogOptionsType{
				TableName:  tableName,
				LogRecords: LogRecordsType{LogRecords: recs},
			})
			//fmt.Printf("params-res: %#v", res)
			mctest.AssertNotEquals(t, err, nil, "error-response should not be: nil")
			mctest.AssertEquals(t, res.Code, "paramsError", "log-action response-code should be: paramsError")
			mctest.AssertEquals(t, strings.Contains(res.Message, "userId is required"), true, "log-action response-message should be: true")
			mctest.AssertEquals(t, strings.Contains(err.Error(), "userId is required"), true, "log-action error-message should be: true")
		},
	})

	mctest.PostTestResult()
}
