// @Author: abbeymart | Abi Akindele | @Created: 2020-12-01 | @Updated: 2020-12-01
// @Company: mConnect.biz | @License: MIT
// @Description: mcdb - db connection for PostgresSQL, MySQL, SQLite3...

package mcdbcrud

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"os"
)

var (
	db  *sql.DB
	err error
)

func (dbConfig DbConfig) OpenDb() (*sql.DB, error) {
	sslMode := dbConfig.SecureOptions.SslMode
	sslCert := dbConfig.SecureOptions.SecureCert
	if sslMode == "" {
		sslMode = "disable"
	}
	switch dbConfig.DbType {
	case "postgres":
		connectionString := ""
		if sslCert != "" {
			connectionString = fmt.Sprintf("postgresql://%v:%v@%v:%v/%v?sslmode=%v;sslrootcert=%v", dbConfig.Username, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.DbName, sslMode, sslCert)
			//connectionString = fmt.Sprintf("port=%d host=%s user=%s password=%s dbname=%s sslmode=%v sslrootcert=%v", dbConfig.Port, dbConfig.Host, dbConfig.Username, dbConfig.Password, dbConfig.DbName, sslMode, sslCert)
		} else {
			connectionString = fmt.Sprintf("postgresql://%v:%v@%v:%v/%v?sslmode=%v", dbConfig.Username, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.DbName, sslMode)
		}
		if os.Getenv("DATABASE_URL") != "" && dbConfig.PermitDBUrl {
			connectionString = os.Getenv("DATABASE_URL")
		}
		db, err = sql.Open(dbConfig.DbType, connectionString)
		if err != nil {
			errMsg := fmt.Sprintf("Database Connection Error: %v", err.Error())
			return nil, errors.New(errMsg)
		}
		return db, nil
	case "mysql", "mariadb":
		connectionString := ""
		if sslCert != "" {
			connectionString = fmt.Sprintf("mysql://%v:%v@%v:%v/%v?sslmode=%v;sslrootcert=%v", dbConfig.Username, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.DbName, sslMode, sslCert)
			//connectionString = fmt.Sprintf("port=%d host=%s user=%s password=%s dbname=%s sslmode=%v sslrootcert=%v", dbConfig.Port, dbConfig.Host, dbConfig.Username, dbConfig.Password, dbConfig.DbName, sslMode, sslCert)
		} else {
			connectionString = fmt.Sprintf("mysql://%v:%v@%v:%v/%v?sslmode=%v", dbConfig.Username, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.DbName, sslMode)
		}
		if os.Getenv("DATABASE_URL") != "" && dbConfig.PermitDBUrl {
			connectionString = os.Getenv("DATABASE_URL")
		}
		db, err = sql.Open(dbConfig.DbType, connectionString)
		if err != nil {
			errMsg := fmt.Sprintf("Database Connection Error: %v", err.Error())
			return nil, errors.New(errMsg)
		}
		return db, nil
	case "sqlite3":
		db, err = sql.Open(dbConfig.DbType, dbConfig.Filename)
		if err != nil {
			errMsg := fmt.Sprintf("Database Connection Error: %v", err.Error())
			return nil, errors.New(errMsg)
		}
		return db, nil
	default:
		return nil, errors.New(fmt.Sprintf("unknown db-type(%v)", dbConfig.DbType))
	}
}

func (dbConfig DbConfig) CloseDb() {
	if db != nil {
		err = db.Close()
		if err != nil {
			// log error to the console
			fmt.Println(err)
		}
	}
}
