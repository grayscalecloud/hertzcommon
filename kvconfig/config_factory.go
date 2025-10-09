package kvconfig

import (
	"fmt"
	"os"
	"strings"

	"github.com/grayscalecloud/hertzcommon/model"
)

// ConfigType 配置类型
type ConfigType string

const (
	ConfigTypeConsul ConfigType = "consul"
	ConfigTypeNacos  ConfigType = "nacos"
)

type ConfigFactoryOptions struct {
	ServerAddr  string
	NamespaceId string
	Group       string
	Username    string
	Password    string
	ConfigType  ConfigType
}

// ConfigFactory 配置工厂
type ConfigFactory struct {
	configType   ConfigType
	nacosClient  *NacosConfigClient
	consulClient *ConsulConfigClient
	options      *ConfigFactoryOptions
}

// NewConfigFactory 创建配置工厂
func NewConfigFactory(options *ConfigFactoryOptions) *ConfigFactory {
	return &ConfigFactory{
		configType:  options.ConfigType,
		nacosClient: nil,
		options:     options,
	}
}
func NewNacosConfigFactory(options *ConfigFactoryOptions) *ConfigFactory {
	return &ConfigFactory{
		configType:  ConfigTypeNacos,
		nacosClient: nil,
		options:     options,
	}
}
func NewConsulConfigFactory(options *ConfigFactoryOptions) *ConfigFactory {
	return &ConfigFactory{
		configType:  ConfigTypeConsul,
		nacosClient: nil,
		options:     options,
	}
}

func (f *ConfigFactory) SetOptions(options *ConfigFactoryOptions) {
	f.options = options
}

// SetConfigType 设置配置类型
func (f *ConfigFactory) SetConfigType(configType ConfigType) {
	f.configType = configType
}

// InitNacosClient 初始化 Nacos 客户端
func (f *ConfigFactory) InitNacosClient(serverAddrs []string, namespaceId, group string, username, password string) error {
	client, err := NewNacosConfigClient(serverAddrs, namespaceId, group, username, password)
	if err != nil {
		return fmt.Errorf("初始化 Nacos 客户端失败: %w", err)
	}
	f.nacosClient = client
	f.configType = ConfigTypeNacos
	return nil
}

// InitNacosClientFromEnv 从环境变量初始化 Nacos 客户端（不设置默认值，缺失则报错）
func (f *ConfigFactory) InitNacosClientFromEnv() error {
	serverAddr := os.Getenv("NACOS_SERVER_ADDR")
	namespaceId := os.Getenv("NACOS_NAMESPACE_ID")
	group := os.Getenv("NACOS_GROUP")
	username := os.Getenv("NACOS_USERNAME")
	password := os.Getenv("NACOS_PASSWORD")

	if serverAddr == "" || namespaceId == "" || group == "" {
		return fmt.Errorf("缺少必要的环境变量: NACOS_SERVER_ADDR / NACOS_NAMESPACE_ID / NACOS_GROUP")
	}

	serverAddrs := strings.Split(serverAddr, ",")
	for i, addr := range serverAddrs {
		serverAddrs[i] = strings.TrimSpace(addr)
	}

	client, err := NewNacosConfigClient(serverAddrs, namespaceId, group, username, password)
	if err != nil {
		return fmt.Errorf("初始化 Nacos 客户端失败: %w", err)
	}
	f.nacosClient = client
	f.configType = ConfigTypeNacos
	return nil
}

