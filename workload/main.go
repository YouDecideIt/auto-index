package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"sync"

	_ "github.com/go-sql-driver/mysql"
)

var mysqlAddr = flag.String("mysql", "localhost:4000", "mysql address")

func main() {
	flag.Parse()
	dsn := fmt.Sprintf("root:@tcp(%v)/test", *mysqlAddr)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalln(err)
	}

	defer db.Close()

	db.Exec("create database if not exists test")
	db.Exec("use test")
	db.Exec("drop table if exists t")
	_, err = db.Exec("create table t(a int, b int, key(a) invisible, key(b), key(a,b) invisible)")
	if err == nil {
		db.Exec("insert into t values(1, 1), (2, 2), (3, 3)")
		for i := 0; i < 10; i++ {
			db.Exec("insert into t select * from t")
		}
	}
	db.SetMaxIdleConns(20)

	for i := 0; i < 10; i++ {
		go func() {
			for {
				rows, err := db.Query("select * from t where a=1 and b=2")
				if err != nil {
					panic(err)
				}
				_, err = fetchAllRows(rows)
				if err != nil {
					panic(err)
				}
				//log.Println(data)
				rows.Close()
			}
		}()
	}
	for i := 0; i < 10; i++ {
		go func() {
			for {
				rows, err := db.Query("select * from t where b=1")
				if err != nil {
					panic(err)
				}
				_, err = fetchAllRows(rows)
				if err != nil {
					panic(err)
				}
				//log.Println(data)
				rows.Close()
			}
		}()
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}
func fetchAllRows(rows *sql.Rows) ([][]string, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	rowsData := make([][]string, 0, 4)
	for rows.Next() {
		values := make([]interface{}, len(cols))
		for i := range values {
			var v sql.NullString
			values[i] = &v
		}
		err = rows.Scan(values...)
		if err != nil {
			return nil, err
		}
		rowData := make([]string, len(cols))
		for i, v := range values {
			rowData[i] = v.(*sql.NullString).String
		}
		rowsData = append(rowsData, rowData)
	}
	return rowsData, nil
}
