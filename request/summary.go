package request

import (
	"encoding/json"
	"github.com/YouDecideIt/auto-index/config"
	"github.com/YouDecideIt/auto-index/utils"
	"github.com/go-resty/resty/v2"
	"github.com/pingcap/log"
	"github.com/pingcap/tidb-dashboard/pkg/apiserver/topsql"
	"go.uber.org/zap"
)

func GetSummary(reqRaw topsql.GetSummaryRequest) (topsql.SummaryResponse, error) {
	req, err := utils.ObjectToMapStringString(reqRaw)
	if err != nil {
		return topsql.SummaryResponse{}, err
	}

	url := "http://" + config.GlobalConfig.NgMonitorConfig.Address + "/topsql/v1/summary"
	log.Debug("get summary", zap.String("url", url), zap.Any("req", req))

	client := resty.New().SetDebug(false)

	resp, err := client.R().SetQueryParams(req).Get(url)
	if err != nil {
		log.Error("get top-sql failed", zap.Error(err))
		return topsql.SummaryResponse{}, err
	}

	respT := topsql.SummaryResponse{}
	err = json.Unmarshal(resp.Body(), &respT)
	if err != nil {
		return topsql.SummaryResponse{}, nil
	}
	log.Info("get top-sql success", zap.Any("resp", respT))
	return respT, nil
}
