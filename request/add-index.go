package request

import (
	"github.com/YouDecideIt/auto-index/context"
	"github.com/pingcap/log"
	"go.uber.org/zap"
	"strconv"
	"sync/atomic"
)

// atomic counter for index name
var ops uint64

// MergeSql merge the tableName and colNames, return the add index sql
func MergeSql(tableName string, colName []string) string {
	atomic.AddUint64(&ops, 1)
	indexName := "AdvisorIndex" + strconv.FormatUint(ops, 10)

	sql := "alter table " + tableName + " add index " + indexName + "("
	for i, col := range colName {
		sql += col
		if i != len(colName)-1 {
			sql += ","
		}
	}
	sql += ");"
	log.Debug("merge sql succeed", zap.String("sql", sql))
	return sql
}

func ApplyIndex(ctx context.Context, tableName string, colName []string) error {
	sql := MergeSql(tableName, colName)

	_, err := ctx.DB.Exec(sql)
	if err != nil {
		log.Error("apply index failed", zap.Error(err))
		return err
	}
	log.Error("apply index succeed", zap.String("sql", sql))
	return nil
}
