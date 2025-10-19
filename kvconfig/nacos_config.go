package kvconfig

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/grayscalecloud/hertzcommon/hdmodel"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"gopkg.in/yaml.v2"
)

// NacosConfig Nacos 配置中心配置
type NacosConfig struct {
	ServerConfigs []constant.ServerConfig `yaml:"server_configs"`
	ClientConfig  constant.ClientConfig   `yaml:"client_config"`
}

// NacosConfigClient Nacos 配置客户端
type NacosConfigClient struct {
	client config_client.IConfigClient
	config *NacosConfig
}

// NewNacosConfigClient 创建 Nacos 配置客户端
func NewNacosConfigClient(serverAddrs []string, namespaceId, group string, username, password string) (*NacosConfigClient, error) {
	// 默认配置
	clientConfig := constant.ClientConfig{
		NamespaceId:         namespaceId,
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "/tmp/nacos/log",
		CacheDir:            "/tmp/nacos/cache",
		LogLevel:            "info",
		// 添加身份验证支持
		Username: username,
		Password: password,
	}

	// 服务器配置
	serverConfigs := make([]constant.ServerConfig, 0, len(serverAddrs))
	for _, addr := range serverAddrs {
		// 解析地址，支持 IP:Port 格式
		var ipAddr string
		var port uint64 = 8848

		if len(addr) > 0 {
			// 简单的地址解析，假设格式为 IP:Port
			if len(addr) > 5 && addr[len(addr)-5:] == ":8848" {
				ipAddr = addr[:len(addr)-5]
			} else {
				ipAddr = addr
			}
		}

		serverConfigs = append(serverConfigs, constant.ServerConfig{
			IpAddr: ipAddr,
			Port:   port,
		})
	}

	// 创建配置客户端
	configClient, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("创建 Nacos 配置客户端失败: %w", err)
	}

	return &NacosConfigClient{
		client: configClient,
		config: &NacosConfig{
			ServerConfigs: serverConfigs,
			ClientConfig:  clientConfig,
		},
	}, nil
}

// NewNacosConfigClientWithConfig 使用自定义配置创建 Nacos 配置客户端
func NewNacosConfigClientWithConfig(config *NacosConfig) (*NacosConfigClient, error) {
	configClient, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &config.ClientConfig,
			ServerConfigs: config.ServerConfigs,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("创建 Nacos 配置客户端失败: %w", err)
	}

	return &NacosConfigClient{
		client: configClient,
		config: config,
	}, nil
}

// GetConfig 获取配置
func (c *NacosConfigClient) GetConfig(dataId, group string) (string, error) {
	content, err := c.client.GetConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  group,
	})
	if err != nil {
		return "", fmt.Errorf("获取配置失败 [dataId: %s, group: %s]: %w", dataId, group, err)
	}
	return content, nil
}

// GetCommonConfig 获取通用配置
func (c *NacosConfigClient) GetCommonConfig(group string) (*CommonConfig, error) {
	content, err := c.GetConfig("common", group)
	if err != nil {
		return nil, err
	}

	conf := new(CommonConfig)
	err = yaml.Unmarshal([]byte(content), &conf)
	if err != nil {
		hlog.Error("解析 YAML 配置失败: %v", err)
		return nil, fmt.Errorf("解析 YAML 配置失败: %w", err)
	}

	return conf, nil
}

// GetKvConfigRaw 获取原始配置内容
func (c *NacosConfigClient) GetKvConfigRaw(dataId, group string) (string, error) {
	return c.GetConfig(dataId, group)
}

// GetPasetoPubConfig 获取 Paseto 公钥配置
func (c *NacosConfigClient) GetPasetoPubConfig(group string) (*hdmodel.PasetoConfig, error) {
	content, err := c.GetConfig("pasetopub", group)
	if err != nil {
		return nil, err
	}

	conf := new(hdmodel.PasetoConfig)
	err = yaml.Unmarshal([]byte(content), &conf)
	if err != nil {
		hlog.Error("解析 Paseto 公钥配置失败: %v", err)
		return nil, fmt.Errorf("解析 Paseto 公钥配置失败: %w", err)
	}

	return conf, nil
}

// GetPasetoSecretConfig 获取 Paseto 密钥配置
func (c *NacosConfigClient) GetPasetoSecretConfig(group string) (*hdmodel.PasetoConfig, error) {
	content, err := c.GetConfig("pasetosecret", group)
	if err != nil {
		return nil, err
	}

	conf := new(hdmodel.PasetoConfig)
	err = yaml.Unmarshal([]byte(content), &conf)
	if err != nil {
		hlog.Error("解析 Paseto 密钥配置失败: %v", err)
		return nil, fmt.Errorf("解析 Paseto 密钥配置失败: %w", err)
	}

	return conf, nil
}

