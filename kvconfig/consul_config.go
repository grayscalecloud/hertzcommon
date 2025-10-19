package kvconfig

import (
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/grayscalecloud/hertzcommon/hdmodel"
	"github.com/hashicorp/consul/api"
	"gopkg.in/yaml.v2"
)

type CommonConfig struct {
	Env   string
	Kitex hdmodel.Kitex `yaml:"kitex"`
	MySQL hdmodel.MySQL `yaml:"mysql"`
	Redis hdmodel.Redis `yaml:"redis"`
	OTel  hdmodel.OTel  `yaml:"otel"`
}

// ConsulConfigClient Consul 配置客户端
type ConsulConfigClient struct {
	client      *api.Client
	serverAddr  string
	namespaceId string
	group       string
	username    string
	password    string
	watchChans  map[string]chan struct{} // 用于停止监听的通道
}

// NewConsulConfigClient 创建 Consul 配置客户端
func NewConsulConfigClient(serverAddr, namespaceId, group, username, password string) (*ConsulConfigClient, error) {
	// 创建 Consul 客户端
	client, err := api.NewClient(&api.Config{
		Address: serverAddr,
		// 如果需要认证，可以在这里添加
		// Token: token,
	})
	if err != nil {
		return nil, fmt.Errorf("创建 Consul 客户端失败: %w", err)
	}

	return &ConsulConfigClient{
		client:      client,
		serverAddr:  serverAddr,
		namespaceId: namespaceId,
		group:       group,
		username:    username,
		password:    password,
		watchChans:  make(map[string]chan struct{}),
	}, nil
}

// GetConfig 获取配置
func (c *ConsulConfigClient) GetConfig(dataId, group string) (string, error) {
	key := c.buildKey(dataId, group)

	content, _, err := c.client.KV().Get(key, nil)
	if err != nil {
		return "", fmt.Errorf("获取配置失败: %w", err)
	}

	if content == nil {
		return "", fmt.Errorf("配置不存在: %s", key)
	}

	return string(content.Value), nil
}

// PublishConfig 发布配置
func (c *ConsulConfigClient) PublishConfig(dataId, group, content string) error {
	key := c.buildKey(dataId, group)

	_, err := c.client.KV().Put(&api.KVPair{
		Key:   key,
		Value: []byte(content),
	}, nil)

	if err != nil {
		return fmt.Errorf("发布配置失败: %w", err)
	}

	return nil
}

// DeleteConfig 删除配置
func (c *ConsulConfigClient) DeleteConfig(dataId, group string) error {
	key := c.buildKey(dataId, group)

	_, err := c.client.KV().Delete(key, nil)
	if err != nil {
		return fmt.Errorf("删除配置失败: %w", err)
	}

	return nil
}

// ListenConfig 监听配置变化（Consul 使用 blocking query）
func (c *ConsulConfigClient) ListenConfig(dataId, group string, callback func(content string)) error {
	key := c.buildKey(dataId, group)

	// 检查是否已经在监听
	if _, exists := c.watchChans[key]; exists {
		return fmt.Errorf("配置已在监听中: %s", key)
	}

	// 创建停止通道
	stopChan := make(chan struct{})
	c.watchChans[key] = stopChan

	// 启动监听 goroutine
	go c.watchKey(key, stopChan, callback)

	hlog.Infof("开始监听 Consul 配置: %s", key)
	return nil
}

// WatchConfig 监听配置变化，返回配置变化通道
func (c *ConsulConfigClient) WatchConfig(dataId, group string) (<-chan string, error) {
	key := c.buildKey(dataId, group)

	// 检查是否已经在监听
	if _, exists := c.watchChans[key]; exists {
		return nil, fmt.Errorf("配置已在监听中: %s", key)
	}

	// 创建配置变化通道
	configChan := make(chan string, 10) // 带缓冲的通道
	stopChan := make(chan struct{})
	c.watchChans[key] = stopChan

	// 启动监听 goroutine
	go c.watchKeyWithChan(key, stopChan, configChan)

	hlog.Infof("开始监听 Consul 配置: %s", key)
	return configChan, nil
}

