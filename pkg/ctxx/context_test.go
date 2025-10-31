package ctxx

import (
	"context"
	"testing"

	"github.com/bytedance/gopkg/cloud/metainfo"
)

func TestSetMetaInfo(t *testing.T) {
	ctx := context.Background()

	// 测试设置租户ID
	ctx = SetMetaInfo(ctx, TenantKey, "tenant123")

	// 验证从 metainfo 中获取
	if value, ok := metainfo.GetValue(ctx, TenantKey); !ok || value != "tenant123" {
		t.Errorf("Expected tenant123, got %s (ok: %t)", value, ok)
	}

	// 验证从 context 中获取
	if value := GetTenantID(ctx); value != "tenant123" {
		t.Errorf("Expected tenant123, got %s", value)
	}
}

func TestGetMetaInfoWithFallback(t *testing.T) {
	ctx := context.Background()

	// 设置 fallback 键名
	ctx = metainfo.WithValue(ctx, "tenant_id", "fallback_tenant")

	// 测试获取主键（不存在）
	if value := GetMetaInfo(ctx, TenantKey); value != "fallback_tenant" {
		t.Errorf("Expected fallback_tenant, got %s", value)
	}

	// 测试自定义 fallback
	ctx = metainfo.WithValue(ctx, "custom_key", "custom_value")
	if value := GetMetaInfoWithFallback(ctx, "primary_key", "custom_key", "another_key"); value != "custom_value" {
		t.Errorf("Expected custom_value, got %s", value)
	}
}

func TestGetMetaInfoWithFallbackPriority(t *testing.T) {
	ctx := context.Background()

	// 设置多个键名，测试优先级
	ctx = metainfo.WithValue(ctx, "tenant", "low_priority")
	ctx = metainfo.WithValue(ctx, "TENANT_ID", "high_priority")

	// 应该优先获取高优先级的键
	if value := GetMetaInfo(ctx, TenantKey); value != "high_priority" {
		t.Errorf("Expected high_priority, got %s", value)
	}
}

func TestGetAllMetaInfo(t *testing.T) {
	ctx := context.Background()

	// 设置多个 metainfo 值
	ctx = SetMetaInfo(ctx, TenantKey, "tenant123")
	ctx = SetMetaInfo(ctx, UserKey, "user456")
	ctx = SetMetaInfo(ctx, RequestKey, "req789")

	allInfo := GetAllMetaInfo(ctx)

	expected := map[string]string{
		TenantKey:  "tenant123",
		UserKey:    "user456",
		RequestKey: "req789",
	}

	for key, expectedValue := range expected {
		if actualValue, exists := allInfo[key]; !exists || actualValue != expectedValue {
			t.Errorf("Expected %s=%s, got %s=%s", key, expectedValue, key, actualValue)
		}
	}
}

func TestSetMultipleMetaInfo(t *testing.T) {
	ctx := context.Background()

	values := map[string]string{
		TenantKey:   "tenant123",
		UserKey:     "user456",
		MerchantKey: "merchant789",
	}

	ctx = SetMultipleMetaInfo(ctx, values)

	for key, expectedValue := range values {
		if actualValue := GetMetaInfo(ctx, key); actualValue != expectedValue {
			t.Errorf("Expected %s=%s, got %s=%s", key, expectedValue, key, actualValue)
		}
	}
}

func TestCopyMetaInfo(t *testing.T) {
	sourceCtx := context.Background()
	targetCtx := context.Background()

	// 在源 context 中设置值
	sourceCtx = SetMetaInfo(sourceCtx, TenantKey, "tenant123")
	sourceCtx = SetMetaInfo(sourceCtx, UserKey, "user456")

	// 复制到目标 context
	targetCtx = CopyMetaInfo(sourceCtx, targetCtx)

	// 验证复制结果
	if value := GetTenantID(targetCtx); value != "tenant123" {
		t.Errorf("Expected tenant123, got %s", value)
	}

	if value := GetUserID(targetCtx); value != "user456" {
		t.Errorf("Expected user456, got %s", value)
	}
}

func TestHasMetaInfo(t *testing.T) {
	ctx := context.Background()

	// 测试不存在的键
	if HasMetaInfo(ctx, TenantKey) {
		t.Error("Expected false for non-existent key")
	}

	// 设置值后测试
	ctx = SetMetaInfo(ctx, TenantKey, "tenant123")
	if !HasMetaInfo(ctx, TenantKey) {
		t.Error("Expected true for existing key")
	}
}

func TestGetMetaInfoOrDefault(t *testing.T) {
	ctx := context.Background()

	// 测试不存在的键，应该返回默认值
	if value := GetMetaInfoOrDefault(ctx, TenantKey, "default_tenant"); value != "default_tenant" {
		t.Errorf("Expected default_tenant, got %s", value)
	}

	// 设置值后测试
	ctx = SetMetaInfo(ctx, TenantKey, "tenant123")
	if value := GetMetaInfoOrDefault(ctx, TenantKey, "default_tenant"); value != "tenant123" {
		t.Errorf("Expected tenant123, got %s", value)
	}
}

func TestContextInfo(t *testing.T) {
	ctx := context.Background()

	// 设置所有信息
	ctx = SetMetaInfo(ctx, TenantKey, "tenant123")
	ctx = SetMetaInfo(ctx, UserKey, "user456")
	ctx = SetMetaInfo(ctx, RequestKey, "req789")
	ctx = SetMetaInfo(ctx, MerchantKey, "merchant101")

	info := GetContextInfo(ctx)

	if info.TenantID != "tenant123" {
		t.Errorf("Expected tenant123, got %s", info.TenantID)
	}
	if info.UserID != "user456" {
		t.Errorf("Expected user456, got %s", info.UserID)
	}
	if info.RequestID != "req789" {
		t.Errorf("Expected req789, got %s", info.RequestID)
	}
	if info.MerchantID != "merchant101" {
		t.Errorf("Expected merchant101, got %s", info.MerchantID)
	}
}
