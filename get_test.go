// @Author: abbeymart | Abi Akindele | @Created: 2020-12-24 | @Updated: 2020-12-24, 2026-06-07
// @Company: mConnect.biz | @License: MIT
// @Description: get/read records test cases

package mcdbcrud

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/abbeymart/mctest"
)

func TestGet(t *testing.T) {
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
		TableName:    GetTable,
		UserInfo:     TestUserInfo,
		RecordIds:    []string{},
		QueryParams:  QueryParamType{},
	}
	crud := NewCrud(crudParams, CrudParamOptions)

	var results []mctest.UnitTestResult

	test1 := mctest.NewTest(mctest.ParamsType{
		Name: "should get records by Id and return success:",
	})
	test1.SetTestFunction(func() {
		crud.RecordIds = []string{GetAuditById}
		res := crud.GetRecord()
		fmt.Printf("get-by-id-response: %#v\n\n", res)
		value, _ := res.Value.(GetResultType)
		logRecs := value.Records[0]["logRecords"]
		strVal, _ := logRecs.(string)
		decoded, _ := base64.StdEncoding.DecodeString(strVal)
		fmt.Printf("json-records: %#v\n\n", logRecs)
		fmt.Printf("decoded-json-records: %#v\n\n", string(decoded))
		test1.AssertEquals(res.Code, "success", "get-task should return code: success")
		test1.AssertEquals(value.Stats.RecordsCount, 1, "get-task-count should be: 1")
		test1.AssertEquals(len(value.Records), 1, "get-result-count should be: 1")
	})
	test1Result := test1.RunTest()
	results = append(results, test1Result)

	test2 := mctest.NewTest(mctest.ParamsType{
		Name: "should get records by Ids and return success:",
	})
	test2.SetTestFunction(func() {
		crud.TableName = GetTable
		crud.RecordIds = GetAuditByIds
		crud.QueryParams = QueryParamType{}
		recLen := len(crud.RecordIds)
		res := crud.GetByIds()
		fmt.Printf("get-by-ids-response: %#v\n\n", res)
		value, _ := res.Value.(GetResultType)
		fmt.Printf("json-records: %#v\n\n", value.Records)
		fmt.Printf("get-by-ids-response, code:recsCount %v:%v :\n", res.Code, value.Stats.RecordsCount)
		test2.AssertEquals(res.Code, "success", "get-task should return code: success")
		test2.AssertEquals(value.Stats.RecordsCount, recLen, fmt.Sprintf("get-task-count should be: %v", recLen))
		test2.AssertEquals(len(value.Records), recLen, fmt.Sprintf("get-result-count should be: %v", recLen))
	})
	test2Result := test2.RunTest()
	results = append(results, test2Result)

	test3 := mctest.NewTest(mctest.ParamsType{
		Name: "should get records by query-params and return success:",
	})
	test3.SetTestFunction(func() {
		crud.TableName = GetTable
		crud.RecordIds = []string{}
		crud.QueryParams = GetAuditByParams
		res := crud.GetByParam()
		value, _ := res.Value.(GetResultType)
		fmt.Printf("get-by-params-response, code:recsCount %v:%v :\n", res.Code, value.Stats.RecordsCount)
		test3.AssertEquals(res.Code, "success", "get-task should return code: success")
		test3.AssertEquals(value.Stats.RecordsCount > 0, true, "get-task-count should be >= 0")
		test3.AssertEquals(len(value.Records) > 0, true, "get-result-count should be >= 0")
	})
	test3Result := test3.RunTest()
	results = append(results, test3Result)

	test4 := mctest.NewTest(mctest.ParamsType{
		Name: "should get all records and return success:",
	})
	test4.SetTestFunction(func() {
		crud.TableName = GetTable
		crud.RecordIds = []string{}
		crud.QueryParams = QueryParamType{}
		res := crud.GetAll()
		value, _ := res.Value.(GetResultType)
		fmt.Printf("get-all-response, code:recsCount %v:%v :\n", res.Code, value.Stats.RecordsCount)
		test4.AssertEquals(res.Code, "success", "get-task should return code: success")
		test4.AssertEquals(value.Stats.RecordsCount > 20, true, "get-task-count should be >= 10")
		test4.AssertEquals(len(value.Records) > 20, true, "get-result-count should be >= 10")
	})
	test4Result := test4.RunTest()
	results = append(results, test4Result)

	test5 := mctest.NewTest(mctest.ParamsType{
		Name: "should get all records by limit/skip(offset) and return success:",
	})
	test5.SetTestFunction(func() {
		crud.TableName = GetTable
		crud.RecordIds = []string{}
		crud.QueryParams = QueryParamType{}
		crud.Skip = 0
		crud.Limit = 20
		res := crud.GetAll()
		value, _ := res.Value.(GetResultType)
		fmt.Printf("get-all-limit-response, code:recsCount %v:%v :\n", res.Code, value.Stats.RecordsCount)
		test5.AssertEquals(res.Code, "success", "get-task should return code: success")
		test5.AssertEquals(value.Stats.RecordsCount == 20, true, "get-task-count should be = 20")
		test5.AssertEquals(len(value.Records) == 20, true, "get-result-count should be = 20")
	})
	test5Result := test5.RunTest()
	results = append(results, test5Result)

	test6 := mctest.NewTest(mctest.ParamsType{
		Name: "custom-query: should get records by Id and return success:",
	})
	test6.SetTestFunction(func() {
		selectFields, fieldErr := QueryFields(model)
		if fieldErr != nil {
			errMsg := fmt.Sprintf("SELECT query fields computation error: %v", fieldErr.Error())
			test6.AssertEquals(fieldErr, nil, errMsg)
		}
		// compute queries
		countQuery := fmt.Sprintf("SELECT COUNT(*) AS total_rows FROM %v", crud.TableName)
		// perform crud-task action
		selectQuery := fmt.Sprintf("SELECT %v FROM %v WHERE id=$1", selectFields, crud.TableName)
		fieldValues := []interface{}{GetAuditById}
		res := crud.CustomSelectQuery(CustomSelectQueryParamsType{
			SelectQuery:                selectQuery,
			CountQuery:                 countQuery,
			TableName:                  crud.TableName,
			QueryPositionalFieldValues: fieldValues,
			ModelPointer:               &modelPtr,
		})
		fmt.Printf("get-by-id-response: %#v\n\n", res)
		value, _ := res.Value.(GetResultType)
		logRecs := value.Records[0]["logRecords"]
		strVal, _ := logRecs.(string)
		decoded, _ := base64.StdEncoding.DecodeString(strVal)
		fmt.Printf("json-records: %#v\n\n", logRecs)
		fmt.Printf("decoded-json-records: %#v\n\n", string(decoded))
		test6.AssertEquals(res.Code, "success", "get-task should return code: success")
		test6.AssertEquals(value.Stats.RecordsCount, 1, "get-task-count should be: 1")
		test6.AssertEquals(len(value.Records), 1, "get-result-count should be: 1")
	})
	test6Result := test6.RunTest()
	results = append(results, test6Result)

	mctest.TestResult(results)

}
