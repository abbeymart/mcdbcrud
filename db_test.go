// @Author: abbeymart | Abi Akindele | @Created: 2020-12-04 | @Updated: 2020-12-04
// @Company: mConnect.biz | @License: MIT
// @Description: db testing

package mcdbcrud

import (
	"fmt"
	"testing"
)
import "github.com/abbeymart/mctest"

// Setup / update test-DB configuration for testing

var MyDb = DbConfig{
	DbType:   "postgres",
	Host:     "localhost",
	Username: "postgres",
	Password: "ab12testing",
	DbName:   "mctest",
	Port:     4550,
	Filename: "testdb.db",
	PoolSize: 20,
	Url:      "localhost:5433",
	Location: "postgres://localhost:5433/mctest",
}

func TestDb(t *testing.T) {
	// test-data: db-configuration settings
	myDb := MyDb
	myDb.Options = DbConnectOptions{}

	sqliteDb := DbConfig{
		DbType:   "sqlite3",
		Filename: "testdb.db",
	}

	mctest.McTest(mctest.OptionValue{
		Name: "should successfully connect to the PostgresDB",
		TestFunc: func() {
			dbc, err := myDb.OpenDb()
			fmt.Printf("pg-dbc: %v\n", dbc)
			defer myDb.CloseDb()
			fmt.Println(dbc)
			fmt.Println("*****************************************")
			mctest.AssertEquals(t, err, nil, "response-code should be: nil")
		},
	})

	mctest.McTest(mctest.OptionValue{
		Name: "should successfully connect to SQLite3 database",
		TestFunc: func() {
			dbc2, err := sqliteDb.OpenDb()
			defer sqliteDb.CloseDb()
			fmt.Println(dbc2)
			mctest.AssertEquals(t, err, nil, "response-code should be: nil")
		},
	})

	mctest.PostTestResult()
}
