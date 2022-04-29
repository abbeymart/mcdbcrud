// @Author: abbeymart | Abi Akindele | @Created: 2020-12-24 | @Updated: 2020-12-24
// @Company: mConnect.biz | @License: MIT
// @Description: delete records test cases

package mcdbcrud

import (
	"fmt"
	"github.com/abbeymart/mctest"
	"testing"
)

func TestDelete(t *testing.T) {
	myDb := MyDb
	myDb.Options = DbConnectOptions{}
	// db-connection
	dbc, err := myDb.OpenDbx()
	// defer dbClose
	defer myDb.CloseDbx()
	// check db-connection-error
	if err != nil {
		fmt.Printf("*****db-connection-error: %v\n", err.Error())
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

	mctest.McTest(mctest.OptionValue{
		Name: "should prevent the delete of all table records and return removeError:",
		TestFunc: func() {
			crud.TableName = DeleteAllTable
			res := crud.DeleteRecord()
			fmt.Printf("delete-all: %#v \n", res)
			mctest.AssertEquals(t, res.Code, "removeError", "delete-task permitted by ids or queryParams only: removeError code expected")
		},
	})

	mctest.McTest(mctest.OptionValue{
		Name: "should delete record by Id and return success or notFound[delete-record-method]:",
		TestFunc: func() {
			crud.TableName = DeleteTable
			crud.RecordIds = []string{DeleteAuditById}
			crud.QueryParams = QueryParamType{}
			// get-record method params
			res := crud.DeleteRecord()
			fmt.Printf("delete-by-id-result: %#v \n", res)
			resCode := res.Code == "success" || res.Code == "notFound"
			mctest.AssertEquals(t, resCode, true, "delete-by-id should return code: success or notFound")
		},
	})

	mctest.McTest(mctest.OptionValue{
		Name: "should delete records by Ids and return success or notFound[delete-record-method]:",
		TestFunc: func() {
			crud.TableName = DeleteTable
			crud.RecordIds = DeleteAuditByIds
			crud.QueryParams = QueryParamType{}
			// get-record method params
			res := crud.DeleteRecord()
			fmt.Printf("delete-by-ids-result: %#v \n", res)
			resCode := res.Code == "success" || res.Code == "notFound"
			mctest.AssertEquals(t, resCode, true, "delete-by-ids should return code: success or notFound")
		},
	})

	mctest.McTest(mctest.OptionValue{
		Name: "should delete records by query-params and return success or notFound[delete-record-method]:",
		TestFunc: func() {
			crud.TableName = DeleteTable
			crud.RecordIds = []string{}
			crud.QueryParams = DeleteAuditByParams
			res := crud.DeleteRecord()
			fmt.Printf("delete-by-params-result: %#v \n", res)
			resCode := res.Code == "success" || res.Code == "notFound"
			mctest.AssertEquals(t, resCode, true, "delete-by-params should return code: success or notFound")
		},
	})

	mctest.PostTestResult()

}
