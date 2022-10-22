package study

import (
	"github.com/YouDecideIt/auto-index/request"
	"github.com/pingcap/log"
	"github.com/pingcap/tidb-dashboard/pkg/apiserver/topsql"
	"go.uber.org/zap"
	"strconv"
	"time"
)

// GetCurrentTopSQLInfo get the top-sql information from ng-monitor.
func GetCurrentTopSQLInfo() (map[topsql.InstanceItem]topsql.SummaryResponse, error) {
	now := time.Now()
	aHourAgo := now.Add(-time.Hour)

	instanceRequest := topsql.GetInstancesRequest{
		Start: strconv.FormatInt(aHourAgo.Unix(), 10),
		End:   strconv.FormatInt(now.Unix(), 10),
	}

	resp, err := request.GetInstancesWithTime(instanceRequest)
	if err != nil {
		return nil, err
	}
	log.Info("get top-sql success", zap.Any("resp", resp))

	result := make(map[topsql.InstanceItem]topsql.SummaryResponse)

	for _, instance := range resp.Data {
		summaryRequest := topsql.GetSummaryRequest{
			Start: strconv.FormatInt(aHourAgo.Unix(), 10),
			End:   strconv.FormatInt(now.Unix(), 10),
			//Instance: tidbInstance,
			Instance:     instance.Instance,
			InstanceType: instance.InstanceType,
			Top:          "5",
			Window:       "21s",
		}

		resp, err := request.GetSummary(summaryRequest)
		if err != nil {
			log.Warn("get summary failed", zap.Error(err))
			return nil, err
		}
		log.Info("get summary succeed", zap.String("instance", instance.Instance))
		for i, summary := range resp.Data {
			log.Debug("summary", zap.Int("top", i+1),
				zap.Any("sql text", summary.SQLText),
				zap.Any("CPUTimeMs", summary.CPUTimeMs),
				zap.Any("ExecCountPerSec", summary.ExecCountPerSec),
				zap.Any("DurationPerExecMs", summary.DurationPerExecMs),
				zap.Any("ScanRecordsPerSec", summary.ScanRecordsPerSec),
				zap.Any("ScanIndexesPerSec", summary.ScanIndexesPerSec),
				zap.Any("IsOther", summary.IsOther),
			)
		}

		result[instance] = resp
		//log.Info("get top-sql success", zap.Any("resp", resp))
	}

	return result, nil
}
