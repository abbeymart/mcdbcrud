// @Author: abbeymart | Abi Akindele | @Created: 2020-12-04 | @Updated: 2020-12-04, 2026-06-06
// @Company: mConnect.biz | @License: MIT
// @Description: db testing

package mcdbcrud

import (
	"fmt"
	"testing"
)
import "github.com/abbeymart/mctest"

func TestDb(t *testing.T) {
	// test-data: db-configuration settings
	myDb := MyDb
	myDb.Options = DbConnectOptions{}

	sqliteDb := DbConfig{
		DbType:   "sqlite3",
		Filename: "testdb.db",
	}

	var results []mctest.UnitTestResult

	test1 := mctest.NewTest(mctest.ParamsType{
		Name: "should successfully connect to the PostgresDB [db]",
	})
	test1.SetTestFunction(func() {
		_, err := myDb.OpenDb()
		defer myDb.CloseDb()
		fmt.Println("*****************************************")
		test1.AssertEquals(err, nil, "response-code should be: nil")
	})
	test1Result := test1.RunTest()
	results = append(results, test1Result)

	test2 := mctest.NewTest(mctest.ParamsType{
		Name: "should successfully connect to SQLite3 database [db]",
	})
	test2.SetTestFunction(func() {
		_, err := sqliteDb.OpenDb()
		defer sqliteDb.CloseDb()
		test2.AssertEquals(err, nil, "response-code should be: nil")
	})
	test2Result := test2.RunTest()
	results = append(results, test2Result)

	mctest.TestResult(results)
}
