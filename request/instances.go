package request

import (
	"encoding/json"
	ctx "github.com/YouDecideIt/auto-index/context"
	"github.com/YouDecideIt/auto-index/utils"
	"github.com/go-resty/resty/v2"
	"github.com/pingcap/log"
	"github.com/pingcap/tidb-dashboard/pkg/apiserver/topsql"
	"go.uber.org/zap"
)

func GetInstances(ctx ctx.Context) (topsql.InstanceResponse, error) {
	return getInstances(ctx, nil)
}

func GetInstancesWithTime(ctx ctx.Context, reqRaw topsql.GetInstancesRequest) (topsql.InstanceResponse, error) {
	req, err := utils.ObjectToMapStringString(reqRaw)
	if err != nil {
		return topsql.InstanceResponse{}, err
	}
	return getInstances(ctx, req)
}

func getInstances(ctx ctx.Context, req map[string]string) (topsql.InstanceResponse, error) {
	url := "http://" + ctx.Cfg.NgMonitorConfig.Address + "/topsql/v1/instances"
	log.Debug("get instances", zap.String("url", url))

	client := resty.New().SetDebug(false)

	resp, err := client.R().SetQueryParams(req).Get(url)
	if err != nil {
		log.Error("get top-sql failed", zap.Error(err))
		return topsql.InstanceResponse{}, err
	}

	respT := topsql.InstanceResponse{}
	err = json.Unmarshal(resp.Body(), &respT)
	if err != nil {
		return topsql.InstanceResponse{}, nil
	}
	log.Info("get top-sql success", zap.Any("resp", respT))
	return respT, nil
}
