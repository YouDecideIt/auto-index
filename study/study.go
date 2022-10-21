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
			return
		}
		log.Info("get top-sql success", zap.Any("resp", resp))
	}
}