// InitNacosClientWithParamsOrEnv 优先使用环境变量，环境变量为空则使用传入参数；仍为空则报错
func (f *ConfigFactory) InitNacosClientWithParamsOrEnv(serverAddr, namespaceId, group, username, password string) error {
	// 优先使用环境变量
	envServerAddr := os.Getenv("NACOS_SERVER_ADDR")
	envNamespaceId := os.Getenv("NACOS_NAMESPACE_ID")
	envGroup := os.Getenv("NACOS_GROUP")
	envUsername := os.Getenv("NACOS_USERNAME")
	envPassword := os.Getenv("NACOS_PASSWORD")

	// 环境变量优先，为空则使用传入参数
	if envServerAddr != "" {
		serverAddr = envServerAddr
	}
	if envNamespaceId != "" {
		namespaceId = envNamespaceId
	}
	if envGroup != "" {
		group = envGroup
	}
	if envUsername != "" {
		username = envUsername
	}
	if envPassword != "" {
		password = envPassword
	}

	if serverAddr == "" || namespaceId == "" || group == "" {
		return fmt.Errorf("缺少必要的配置: serverAddr/namespaceId/group")
	}

	serverAddrs := strings.Split(serverAddr, ",")
	for i, addr := range serverAddrs {
		serverAddrs[i] = strings.TrimSpace(addr)
	}

	client, err := NewNacosConfigClient(serverAddrs, namespaceId, group, username, password)
	if err != nil {
		return fmt.Errorf("初始化 Nacos 客户端失败: %w", err)
	}
	f.nacosClient = client
	f.configType = ConfigTypeNacos
	return nil
}

func (f *ConfigFactory) InitConsulClientWithParamsOrEnv(serverAddr, namespaceId, group, username, password string) error {
	// 优先使用环境变量
	envServerAddr := os.Getenv("CONSUL_SERVER_ADDR")
	envNamespaceId := os.Getenv("CONSUL_NAMESPACE_ID")
	envGroup := os.Getenv("CONSUL_GROUP")
	envUsername := os.Getenv("CONSUL_USERNAME")
	envPassword := os.Getenv("CONSUL_PASSWORD")

	// 环境变量优先，为空则使用传入参数
	if envServerAddr != "" {
		serverAddr = envServerAddr
	}
	if envNamespaceId != "" {
		namespaceId = envNamespaceId
	}
	if envGroup != "" {
		group = envGroup
	}
	if envUsername != "" {
		username = envUsername
	}
	if envPassword != "" {
		password = envPassword
	}

	if serverAddr == "" || namespaceId == "" || group == "" {
		return fmt.Errorf("缺少必要的配置: serverAddr/namespaceId/group")
	}

	client, err := NewConsulConfigClient(serverAddr, namespaceId, group, username, password)
	if err != nil {
		return fmt.Errorf("初始化 Consul 客户端失败: %w", err)
	}
	f.consulClient = client
	f.configType = ConfigTypeConsul
	return nil
}

// GetCommonConfig 获取通用配置（兼容接口）
func (f *ConfigFactory) GetCommonConfig(group string) (*CommonConfig, error) {
	switch f.configType {
	case ConfigTypeNacos:
		if f.nacosClient == nil {
			return nil, fmt.Errorf("nacos 客户端未初始化")
		}
		return f.nacosClient.GetCommonConfig(group)
	case ConfigTypeConsul:
		if f.consulClient == nil {
			return nil, fmt.Errorf("consul 客户端未初始化")
		}
		return f.consulClient.GetCommonConfig(group)
	default:
		return nil, fmt.Errorf("不支持的配置类型: %s", f.configType)
	}
}

// GetKvConfig 获取键值配置（兼容接口）
func (f *ConfigFactory) GetKvConfig(dataId, group string) (string, error) {
	switch f.configType {
	case ConfigTypeNacos:
		if f.nacosClient == nil {
			return "", fmt.Errorf("nacos 客户端未初始化")
		}
		return f.nacosClient.GetConfig(dataId, group)
	case ConfigTypeConsul:
		if f.consulClient == nil {
			return "", fmt.Errorf("consul 客户端未初始化")
		}
		return f.consulClient.GetConfig(dataId, group)
	default:
		return "", fmt.Errorf("不支持的配置类型: %s", f.configType)
	}
}

// GetPasetoPubConfig 获取 Paseto 公钥配置（兼容接口）
func (f *ConfigFactory) GetPasetoPubConfig(group string) (*model.PasetoConfig, error) {
	switch f.configType {
	case ConfigTypeNacos:
		if f.nacosClient == nil {
			return nil, fmt.Errorf("nacos 客户端未初始化")
		}
		return f.nacosClient.GetPasetoPubConfig(group)
	case ConfigTypeConsul:
		if f.consulClient == nil {
			return nil, fmt.Errorf("consul 客户端未初始化")
		}
		return f.consulClient.GetPasetoPubConfig(group)
	default:
		return nil, fmt.Errorf("不支持的配置类型: %s", f.configType)
	}
}