// ListenConfig 监听配置变化
func (c *NacosConfigClient) ListenConfig(dataId, group string, callback func(string)) error {
	err := c.client.ListenConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  group,
		OnChange: func(namespace, group, dataId, data string) {
			hlog.Infof("配置发生变化 [namespace: %s, group: %s, dataId: %s]", namespace, group, dataId)
			callback(data)
		},
	})
	if err != nil {
		return fmt.Errorf("监听配置失败 [dataId: %s, group: %s]: %w", dataId, group, err)
	}
	return nil
}

// PublishConfig 发布配置
func (c *NacosConfigClient) PublishConfig(dataId, group, content string) error {
	success, err := c.client.PublishConfig(vo.ConfigParam{
		DataId:  dataId,
		Group:   group,
		Content: content,
	})
	if err != nil {
		return fmt.Errorf("发布配置失败 [dataId: %s, group: %s]: %w", dataId, group, err)
	}
	if !success {
		return fmt.Errorf("发布配置失败，返回 false [dataId: %s, group: %s]", dataId, group)
	}
	return nil
}

// DeleteConfig 删除配置
func (c *NacosConfigClient) DeleteConfig(dataId, group string) error {
	success, err := c.client.DeleteConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  group,
	})
	if err != nil {
		return fmt.Errorf("删除配置失败 [dataId: %s, group: %s]: %w", dataId, group, err)
	}
	if !success {
		return fmt.Errorf("删除配置失败，返回 false [dataId: %s, group: %s]", dataId, group)
	}
	return nil
}

// Close 关闭客户端
func (c *NacosConfigClient) Close() error {
	// Nacos 客户端没有显式的关闭方法，这里可以做一些清理工作
	hlog.Info("Nacos 配置客户端已关闭")
	return nil
}

// GetConfigWithRetry 带重试的配置获取
func (c *NacosConfigClient) GetConfigWithRetry(dataId, group string, maxRetries int, retryInterval time.Duration) (string, error) {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		content, err := c.GetConfig(dataId, group)
		if err == nil {
			return content, nil
		}
		lastErr = err
		if i < maxRetries-1 {
			hlog.Warnf("获取配置失败，第 %d 次重试 [dataId: %s, group: %s]: %v", i+1, dataId, group, err)
			time.Sleep(retryInterval)
		}
	}
	return "", fmt.Errorf("获取配置失败，已重试 %d 次: %w", maxRetries, lastErr)
}

// BatchGetConfigs 批量获取配置
func (c *NacosConfigClient) BatchGetConfigs(configs []ConfigRequest) (map[string]string, error) {
	results := make(map[string]string)
	for _, req := range configs {
		content, err := c.GetConfig(req.DataId, req.Group)
		if err != nil {
			return nil, fmt.Errorf("批量获取配置失败 [dataId: %s, group: %s]: %w", req.DataId, req.Group, err)
		}
		results[req.Key] = content
	}
	return results, nil
}

// ConfigRequest 配置请求结构
type ConfigRequest struct {
	Key    string
	DataId string
	Group  string
}

// WatchConfigs 批量监听配置变化
func (c *NacosConfigClient) WatchConfigs(configs []ConfigRequest, callback func(string, string)) error {
	for _, req := range configs {
		err := c.ListenConfig(req.DataId, req.Group, func(content string) {
			callback(req.Key, content)
		})
		if err != nil {
			return fmt.Errorf("监听配置失败 [dataId: %s, group: %s]: %w", req.DataId, req.Group, err)
		}
	}
	return nil
}

// GetConfigWithContext 带上下文的配置获取
func (c *NacosConfigClient) GetConfigWithContext(ctx context.Context, dataId, group string) (string, error) {
	// 创建一个带超时的上下文
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// 使用 channel 来接收结果
	type result struct {
		content string
		err     error
	}
	resultChan := make(chan result, 1)

	go func() {
		content, err := c.GetConfig(dataId, group)
		resultChan <- result{content: content, err: err}
	}()

	select {
	case res := <-resultChan:
		return res.content, res.err
	case <-ctx.Done():
		return "", fmt.Errorf("获取配置超时 [dataId: %s, group: %s]: %w", dataId, group, ctx.Err())
	}
}
