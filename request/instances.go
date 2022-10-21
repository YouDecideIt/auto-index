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

func GetInstances() (topsql.InstanceResponse, error) {
	return getInstances(nil)
}

func GetInstancesWithTime(req topsql.GetInstancesRequest) (topsql.InstanceResponse, error) {
	reqq, err := utils.ObjectToMapStringString(req)
	if err != nil {
		return topsql.InstanceResponse{}, err
	}
	return getInstances(reqq)
}

func getInstances(req map[string]string) (topsql.InstanceResponse, error) {
	url := "http://" + config.GlobalConfig.NgMonitorConfig.Address + "/topsql/v1/instances"
	log.Debug("get instances", zap.String("url", url))

	client := resty.New().SetDebug(true)

	resp, err := client.R().Get(url)
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