// GetPasetoSecretConfig 获取 Paseto 密钥配置（兼容接口）
func (f *ConfigFactory) GetPasetoSecretConfig(group string) (*model.PasetoConfig, error) {
	switch f.configType {
	case ConfigTypeNacos:
		if f.nacosClient == nil {
			return nil, fmt.Errorf("nacos 客户端未初始化")
		}
		return f.nacosClient.GetPasetoSecretConfig(group)
	case ConfigTypeConsul:
		if f.consulClient == nil {
			return nil, fmt.Errorf("consul 客户端未初始化")
		}
		return f.consulClient.GetPasetoSecretConfig(group)
	default:
		return nil, fmt.Errorf("不支持的配置类型: %s", f.configType)
	}
}

// GetNacosClient 获取 Nacos 客户端（用于高级操作）
func (f *ConfigFactory) GetNacosClient() *NacosConfigClient {
	return f.nacosClient
}

// GetConsulClient 获取 Consul 客户端（用于高级操作）
func (f *ConfigFactory) GetConsulClient() *ConsulConfigClient {
	return f.consulClient
}

// Close 关闭配置工厂
func (f *ConfigFactory) Close() error {
	if f.nacosClient != nil {
		if err := f.nacosClient.Close(); err != nil {
			return err
		}
	}
	if f.consulClient != nil {
		if err := f.consulClient.Close(); err != nil {
			return err
		}
	}
	return nil
}

// 全局配置工厂实例
var globalConfigFactory *ConfigFactory

// InitGlobalConfigFactory 初始化全局配置工厂
func InitGlobalConfigFactory(options *ConfigFactoryOptions) error {
	globalConfigFactory = NewConfigFactory(options)
	globalConfigFactory.SetConfigType(options.ConfigType)

	if options.ConfigType == ConfigTypeNacos {
		return globalConfigFactory.InitNacosClientWithParamsOrEnv(options.ServerAddr, options.NamespaceId, options.Group, options.Username, options.Password)
	} else if options.ConfigType == ConfigTypeConsul {
		return globalConfigFactory.InitConsulClientWithParamsOrEnv(options.ServerAddr, options.NamespaceId, options.Group, options.Username, options.Password)
	}

	return nil
}

// InitGlobalConfigFactoryWithNacos 使用指定参数，环境变量优先
func InitGlobalConfigFactoryWithNacos(options *ConfigFactoryOptions) error {
	globalConfigFactory = NewConfigFactory(options)
	globalConfigFactory.SetConfigType(ConfigTypeNacos)
	return globalConfigFactory.InitNacosClientWithParamsOrEnv(options.ServerAddr, options.NamespaceId, options.Group, options.Username, options.Password)
}

// InitGlobalConfigFactoryWithConsul 使用指定参数，环境变量优先
func InitGlobalConfigFactoryWithConsul(options *ConfigFactoryOptions) error {
	globalConfigFactory = NewConfigFactory(options)
	globalConfigFactory.SetConfigType(ConfigTypeConsul)
	return globalConfigFactory.InitConsulClientWithParamsOrEnv(options.ServerAddr, options.NamespaceId, options.Group, options.Username, options.Password)
}

// GetGlobalConfigFactory 获取全局配置工厂
func GetGlobalConfigFactory() *ConfigFactory {
	if globalConfigFactory == nil {
		// 默认使用 Consul
		globalConfigFactory = NewConfigFactory(&ConfigFactoryOptions{
			ConfigType: ConfigTypeConsul,
		})
	}
	return globalConfigFactory
}

// 全局便捷函数
func GetCommonConfigGlobal(group string) (*CommonConfig, error) {
	return GetGlobalConfigFactory().GetCommonConfig(group)
}

func GetKvConfigGlobal(keyName, group string) (string, error) {
	factory := GetGlobalConfigFactory()
	return factory.GetKvConfig(keyName, group)
}

func GetPasetoPubConfigGlobal(group string) (*model.PasetoConfig, error) {
	return GetGlobalConfigFactory().GetPasetoPubConfig(group)
}

func GetPasetoSecretConfigGlobal(group string) (*model.PasetoConfig, error) {
	return GetGlobalConfigFactory().GetPasetoSecretConfig(group)
}
