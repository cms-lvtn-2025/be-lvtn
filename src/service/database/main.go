package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var Db *sql.DB

func GetInstanceDB(username, password, host, port, databaseName string) *sql.DB {
	if Db != nil {
		return Db
	}
	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		username, password, host, port, databaseName)

	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		panic(err)
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	Db = db
	return Db
}

func Close() {
	if Db != nil {
		Db.Close()
	}
}
