package study

import (
	"fmt"
	"github.com/YouDecideIt/auto-index/context"
	"github.com/pingcap/log"
	"github.com/pingcap/tidb-dashboard/pkg/apiserver/topsql"
	"go.uber.org/zap"
)

// Study studies the topsql and return the need-optimize sqls
func Study(ctx context.Context) (topsql.SummaryItem, error) {
	info, err := GetCurrentTopSQLInfo(ctx)
	if err != nil {
		return topsql.SummaryItem{}, err
	}
	hasTiDB := false
	topOne := topsql.SummaryItem{}
	for instance, summarys := range info {
		if instance.InstanceType != "tidb" {
			continue
		}
		hasTiDB = true
		if topOne.SQLText == "" || topOne.CPUTimeMs < summarys.Data[0].CPUTimeMs {
			topOne = summarys.Data[0]
		}
	}
	if !hasTiDB {
		return topsql.SummaryItem{}, fmt.Errorf("no tidb instance")
	}
	log.Debug("study succeed", zap.Any("find sql", topOne))
	return topOne, nil
}

func StudySQL(ctx context.Context) (string, error) {
	item, err := Study(ctx)
	if err != nil {
		return "", err
	}

	return item.SQLText, nil
}
