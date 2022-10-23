package experiment

import (
	"github.com/YouDecideIt/auto-index/context"
	"github.com/YouDecideIt/auto-index/operations"
	"github.com/YouDecideIt/auto-index/request"
	"github.com/YouDecideIt/auto-index/study"
	"github.com/YouDecideIt/auto-index/utils"
	"github.com/pingcap/log"
	"github.com/pingcap/tidb-dashboard/pkg/apiserver/topsql"
	"go.uber.org/zap"
	"strings"
	"time"
)

func Experiment(ctx context.Context, endpoint *operations.BClusterEndpoint, oldItem topsql.SummaryItem, indexes []request.Index) (float64, error) {
	bdb := utils.OpenDatabase(endpoint.SQLEndpoint)
	defer utils.CloseDatabase(bdb)

	err := request.ApplyIndex(bdb, indexes)
	if err != nil {
		log.Error("bdb: failed to apply index", zap.Error(err))
		return 0, err
	}

	time.Sleep(ctx.Cfg.EvaluateConfig.WaitAfterApply)

	infos, err := study.GetCurrentTopSQLInfo(endpoint.NgmEndpoint)

	found := false
	oldCPUTimeMs := oldItem.CPUTimeMs
	var newCPUTimeMs uint64

	for instance, summarys := range infos {
		if instance.InstanceType != "tidb" {
			continue
		}
		for _, item := range summarys.Data {
			if strings.Contains(item.SQLText, "mysql") {
				continue
			}
			study.Instantiate(&item)
			if item.SQLText == oldItem.SQLText {
				found = true
				newCPUTimeMs = item.CPUTimeMs
			}
		}
	}

	if err != nil {
		log.Error("bdb: failed to study", zap.Error(err))
		return 0, err
	}

	if !found {

	}

	return float64((oldCPUTimeMs - newCPUTimeMs) / oldCPUTimeMs), nil
}
