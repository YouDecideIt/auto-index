package experiment

import (
	b_cluster "github.com/YouDecideIt/auto-index/b-cluster"
	"github.com/YouDecideIt/auto-index/context"
	"github.com/YouDecideIt/auto-index/request"
	"github.com/YouDecideIt/auto-index/study"
	"github.com/YouDecideIt/auto-index/utils"
	"github.com/pingcap/log"
	"github.com/pingcap/tidb-dashboard/pkg/apiserver/topsql"
	"go.uber.org/zap"
	"time"
)

func Experiment(ctx context.Context, endpoint *b_cluster.BClusterEndpoint, item topsql.SummaryItem, indexes []request.Index) (float64, error) {
	bdb := utils.OpenDatabase(endpoint.SQLEndpoint)
	defer utils.CloseDatabase(bdb)

	err := request.ApplyIndex(bdb, indexes)
	if err != nil {
		log.Error("bdb: failed to apply index", zap.Error(err))
		return 0, err
	}

	time.Sleep(ctx.Cfg.EvaluateConfig.WaitAfterApply)

	newItem, err := study.Study(endpoint.NgmEndpoint)
	if err != nil {
		log.Error("bdb: failed to study", zap.Error(err))
		return 0, err
	}

	return float64((item.CPUTimeMs - newItem.CPUTimeMs) / item.CPUTimeMs), nil
}
