package main

import (
	"github.com/grayscalecloud/hertzcommon/hdmodel"
	"github.com/grayscalecloud/hertzcommon/hdserver"
)

func main() {
	hertzCfg := &hdmodel.Hertz{
		Service:       "user-service",
		Address:       ":8080",
		EnableSwagger: true,
	}
	monitorCfg := &hdmodel.Monitor{
		Prometheus: hdmodel.Prometheus{
			Enable:      true,
			MetricsPort: 9090,
		},
		OTel: hdmodel.OTel{
			Enable:   true,
			Endpoint: "localhost:4317",
		},
	}
	h := hdserver.NewHdServer(hertzCfg, monitorCfg)

	h.Spin()
}
