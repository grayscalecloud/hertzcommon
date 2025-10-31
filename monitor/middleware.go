package monitor

import (
    "context"

    "github.com/cloudwego/hertz/pkg/app"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/trace"

    "github.com/grayscalecloud/hertzcommon/pkg/ctxx"
)

// AttachTenantAttributes 作为最后一个中间件执行，在业务与其他中间件运行后，把租户/商户/用户信息写入当前 Span
func AttachTenantAttributes() app.HandlerFunc {
    return func(ctx context.Context, c *app.RequestContext) {
        // 先执行后续中间件与业务逻辑，确保上下文中已写入 tenant/merchant/user
        c.Next(ctx)

        span := trace.SpanFromContext(ctx)
        if span == nil || !span.SpanContext().IsValid() {
            return
        }

        tid := ctxx.GetTenantID(ctx)
        mid := ctxx.GetMerchantID(ctx)
        uid := ctxx.GetUserID(ctx)

        if tid != "" {
            span.SetAttributes(attribute.String("tenant.id", tid))
        } else {
            span.SetAttributes(attribute.String("tenant.id.status", "没有租户信息"))
        }

        if mid != "" {
            span.SetAttributes(attribute.String("merchant.id", mid))
        } else {
            span.SetAttributes(attribute.String("merchant.id.status", "没有商户信息"))
        }

        span.SetAttributes(attribute.String("user.id", uid))
    }
}


