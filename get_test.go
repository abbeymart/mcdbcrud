// @Author: abbeymart | Abi Akindele | @Created: 2020-12-24 | @Updated: 2020-12-24
// @Company: mConnect.biz | @License: MIT
// @Description: get/read records test cases

package mcdbcrud

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/abbeymart/mctest"
	"testing"
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
	audit := Audit{}
	crudParams := CrudParamsType{
		AppDb:        dbc,
		ModelRef:     audit,
		ModelPointer: &audit,
		TableName:    GetTable,
		UserInfo:     TestUserInfo,
		RecordIds:    []string{},
		QueryParams:  QueryParamType{},
	}
	var crud = NewCrud(crudParams, CrudParamOptions)

	mctest.McTest(mctest.OptionValue{
		Name: "should get records by Id and return success:",
		TestFunc: func() {
			crud.RecordIds = []string{GetAuditById}
			res := crud.GetRecord()
			fmt.Printf("get-by-id-response: %#v\n\n", res)
			value, _ := res.Value.(GetResultType)
			var logRecords interface{}
			logRecs := value.Records[0]["logRecords"]
			jsonVal, _ := json.Marshal(logRecs)
			_ = json.Unmarshal(jsonVal, &logRecords)
			strVal, _ := logRecs.(string)
			decoded, _ := base64.StdEncoding.DecodeString(strVal)
			fmt.Printf("json-records: %#v\n\n %#v \n\n", logRecs, logRecords)
			fmt.Printf("decoded-json-records: %#v\n\n", string(decoded))
			fmt.Printf("get-by-id-response, code:recsCount %v:%v :\n", res.Code, value.Stats.RecordsCount)
			mctest.AssertEquals(t, res.Code, "success", "get-task should return code: success")
			mctest.AssertEquals(t, value.Stats.RecordsCount, 1, "get-task-count should be: 1")
			mctest.AssertEquals(t, len(value.Records), 1, "get-result-count should be: 1")
		},
	})

	mctest.McTest(mctest.OptionValue{
		Name: "should get records by Ids and return success:",
		TestFunc: func() {
			crud.TableName = GetTable
			crud.RecordIds = GetAuditByIds
			crud.QueryParams = QueryParamType{}
			recLen := len(crud.RecordIds)
			res := crud.GetByIds()
			fmt.Printf("get-by-ids-response: %#v\n\n", res)
			value, _ := res.Value.(GetResultType)
			fmt.Printf("json-records: %#v\n\n", value.Records)
			fmt.Printf("get-by-ids-response, code:recsCount %v:%v :\n", res.Code, value.Stats.RecordsCount)
			mctest.AssertEquals(t, res.Code, "success", "get-task should return code: success")
			mctest.AssertEquals(t, value.Stats.RecordsCount, recLen, fmt.Sprintf("get-task-count should be: %v", recLen))
			mctest.AssertEquals(t, len(value.Records), recLen, fmt.Sprintf("get-result-count should be: %v", recLen))
		},
	})
	mctest.McTest(mctest.OptionValue{
		Name: "should get records by query-params and return success:",
		TestFunc: func() {
			crud.TableName = GetTable
			crud.RecordIds = []string{}
			crud.QueryParams = GetAuditByParams
			res := crud.GetByParam()
			//fmt.Printf("get-by-param-response: %#v\n", res)
			value, _ := res.Value.(GetResultType)
			//fmt.Printf("json-records: %#v\n\n", value.Records)
			fmt.Printf("get-by-params-response, code:recsCount %v:%v :\n", res.Code, value.Stats.RecordsCount)
			mctest.AssertEquals(t, res.Code, "success", "get-task should return code: success")
			mctest.AssertEquals(t, value.Stats.RecordsCount > 0, true, "get-task-count should be >= 0")
			mctest.AssertEquals(t, len(value.Records) > 0, true, "get-result-count should be >= 0")
		},
	})

	mctest.McTest(mctest.OptionValue{
		Name: "should get all records and return success:",
		TestFunc: func() {
			crud.TableName = GetTable
			crud.RecordIds = []string{}
			crud.QueryParams = QueryParamType{}
			res := crud.GetAll()
			value, _ := res.Value.(GetResultType)
			fmt.Printf("get-all-response, code:recsCount %v:%v :\n", res.Code, value.Stats.RecordsCount)
			mctest.AssertEquals(t, res.Code, "success", "get-task should return code: success")
			mctest.AssertEquals(t, value.Stats.RecordsCount > 20, true, "get-task-count should be >= 10")
			mctest.AssertEquals(t, len(value.Records) > 20, true, "get-result-count should be >= 10")
		},
	})
	mctest.McTest(mctest.OptionValue{
		Name: "should get all records by limit/skip(offset) and return success:",
		TestFunc: func() {
			crud.TableName = GetTable
			crud.RecordIds = []string{}
			crud.QueryParams = QueryParamType{}
			crud.Skip = 0
			crud.Limit = 20
			res := crud.GetAll()
			value, _ := res.Value.(GetResultType)
			fmt.Printf("get-all-limit-response, code:recsCount %v:%v :\n", res.Code, value.Stats.RecordsCount)
			mctest.AssertEquals(t, res.Code, "success", "get-task should return code: success")
			mctest.AssertEquals(t, value.Stats.RecordsCount == 20, true, "get-task-count should be = 20")
			mctest.AssertEquals(t, len(value.Records) == 20, true, "get-result-count should be = 20")
		},
	})
	mctest.PostTestResult()

}
