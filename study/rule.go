package study

import (
	"fmt"
	"github.com/pingcap/log"
	"github.com/pingcap/tidb-dashboard/pkg/apiserver/topsql"
	"go.uber.org/zap"
	"strings"
)

// Study studies the topsql and return the need-optimize sqls
func Study(ngmEndpoint string) ([]topsql.SummaryItem, error) {
	info, err := GetCurrentTopSQLInfo(ngmEndpoint)
	if err != nil {
		return []topsql.SummaryItem{}, err
	}
	hasTiDB := false
	topOne := topsql.SummaryItem{}
	topTwo := topsql.SummaryItem{}
	for instance, summarys := range info {
		if instance.InstanceType != "tidb" {
			continue
		}
		if len(summarys.Data) == 0 {
			continue
		}
		hasTiDB = true
		for _, item := range summarys.Data {
			if strings.Contains(item.SQLText, "mysql") {
				continue
			}
			if topOne.SQLText == "" || topOne.CPUTimeMs < item.CPUTimeMs {
				topOne = item
			} else {
				if topTwo.SQLText == "" || topTwo.CPUTimeMs < item.CPUTimeMs {
					topTwo = item
				}
			}
		}

	}
	if !hasTiDB {
		log.Error("no tidb instance found")
		return []topsql.SummaryItem{}, fmt.Errorf("no tidb instance")
	}
	log.Debug("study succeed", zap.Any("find sql", topOne))
	log.Debug("study succeed", zap.Any("find sql", topOne))

	Instantiate(&topOne)
	Instantiate(&topTwo)
	return []topsql.SummaryItem{topOne, topTwo}, nil
}

func Instantiate(item *topsql.SummaryItem) {
	item.SQLText = strings.Replace(item.SQLText, "?", "1", -1)
	log.Debug("Instantiate succeed", zap.Any("sql", item.SQLText))
}

//func StudySQL(ctx context.Context) (string, error) {
//	item, err := Study(ctx.Cfg.NgMonitorConfig.Address)
//	if err != nil {
//		return "", err
//	}
//
//	return item.SQLText, nil
//}
