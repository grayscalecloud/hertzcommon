package main

import (
	"github.com/grayscalecloud/hertzcommon/hdserver"
	"github.com/grayscalecloud/hertzcommon/model"
)

func main() {
	hertzCfg := &model.Hertz{
		Service:       "user-service",
		Address:       ":8080",
		EnableSwagger: true,
	}
	monitorCfg := &model.Monitor{
		Prometheus: model.Prometheus{
			Enable:      true,
			MetricsPort: 9090,
		},
		OTel: model.OTel{
			Enable:   true,
			Endpoint: "localhost:4317",
		},
	}
	h := hdserver.NewHdServer(hertzCfg, monitorCfg)

	h.Spin()
}
