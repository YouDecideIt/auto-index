package request

import (
	"database/sql"
	"fmt"
	"github.com/YouDecideIt/auto-index/context"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"testing"
)

func TestWhatif(t *testing.T) {
	db, err := sql.Open("mysql", "root:@tcp(localhost:4000)/")
	if err != nil {
		log.Println(err)
		t.Skip(err)
	}
	defer db.Close()
	ctx := context.Context{
		DB: db,
	}
	db.Exec("create database if not exists test")
	db.Exec("use test")
	db.Exec("drop table if exists t")
	db.Exec("create table t(a int, b int, key(a) invisible, key(b) invisible, key(a,b) invisible)")
	db.Exec("insert into t values(1, 1), (2, 2), (3, 3)")
	for i := 0; i < 10; i++ {
		db.Exec("insert into t select * from t")
	}
	db.Exec("analyze table t")

	indexes, rate, err := WhatIf(ctx, "select * from t where a=1 and b=2")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(indexes, rate)
	if len(indexes) != 1 || len(indexes[0].Cols) != 2 {
		t.Fatal("wrong indexes")
	}
	if indexes[0].TableName != "t" || indexes[0].Cols[0] != "a" || indexes[0].Cols[1] != "b" {
		t.Fatal("wrong index")
	}
	if rate < 0.9 {
		t.Fatal("wrong rate")
	}
}
