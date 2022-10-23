package utils

import (
	"bytes"
	sysctx "context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/pingcap/log"
	"go.uber.org/zap"
	"time"
)

func ObjectToMapStringString(content interface{}) (map[string]string, error) {
	var name map[string]string
	marshalContent, err := json.Marshal(content)
	if err != nil {
		return nil, err
	}
	d := json.NewDecoder(bytes.NewReader(marshalContent))
	d.UseNumber()
	if err := d.Decode(&name); err != nil {
		return nil, err
	}
	return name, nil
}

func OpenDatabase(endpoint string) *sql.DB {
	now := time.Now()
	log.Info("setting up database")
	defer func() {
		log.Info("init database done", zap.Duration("in", time.Since(now)))
	}()

	db, err := sql.Open("mysql", fmt.Sprintf("root@tcp(%s)/test", endpoint))
	{
		if err != nil {
			log.Fatal("failed to open db", zap.Error(err))
		}
		sqlCtx, cancel := sysctx.WithTimeout(sysctx.Background(), time.Duration(5)*time.Second)
		defer cancel()
		err = db.PingContext(sqlCtx)
		if err != nil {
			log.Fatal("failed to open db", zap.Error(err))
		}
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	return db
}

func CloseDatabase(db *sql.DB) {
	now := time.Now()
	log.Info("closing database")
	defer func() {
		log.Info("close database done", zap.Duration("in", time.Since(now)))
	}()

	if err := db.Close(); err != nil {
		log.Warn("failed to close database", zap.Error(err))
	}
}
