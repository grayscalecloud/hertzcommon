package kvconfig

import (
	"os"
	"testing"
)

// 仅保留：Kitex 使用场景下的最小化用例（初始化 + 获取配置）。
func TestKitexStyle_InitAndGet(t *testing.T) {
	server := os.Getenv("NACOS_SERVER_ADDR")
	ns := os.Getenv("NACOS_NAMESPACE_ID")
	group := os.Getenv("NACOS_GROUP")
	user := os.Getenv("NACOS_USERNAME")
	pass := os.Getenv("NACOS_PASSWORD")

	if server == "" || ns == "" || group == "" {
		t.Skip("未设置 NACOS_* 环境变量，跳过 Kitex 集成测试")
		return
	}

	// 使用新的 API
	options := &ConfigFactoryOptions{
		ServerAddr:  server,
		NamespaceId: ns,
		Group:       group,
		Username:    user,
		Password:    pass,
		ConfigType:  ConfigTypeNacos,
	}

	factory := NewConfigFactory(options)
	err := factory.InitNacosClientWithParamsOrEnv(server, ns, group, user, pass)
	if err != nil {
		t.Fatalf("初始化 Nacos 客户端失败: %v", err)
	}
	defer factory.Close()

	// 按 Kitex 约定读取 common 配置
	if _, err := factory.GetCommonConfig(group); err != nil {
		t.Logf("获取 common 配置失败（可能未创建）: %v", err)
	}
}

// 测试外部传入参数的初始化方式
func TestKitexStyle_InitWithParams(t *testing.T) {
	// 方式1：直接传入参数，环境变量优先
	options := &ConfigFactoryOptions{
		ServerAddr:  "115.190.176.125:8848",
		NamespaceId: "6a4a9a5b-bf1b-4e3c-8c0d-56cc393a616a",
		Group:       "DEFAULT_GROUP",
		Username:    "nacos",
		Password:    "nacos",
		ConfigType:  ConfigTypeNacos,
	}

	// 创建配置工厂并初始化
	factory := NewConfigFactory(options)
	err := factory.InitNacosClientWithParamsOrEnv(options.ServerAddr, options.NamespaceId, options.Group, options.Username, options.Password)
	if err != nil {
		t.Logf("初始化 Nacos 客户端失败（可能服务器不可达）: %v", err)
		return
	}
	defer factory.Close()

	// 测试获取配置
	if _, err := factory.GetCommonConfig(options.Group); err != nil {
		t.Logf("获取 common 配置失败（可能未创建）: %v", err)
	}
}

// 测试全局配置工厂的外部参数初始化
func TestGlobalConfigFactoryWithParams(t *testing.T) {
	// 方式2：使用全局配置工厂，环境变量优先
	options := &ConfigFactoryOptions{
		ServerAddr:  "115.190.176.125:8848",
		NamespaceId: "6a4a9a5b-bf1b-4e3c-8c0d-56cc393a616a",
		Group:       "DEFAULT_GROUP",
		Username:    "nacos",
		Password:    "nacos",
		ConfigType:  ConfigTypeNacos,
	}

	err := InitGlobalConfigFactoryWithNacos(options)
	if err != nil {
		t.Logf("初始化全局 Nacos 配置工厂失败（可能服务器不可达）: %v", err)
		return
	}

	// 测试通过全局工厂获取配置
	if _, err := GetCommonConfigGlobal(options.Group); err != nil {
		t.Logf("通过全局工厂获取 common 配置失败（可能未创建）: %v", err)
	}
}

// 测试 Consul 配置工厂
func TestConsulConfigFactory(t *testing.T) {
	// 测试 Consul 配置工厂
	options := &ConfigFactoryOptions{
		ServerAddr:  "127.0.0.1:8500",
		NamespaceId: "test-namespace",
		Group:       "DEFAULT_GROUP",
		Username:    "",
		Password:    "",
		ConfigType:  ConfigTypeConsul,
	}

	factory := NewConfigFactory(options)
	err := factory.InitConsulClientWithParamsOrEnv(options.ServerAddr, options.NamespaceId, options.Group, options.Username, options.Password)
	if err != nil {
		t.Logf("初始化 Consul 客户端失败（可能服务器不可达）: %v", err)
		return
	}
	defer factory.Close()

	// 测试获取配置
	if _, err := factory.GetCommonConfig(options.Group); err != nil {
		t.Logf("获取 common 配置失败（可能未创建）: %v", err)
	}
}
