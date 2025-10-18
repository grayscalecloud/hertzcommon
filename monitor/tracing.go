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

	"github.com/cloudwego/hertz/pkg/route"
	"github.com/grayscalecloud/hertzcommon/model"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

var TracerProvider *tracesdk.TracerProvider

func InitTracing(serviceName string, cfg *model.Monitor) route.CtxCallback {
	exporter, err := otlptracegrpc.New(
		context.Background(),
		otlptracegrpc.WithEndpoint(cfg.OTel.Endpoint),
		// otlptracegrpc.WithEndpoint(conf.GetConf().Telemetry.Endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		panic(err)
	}
	processor := tracesdk.NewBatchSpanProcessor(exporter)

	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		res = resource.Default()
	}
	TracerProvider = tracesdk.NewTracerProvider(
		tracesdk.WithSpanProcessor(processor),
		tracesdk.WithResource(res),
	)
	otel.SetTracerProvider(TracerProvider)

	return route.CtxCallback(func(ctx context.Context) {
		// 移除 shutdown，应该在程序退出时执行
	})
}

func CleanupTracing(ctx context.Context) {
	if TracerProvider != nil {
		TracerProvider.Shutdown(ctx) //nolint:errcheck
	}
}