// watchKey 监听指定 key 的变化
func (c *ConsulConfigClient) watchKey(key string, stopChan chan struct{}, callback func(content string)) {
	defer func() {
		// 清理停止通道
		delete(c.watchChans, key)
		hlog.Infof("停止监听 Consul 配置: %s", key)
	}()

	var lastIndex uint64 = 0

	for {
		select {
		case <-stopChan:
			return
		default:
			// 使用 blocking query 监听 key 变化
			queryOptions := &api.QueryOptions{
				WaitIndex: lastIndex,
				WaitTime:  30 * time.Second, // 30秒超时
			}

			kvPair, meta, err := c.client.KV().Get(key, queryOptions)
			if err != nil {
				hlog.Errorf("监听配置失败 [%s]: %v", key, err)
				time.Sleep(5 * time.Second) // 出错后等待5秒再重试
				continue
			}

			// 更新 lastIndex
			lastIndex = meta.LastIndex

			// 检查是否有变化
			if kvPair != nil {
				// 配置存在，调用回调函数
				callback(string(kvPair.Value))
			} else {
				// 配置被删除
				callback("")
			}
		}
	}
}

// watchKeyWithChan 监听指定 key 的变化，通过 channel 发送配置
func (c *ConsulConfigClient) watchKeyWithChan(key string, stopChan chan struct{}, configChan chan string) {
	defer func() {
		// 清理停止通道
		delete(c.watchChans, key)
		close(configChan) // 关闭配置通道
		hlog.Infof("停止监听 Consul 配置: %s", key)
	}()

	var lastIndex uint64 = 0

	for {
		select {
		case <-stopChan:
			return
		default:
			// 使用 blocking query 监听 key 变化
			queryOptions := &api.QueryOptions{
				WaitIndex: lastIndex,
				WaitTime:  30 * time.Second, // 30秒超时
			}

			kvPair, meta, err := c.client.KV().Get(key, queryOptions)
			if err != nil {
				hlog.Errorf("监听配置失败 [%s]: %v", key, err)
				time.Sleep(5 * time.Second) // 出错后等待5秒再重试
				continue
			}

			// 更新 lastIndex
			lastIndex = meta.LastIndex

			// 检查是否有变化
			if kvPair != nil {
				// 配置存在，发送到通道
				select {
				case configChan <- string(kvPair.Value):
				case <-stopChan:
					return
				}
			} else {
				// 配置被删除，发送空字符串
				select {
				case configChan <- "":
				case <-stopChan:
					return
				}
			}
		}
	}
}

// GetCommonConfig 获取通用配置
func (c *ConsulConfigClient) GetCommonConfig(group string) (*CommonConfig, error) {
	content, err := c.GetConfig("common", group)
	if err != nil {
		return nil, err
	}

	conf := new(CommonConfig)
	err = yaml.Unmarshal([]byte(content), &conf)
	if err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	return conf, nil
}

// GetPasetoPubConfig 获取 Paseto 公钥配置
func (c *ConsulConfigClient) GetPasetoPubConfig(group string) (*hdmodel.PasetoConfig, error) {
	content, err := c.GetConfig("pasetopub", group)
	if err != nil {
		return nil, err
	}

	conf := new(hdmodel.PasetoConfig)
	err = yaml.Unmarshal([]byte(content), &conf)
	if err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	return conf, nil
}

// GetPasetoSecretConfig 获取 Paseto 密钥配置
func (c *ConsulConfigClient) GetPasetoSecretConfig(group string) (*hdmodel.PasetoConfig, error) {
	content, err := c.GetConfig("pasetosecret", group)
	if err != nil {
		return nil, err
	}

	conf := new(hdmodel.PasetoConfig)
	err = yaml.Unmarshal([]byte(content), &conf)
	if err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	return conf, nil
}

