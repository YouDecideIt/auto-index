package study

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/pingcap/log"
	"github.com/pingcap/tidb-dashboard/pkg/apiserver/topsql"
	"go.uber.org/zap"
)

// Study get the top-sql information from ng-monitor.
func Study(address string) {
	//a := topsql.Service{params: topsql.ServiceParams{NgmProxy: &topsql.NgmProxy{}}, FeatureTopSQL: nil}
	//
	//clientCfg := client.Config{Endpoints: []string{address}}
	//cli, _ := client.New(clientCfg)
	//ngm, _ := utils.NewNgmProxy(nil, cli)
	////params := topsql.ServiceParams{NgmProxy: ngm}
	//service := topsql.Service{FeatureTopSQL: nil}
	//service.GetSummary(nil)

	client := resty.New().SetDebug(true)
	log.Debug("get summary", zap.String("address", address))

	resp, err := client.R().Get("http://" + address + "/topsql/v1/instances")
	if err != nil {
		log.Error("get top-sql failed", zap.Error(err))
		return
	}

	respT := topsql.InstanceResponse{}
	err = json.Unmarshal(resp.Body(), &respT)
	if err != nil {
		return
	}
	log.Info("get top-sql success", zap.Any("resp", respT))

	//req := topsql.GetInstancesRequest{Start: "1", End: "2"}
	//client.R().SetQueryParams(JSONMethod(req)).Get("http://" + address + "/topsql/v1/instances")

	//_ = resp
}

func JSONMethod(content interface{}) map[string]string {
	var name map[string]string
	if marshalContent, err := json.Marshal(content); err != nil {
		fmt.Println(err)
	} else {
		d := json.NewDecoder(bytes.NewReader(marshalContent))
		d.UseNumber() // 设置将float64转为一个number
		if err := d.Decode(&name); err != nil {
			fmt.Println(err)
		} else {
			for k, v := range name {
				name[k] = string(v)
			}
		}
	}
	return name
}
