package dbwrapper

import (
	"fmt"
	"os"
	"database/sql"

	"github.com/joho/godotenv"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func init() {
	err := godotenv.Load()
	if err != nil {
		panic(err.Error())
	}
	var conn *sql.DB
	if os.Getenv("mode") == "dev" {
		//"user:password@/database"
		conn, err = sql.Open("mysql", fmt.Sprintf("%s:%s@/%s", os.Getenv("db_user"), os.Getenv("db_pass"), os.Getenv("db_name")) + "?parseTime=true")
	} else if os.Getenv("mode") == "prod" {
		//"user:password@tcp(endpoint:port)/name"
		conn, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
			os.Getenv("db_user"), os.Getenv("db_pass"), os.Getenv("db_endpoint"), os.Getenv("db_port"), os.Getenv("db_name")) + "?parseTime=true")
	}

	if err != nil {
		panic(err.Error())
	}

	db = conn
}

func GetDB() (*sql.DB) {
	return db
}