// buildKey 构建 Consul KV 的 key
func (c *ConsulConfigClient) buildKey(dataId, group string) string {
	// 使用 namespaceId 和 group 构建 key
	// 格式: namespaceId/group/dataId
	parts := []string{c.namespaceId, group, dataId}
	return strings.Join(parts, "/")
}

// StopListenConfig 停止监听指定配置
func (c *ConsulConfigClient) StopListenConfig(dataId, group string) error {
	key := c.buildKey(dataId, group)

	if stopChan, exists := c.watchChans[key]; exists {
		close(stopChan)
		return nil
	}

	return fmt.Errorf("配置监听不存在: %s", key)
}

// StopAllListenConfigs 停止所有配置监听
func (c *ConsulConfigClient) StopAllListenConfigs() error {
	for key, stopChan := range c.watchChans {
		close(stopChan)
		hlog.Infof("停止监听 Consul 配置: %s", key)
	}

	// 清空监听通道映射
	c.watchChans = make(map[string]chan struct{})
	return nil
}

// Close 关闭客户端
func (c *ConsulConfigClient) Close() error {
	// 停止所有监听
	c.StopAllListenConfigs()

	// Consul 客户端不需要显式关闭
	return nil
}

func GetCommonConfig(registryAddr string) (*CommonConfig, error) {
	client, err := api.NewClient(&api.Config{Address: registryAddr})
	if err != nil {
		fmt.Println("Error creating Consul client:", err)
		return nil, err
	}
	//获取配置
	content, _, err := client.KV().Get("onebids/common", nil)
	if err != nil {
		fmt.Println("Error getting config:", err)
		return nil, err
	}
	conf := new(CommonConfig)
	err = yaml.Unmarshal(content.Value, &conf)
	if err != nil {
		hlog.Error("parse yaml error - %v", err)
		panic(err)
	}
	return conf, nil
}

func GetKvConfig[T any](registryAddr string, keyName string) (*T, error) {
	client, err := api.NewClient(&api.Config{Address: registryAddr})
	if err != nil {
		fmt.Println("Error creating Consul client:", err)
		return nil, err
	}
	//获取配置
	content, _, err := client.KV().Get(keyName, nil)
	if err != nil {
		fmt.Println("Error getting config:", err)
		return nil, err
	}
	conf := new(T)
	err = yaml.Unmarshal(content.Value, &conf)
	if err != nil {
		hlog.Error("parse yaml error - %v", err)
		panic(err)
	}
	return conf, nil
}

func GetPasetoPubConfig(registryAddr string) (*hdmodel.PasetoConfig, error) {
	client, err := api.NewClient(&api.Config{Address: registryAddr})
	if err != nil {
		fmt.Println("Error creating Consul client:", err)
		return nil, err
	}
	//获取配置
	content, _, err := client.KV().Get("onebids/pasetopub", nil)
	if err != nil {
		fmt.Println("Error getting config:", err)
		return nil, err
	}
	conf := new(hdmodel.PasetoConfig)
	err = yaml.Unmarshal(content.Value, &conf)
	if err != nil {
		hlog.Error("parse yaml error - %v", err)
		panic(err)
	}

	return conf, nil
}
func GetPasetoSecretConfig(registryAddr string) (*hdmodel.PasetoConfig, error) {
	client, err := api.NewClient(&api.Config{Address: registryAddr})
	if err != nil {
		fmt.Println("Error creating Consul client:", err)
		return nil, err
	}
	//获取配置
	content, _, err := client.KV().Get("onebids/pasetosecret", nil)
	if err != nil {
		fmt.Println("Error getting config:", err)
		return nil, err
	}
	conf := new(hdmodel.PasetoConfig)
	err = yaml.Unmarshal(content.Value, &conf)
	if err != nil {
		hlog.Error("parse yaml error - %v", err)
		panic(err)
	}

	return conf, nil
}
