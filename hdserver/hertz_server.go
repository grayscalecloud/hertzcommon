package hdserver

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/grayscalecloud/hertzcommon/model"
	"github.com/grayscalecloud/hertzcommon/monitor"
	hertzlogrus "github.com/hertz-contrib/obs-opentelemetry/logging/logrus"
	hertzotelprovider "github.com/hertz-contrib/obs-opentelemetry/provider"
	hertzoteltracing "github.com/hertz-contrib/obs-opentelemetry/tracing"
	oteltrace "go.opentelemetry.io/otel/trace"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func NewHdServer(hzCfg *model.Hertz, monitorCfg *model.Monitor) *server.Hertz {
	var opts []config.Option
	var cfg *hertzoteltracing.Config
	if monitorCfg.OTel.Enable {
		monitor.InitMtl(hzCfg.Service, monitorCfg)
		var tracer config.Option
		tracer, cfg = hertzoteltracing.NewServerTracer(
			hertzoteltracing.WithCustomResponseHandler(func(ctx context.Context, c *app.RequestContext) {
				c.Header("x-trace-id", oteltrace.SpanFromContext(ctx).SpanContext().TraceID().String())
			}))
		// 创建Hertz服务器
		opts = hertzInit(hzCfg, monitorCfg, tracer)
	} else {
		// 服务地址
		opts = append(opts, server.WithHostPorts(hzCfg.Address))

		opts = append(opts, server.WithHandleMethodNotAllowed(true))
	}

	h := server.New(opts...)

	registerMiddleware(h, cfg, hzCfg)

	return h
}

func hertzInit(hzCfg *model.Hertz, monitorCfg *model.Monitor, tracer config.Option) (opts []config.Option) {
	address := hzCfg.Address
	opts = append(opts, tracer)
	// 服务地址
	opts = append(opts, server.WithHostPorts(address))

	opts = append(opts, server.WithHandleMethodNotAllowed(true))
	// 服务发现
	if monitorCfg.OTel.Enable {
		_ = hertzotelprovider.NewOpenTelemetryProvider(
			hertzotelprovider.WithServiceName(hzCfg.Service),
			hertzotelprovider.WithExportEndpoint(monitorCfg.OTel.Endpoint),
			hertzotelprovider.WithSdkTracerProvider(monitor.TracerProvider),
			hertzotelprovider.WithEnableMetrics(false),
		)
		//defer p.Shutdown(context.Background())
	}

	return
}

// registerMiddleware 注册中间件
// 为Hertz服务器注册各种中间件，包括日志、pprof、gzip压缩、访问日志、恢复和CORS等
func registerMiddleware(h *server.Hertz, cfg *hertzoteltracing.Config, hzCfg *model.Hertz) {
	// log
	if cfg != nil {
		h.Use(hertzoteltracing.ServerMiddleware(cfg))
		logger := hertzlogrus.NewLogger()
		hlog.SetLogger(logger)
		hlog.SetLevel(logLevel(hzCfg.LogLevel))

		// 创建一个多写入器，同时写入控制台和文件
		var writers []io.Writer
		// 添加控制台输出
		writers = append(writers, os.Stdout)

		// 添加文件输出
		if hzCfg.LogFileName != "" {
			asyncWriter := &zapcore.BufferedWriteSyncer{
				WS: zapcore.AddSync(&lumberjack.Logger{
					Filename:   hzCfg.LogFileName,
					MaxSize:    hzCfg.LogMaxSize,
					MaxBackups: hzCfg.LogMaxBackups,
					MaxAge:     hzCfg.LogMaxAge,
				}),
				FlushInterval: time.Minute,
			}
			writers = append(writers, asyncWriter)
			//h.OnShutdown = append(h.OnShutdown, func(ctx context.Context) {
			//	asyncWriter.Sync()
			//})
		}

		// 设置多写入器
		hlog.SetOutput(io.MultiWriter(writers...))
	}

}
func logLevel(level string) hlog.Level {
	switch level {
	case "trace":
		return hlog.LevelTrace
	case "debug":
		return hlog.LevelDebug
	case "info":
		return hlog.LevelInfo
	case "notice":
		return hlog.LevelNotice
	case "warn":
		return hlog.LevelWarn
	case "error":
		return hlog.LevelError
	case "fatal":
		return hlog.LevelFatal
	default:
		return hlog.LevelInfo
	}
}
