package study

import (
	"github.com/YouDecideIt/auto-index/request"
	"github.com/pingcap/log"
	"github.com/pingcap/tidb-dashboard/pkg/apiserver/topsql"
	"go.uber.org/zap"
	"strconv"
	"time"
)

// Study get the top-sql information from ng-monitor.
func Study() {
	now := time.Now()
	aHourAgo := now.Add(-time.Hour)

	instanceRequest := topsql.GetInstancesRequest{
		Start: strconv.FormatInt(aHourAgo.Unix(), 10),
		End:   strconv.FormatInt(now.Unix(), 10),
	}

	resp, err := request.GetInstancesWithTime(instanceRequest)
	if err != nil {
		return
	}
	log.Info("get top-sql success", zap.Any("resp", resp))

	//tidbInstance := ""
	//for _, instance := range resp.Data {
	//	if instance.InstanceType == "tidb" {
	//		tidbInstance = instance.Instance
	//		break
	//	}
	//}
	//
	//if tidbInstance == "" {
	//	log.Info("no tidb instance found")
	//	return
	//}

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
			return
		}
		log.Info("get summary succeed", zap.String("instance", instance.Instance))
		for i, summary := range resp.Data {
			log.Info("summary", zap.Int("top", i+1),
				zap.Any("sql text", summary.SQLText),
				zap.Any("CPUTimeMs", summary.CPUTimeMs),
				zap.Any("ExecCountPerSec", summary.ExecCountPerSec),
				zap.Any("DurationPerExecMs", summary.DurationPerExecMs),
				zap.Any("ScanRecordsPerSec", summary.ScanRecordsPerSec),
				zap.Any("ScanIndexesPerSec", summary.ScanIndexesPerSec),
			)
		}

		//log.Info("get top-sql success", zap.Any("resp", resp))
	}
}
