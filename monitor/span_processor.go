package monitor

import (
	"context"

	"github.com/grayscalecloud/hertzcommon/pkg/ctxx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace"
)

// 实现 SpanProcessor
type tenantIDProcessor struct{ next trace.SpanProcessor }

func NewTenantIDProcessor(next trace.SpanProcessor) trace.SpanProcessor {
	return &tenantIDProcessor{next: next}
}

func (p *tenantIDProcessor) OnStart(ctx context.Context, s trace.ReadWriteSpan) {
	// 从 context 中获取 tenantID 和 merchantID
	tid := ctxx.GetTenantID(ctx)
	mid := ctxx.GetMerchantID(ctx)
	uid := ctxx.GetUserID(ctx)

	// 添加 tenantID
	if tid != "" {
		s.SetAttributes(attribute.String("tenant.id", tid))
	} else {
		// 测试用：即使 tenantID 为空，也设置一个标记，用于调试
		s.SetAttributes(attribute.String("tenant.id.status", "没有租户信息"))
	}

	// 添加 merchantID
	if mid != "" {
		s.SetAttributes(attribute.String("merchant.id", mid))
	} else {
		// 测试用：即使 merchantID 为空，也设置一个标记，用于调试
		s.SetAttributes(attribute.String("merchant.id.status", "没有商户信息"))
	}
	s.SetAttributes(attribute.String("user.id", uid))
}

func (p *tenantIDProcessor) Shutdown(ctx context.Context) error   { return p.next.Shutdown(ctx) }
func (p *tenantIDProcessor) ForceFlush(ctx context.Context) error { return p.next.ForceFlush(ctx) }
func (p *tenantIDProcessor) OnEnd(s trace.ReadOnlySpan)           { p.next.OnEnd(s) }
