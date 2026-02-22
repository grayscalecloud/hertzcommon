// Copyright 2024 CloudWeGo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package monitor

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/route"
	"github.com/grayscalecloud/hertzcommon/hdmodel"
	"github.com/grayscalecloud/hertzcommon/utils"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var Reg *prometheus.Registry

func initMetric(serverName string, cfg *hdmodel.Monitor) route.CtxCallback {
	if cfg.Prometheus.Enable {
		hlog.Info("开启Prometheus监控")
	} else {
		hlog.Info("未开启Prometheus监控")
		return func(ctx context.Context) {}
	}
	Reg = prometheus.NewRegistry()
	Reg.MustRegister(collectors.NewGoCollector())
	Reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	// 解析Nacos服务器地址和端口
	host, port, err := net.SplitHostPort(cfg.Registry.RegistryAddress)
	if err != nil {
		hlog.Error("解析Nacos服务器地址失败:", err)
		return func(ctx context.Context) {}
	}
	portInt, _ := strconv.Atoi(port)

	// Nacos配置
	sc := []constant.ServerConfig{
		*constant.NewServerConfig(host, uint64(portInt)),
	}

	cc := constant.ClientConfig{
		NamespaceId:         cfg.Registry.NamespaceId,
		Username:            cfg.Registry.Username,
		Password:            cfg.Registry.Password,
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "/tmp/nacos/log",
		CacheDir:            "/tmp/nacos/cache",
		LogLevel:            "info",
	}

	// 创建Nacos客户端
	client, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		},
	)
	if err != nil {
		hlog.Error("创建Nacos客户端失败:", err)
		return func(ctx context.Context) {}
	}

	// 获取本地IP
	localIp := utils.MustGetLocalIPv4()
	metricsAddr := fmt.Sprintf("%s:%d", localIp, cfg.Prometheus.MetricsPort)
	hlog.Info("Metrics地址:", metricsAddr)

	// 验证metrics地址是否合法
	_, err = net.ResolveTCPAddr("tcp", metricsAddr)
	if err != nil {
		hlog.Error("解析metrics地址失败:", err)
		return func(ctx context.Context) {}
	}

	// 注册服务到Nacos
	serviceName := serverName + "_metrics"
	_, err = client.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          localIp,
		Port:        uint64(cfg.Prometheus.MetricsPort),
		ServiceName: serviceName,
		GroupName:   cfg.Registry.Group,
		Weight:      10,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		Metadata: map[string]string{
			"service": serverName,
		},
	})
	if err != nil {
		hlog.Error("注册服务到Nacos失败:", err)
	}

	// 启动metrics服务
	http.Handle("/metrics", promhttp.HandlerFor(Reg, promhttp.HandlerOpts{}))
	go func() {
		err := http.ListenAndServe(metricsAddr, nil)
		if err != nil {
			hlog.Error("启动metrics服务失败:", err)
		}
	}()

	// 返回取消注册函数
	return func(ctx context.Context) {
		// 取消注册服务
		_, err = client.DeregisterInstance(vo.DeregisterInstanceParam{
			Ip:          localIp,
			Port:        uint64(cfg.Prometheus.MetricsPort),
			ServiceName: serviceName,
			GroupName:   cfg.Registry.Group,
			Ephemeral:   true,
		})
	}
}
