// @Author: abbeymart | Abi Akindele | @Created: 2020-12-24 | @Updated: 2020-12-24, 2026-06-07
// @Company: mConnect.biz | @License: MIT
// @Description: delete records' test cases

package mcdbcrud

import (
	"fmt"
	"testing"

	"github.com/abbeymart/mctest"
)

func TestDelete(t *testing.T) {
	myDb := MyDb
	myDb.Options = DbConnectOptions{}
	// db-connection
	dbc, dbcErr := myDb.OpenDbx()
	// defer dbClose
	defer myDb.CloseDbx()
	// check db-connection-error
	if dbcErr != nil {
		fmt.Printf("*****db-connection-error: %v\n", dbcErr.Error())
		return
	}
	model := Audit{}
	modelPtr := AuditPtr{}
	crudParams := CrudParamsType{
		AppDb:        dbc,
		ModelRef:     model,
		ModelPointer: &modelPtr,
		TableName:    DeleteTable,
		UserInfo:     TestUserInfo,
		RecordIds:    []string{},
		QueryParams:  QueryParamType{},
	}
	crud := NewCrud(crudParams, CrudParamOptions)

	var results []mctest.UnitTestResult

	test1 := mctest.NewTest(mctest.ParamsType{
		Name: "should prevent the delete of all table records and return removeError:",
	})
	test1.SetTestFunction(func() {
		crud.TableName = DeleteAllTable
		res := crud.DeleteRecord()
		fmt.Printf("delete-all: %#v \n", res)
		test1.AssertEquals(res.Code, "removeError", "delete-task permitted by ids or queryParams only: removeError code expected")
	})
	test1Result := test1.RunTest()
	results = append(results, test1Result)

	test2 := mctest.NewTest(mctest.ParamsType{
		Name: "should delete record by Id and return success or notFound[delete-record-method]:",
	})
	test2.SetTestFunction(func() {
		crud.TableName = DeleteTable
		crud.RecordIds = []string{DeleteAuditById}
		crud.QueryParams = QueryParamType{}
		// get-record method params
		res := crud.DeleteRecord()
		fmt.Printf("delete-by-id-result: %#v \n", res)
		resCode := res.Code == "success" || res.Code == "notFound"
		test2.AssertEquals(resCode, true, "delete-by-id should return code: success or notFound")
	})
	test2Result := test2.RunTest()
	results = append(results, test2Result)

	test3 := mctest.NewTest(mctest.ParamsType{
		Name: "should delete records by Ids and return success or notFound[delete-record-method]:",
	})
	test3.SetTestFunction(func() {
		crud.TableName = DeleteTable
		crud.RecordIds = DeleteAuditByIds
		crud.QueryParams = QueryParamType{}
		// get-record method params
		res := crud.DeleteRecord()
		fmt.Printf("delete-by-ids-result: %#v \n", res)
		resCode := res.Code == "success" || res.Code == "notFound"
		test3.AssertEquals(resCode, true, "delete-by-ids should return code: success or notFound")
	})
	test3Result := test3.RunTest()
	results = append(results, test3Result)

	test4 := mctest.NewTest(mctest.ParamsType{
		Name: "should delete records by query-params and return success or notFound[delete-record-method]:",
	})
	test4.SetTestFunction(func() {
		crud.TableName = DeleteTable
		crud.RecordIds = []string{}
		crud.QueryParams = DeleteAuditByParams
		res := crud.DeleteRecord()
		fmt.Printf("delete-by-params-result: %#v \n", res)
		resCode := res.Code == "success" || res.Code == "notFound"
		test4.AssertEquals(resCode, true, "delete-by-params should return code: success or notFound")
	})

	mctest.TestResult(results)

}
