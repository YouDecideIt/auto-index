package request

import (
	"database/sql"
	"github.com/YouDecideIt/auto-index/context"
	"github.com/pingcap/log"
	"go.uber.org/zap"
	"strconv"
	"strings"
)

type Index struct {
	TableName string
	Cols      []string
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

func WhatIf(ctx context.Context, sql string) (index []Index, rate float64, err error) {
	sql = "explain format='verbose' " + sql
	//log.Info("what if", zap.Any("sql", sql))
	rows0, err := ctx.DB.Query(sql)
	if err != nil {
		return
	}
	data0, err := fetchAllRows(rows0)
	if err != nil {
		return
	}
	rows0.Close()

	//log.Info("what if", zap.Any("sql", sql))
	_, err = ctx.DB.Exec("set @@try_best_index=1")
	if err != nil {
		return
	}

	//log.Info("what if", zap.Any("sql", sql))
	rows1, err := ctx.DB.Query(sql)
	if err != nil {
		return
	}
	data1, err := fetchAllRows(rows1)
	if err != nil {
		return
	}
	rows1.Close()

	log.Debug("what if", zap.Any("original cost", data0[0][2]))
	log.Debug("what if", zap.Any("with index cost", data1[0][2]))
	//log.Info(data0[0][2], data1[0][2])

	estCost0, err := strconv.ParseFloat(data0[0][2], 64)
	if err != nil {
		return
	}
	estCost1, err := strconv.ParseFloat(data1[0][2], 64)
	if err != nil {
		return
	}
	rate = (estCost0 - estCost1) / estCost0

	// fetch all indexes
	for i := 0; i < len(data1); i++ {
		accessObject := data1[i][4] // table:t1, index:a(a, b)
		if strings.Contains(accessObject, "index:") {
			table := parseTable(accessObject)
			cols := parseCols(accessObject)
			index = append(index, Index{TableName: table, Cols: cols})
		}
	}
	return
}

func parseTable(accessObject string) string {
	firstColon := strings.Index(accessObject, ":")
	firstComma := strings.Index(accessObject, ",")
	table := accessObject[firstColon+1 : firstComma]
	return table
}

func parseCols(accessObject string) []string {
	leftBracket := strings.Index(accessObject, "(")
	rightBracket := strings.Index(accessObject, ")")
	cols := strings.Split(accessObject[leftBracket+1:rightBracket], ",")
	for i := range cols {
		cols[i] = strings.TrimSpace(cols[i])
	}
	return cols
}
