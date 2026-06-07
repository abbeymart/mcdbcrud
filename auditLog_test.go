// @Author: abbeymart | Abi Akindele | @Created: 2020-12-05 | @Updated: 2020-12-05, 2026-06-06
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

type TestParam struct {
	Name     string
	Desc     string
	Url      string
	Priority int
	Cost     float64
}

func TestAuditLog(t *testing.T) {
	// test-data: db-configuration settings

	tableName := "services"
	userId := "085f48c5-8763-4e22-a1c6-ac1a68ba07de"
	recs := TestParam{Name: "Abi", Desc: "Testing only", Url: "localhost:9000", Priority: 1, Cost: 1000.00}
	newRecs := TestParam{Name: "Abi Akindele", Desc: "Testing only - updated", Url: "localhost:9900", Priority: 1, Cost: 2000.00}
	readP := map[string][]string{"keywords": {"lagos", "nigeria", "ghana", "accra"}}

	myDb := MyDb
	myDb.Options = DbConnectOptions{}

	// db-connection
	dbc, dbcErr := myDb.OpenDb()
	// defer dbClose
	defer myDb.CloseDb()
	// check db-connection-error
	if dbcErr != nil {
		fmt.Printf("*****db-connection-error: %v\n", dbcErr.Error())
		return
	}
	// expected db-connection result
	mcLogResult := LogParam{AuditDb: dbc, AuditTable: "audits"}

	var results []mctest.UnitTestResult

	// audit-log instance
	mcLog := NewAuditLog(dbc, "audits")

	test1 := mctest.NewTest(mctest.ParamsType{
		Name: "[Sql]should connect to the DB and return an instance object:",
	})
	test1.SetTestFunction(func() {
		test1.AssertEquals(dbcErr, nil, "error-response should be: nil")
		test1.AssertEquals(mcLog, mcLogResult, "db-connection instance should be: "+mcLogResult.String())
	})
	test1Result := test1.RunTest()
	results = append(results, test1Result)

	test2 := mctest.NewTest(mctest.ParamsType{
		Name: "[Sql]should store create-transaction log and return success:",
	})
	test2.SetTestFunction(func() {
		res, err := mcLog.AuditLog(CreateLog, userId, AuditLogOptionsType{
			TableName:  tableName,
			LogRecords: LogRecordsType{LogRecords: recs},
		})
		test2.AssertEquals(err, nil, "error-response should be: nil")
		test2.AssertEquals(res.Code, "success", "log-action response-code should be: success")
	})
	test2Result := test2.RunTest()
	results = append(results, test2Result)

	test3 := mctest.NewTest(mctest.ParamsType{
		Name: "[Sql]should store update-transaction log and return success:",
	})
	test3.SetTestFunction(func() {
		res, err := mcLog.AuditLog(UpdateLog, userId, AuditLogOptionsType{
			TableName:     tableName,
			LogRecords:    LogRecordsType{LogRecords: recs},
			NewLogRecords: LogRecordsType{LogRecords: newRecs},
		})
		test3.AssertEquals(err, nil, "error-response should be: nil")
		test3.AssertEquals(res.Code, "success", "log-action response-code should be: success")
	})
	test3Result := test3.RunTest()
	results = append(results, test3Result)

	test4 := mctest.NewTest(mctest.ParamsType{
		Name: "[Sql]should store read-transaction log and return success:",
	})
	test4.SetTestFunction(func() {
		res, err := mcLog.AuditLog(ReadLog, userId, AuditLogOptionsType{
			TableName:  tableName,
			LogRecords: LogRecordsType{LogRecords: readP},
		})
		test4.AssertEquals(err, nil, "error-response should be: nil")
		test4.AssertEquals(res.Code, "success", "log-action response-code should be: success")
	})
	test4Result := test4.RunTest()
	results = append(results, test4Result)

	test5 := mctest.NewTest(mctest.ParamsType{
		Name: "[Sql]should store delete-transaction log and return success:",
	})
	test5.SetTestFunction(func() {
		res, err := mcLog.AuditLog(DeleteLog, userId, AuditLogOptionsType{
			TableName:  tableName,
			LogRecords: LogRecordsType{LogRecords: recs},
		})
		test5.AssertEquals(err, nil, "error-response should be: nil")
		test5.AssertEquals(res.Code, "success", "log-action response-code should be: success")
	})
	test5Result := test5.RunTest()
	results = append(results, test5Result)

	test6 := mctest.NewTest(mctest.ParamsType{
		Name: "[Sql]should store login-transaction log and return success:",
	})
	test6.SetTestFunction(func() {
		res, err := mcLog.AuditLog(LoginLog, userId, AuditLogOptionsType{
			TableName:  tableName,
			LogRecords: LogRecordsType{LogRecords: recs},
		})
		test6.AssertEquals(err, nil, "error-response should be: nil")
		test6.AssertEquals(res.Code, "success", "log-action response-code should be: success")
	})
	test6Result := test6.RunTest()
	results = append(results, test6Result)

	test7 := mctest.NewTest(mctest.ParamsType{
		Name: "[Sql]should store logout-transaction log and return success:",
	})
	test7.SetTestFunction(func() {
		res, err := mcLog.AuditLog(LogoutLog, userId, AuditLogOptionsType{
			TableName:  tableName,
			LogRecords: LogRecordsType{LogRecords: recs},
		})
		test7.AssertEquals(err, nil, "error-response should be: nil")
		test7.AssertEquals(res.Code, "success", "log-action response-code should be: success")
	})
	test7Result := test7.RunTest()
	results = append(results, test7Result)

	test8 := mctest.NewTest(mctest.ParamsType{
		Name: "[Sql]should return paramsError for incomplete/undefined inputs:",
	})
	test8.SetTestFunction(func() {
		res, err := mcLog.AuditLog(CreateLog, "", AuditLogOptionsType{
			TableName:  tableName,
			LogRecords: LogRecordsType{LogRecords: recs},
		})
		test8.AssertNotEquals(err, nil, "error-response should not be: nil")
		test8.AssertEquals(res.Code, "paramsError", "log-action response-code should be: paramsError")
		test8.AssertEquals(strings.Contains(res.Message, "userId is required"), true, "log-action response-message should be: true")
		test8.AssertEquals(strings.Contains(err.Error(), "userId is required"), true, "log-action error-message should be: true")
	})
	test8Result := test8.RunTest()
	results = append(results, test8Result)

	mctest.TestResult(results)
}
