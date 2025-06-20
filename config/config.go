package config

import (
	"os"
	"time"

	"github.com/spf13/viper"
	"golang.org/x/xerrors"
)

type AppConfig struct {
	Server             Server      `yaml:"server"`
	DCE                DCE         `yaml:"dce"`
	Tracer             Tracer      `yaml:"tracer"`
	Kube               Kube        `yaml:"kube"`
	MySQL              DBConfig    `yaml:"mysql"`
	Redis              RedisConfig `yaml:"redis"`
	InsecureSkipVerify bool        `yaml:"insecureSkipVerify"`
	Level              string      `yaml:"level"`
	HostCluster        string      `yaml:"hostCluster"`
	HostNamespace      string      `yaml:"hostNamespace"`
	EnvConfs           []*EnvConf  `yaml:"envConfs"` // 环境配置,支持多个环境
}

// MySQL DB 配置
type DBConfig struct {
	DBType               string        `yaml:"dbType"`               // 数据库类型，默认 mysql
	DSN                  string        `yaml:"dsn"`                  // data source name, e.g. root:123456@tcp(127.0.0.1:3306)/hydra
	MaxIdleConns         int           `yaml:"maxIdleConns"`         // 最大空闲连接数
	MaxOpenConns         int           `yaml:"maxOpenConns"`         // 最大连接数
	AutoMigrate          bool          `yaml:"autoMigrate"`          // 自动建表，补全缺失字段，初始化数据
	Debug                bool          `yaml:"debug"`                // 是否开启调试模式
	CacheFlag            bool          `yaml:"cacheFlag"`            // 是否开启查询缓存
	CacheExpiration      time.Duration `yaml:"cacheExpiration"`      // 缓存过期时间
	CacheCleanupInterval time.Duration `yaml:"cacheCleanupInterval"` // 缓存清理时间间隔
}

// Redis 配置
type RedisConfig struct {
	URL string `yaml:"port"`
}

// Server Port 配置
type Server struct {
	Port uint16 `yaml:"port"`
}

// DCE 配置
type DCE struct {
	URL     string `yaml:"url"`
	Insight string `yaml:"insight"`
	Token   string `yaml:"token"`
}

// Tracer 配置
type Tracer struct {
	Enable   bool   `yaml:"enable"`
	Endpoint string `yaml:"endpoint"`
}

// Kube 配置
type Kube struct {
	ConfigPath string `yaml:"configPath"`
}

type EnvConf struct {
	Name      string `yaml:"name"`      // 环境名称，唯一，如：test，prod, 注意，该值影响当前部署环境Token的认证，如，当前环境为test，则在Token的认证时只会查询EnvName为test的值，请谨慎修改
	Alias     string `yaml:"alias"`     // 环境别名，如：测试环境、生产环境，用于展示
	IsDefault bool   `yaml:"isDefault"` // 是否默认环境，默认环境表示当前环境，默认环境只能有一个
}

var (
	CurrentEnvName string
	CurrentAlias   string
	EnvConfs       []*EnvConf
)

func GetEnvMap() map[string]string {
	envMap := make(map[string]string)
	for _, envConf := range EnvConfs {
		envMap[envConf.Name] = envConf.Alias
	}
	return envMap
}

func InitConfig() (*AppConfig, error) {
	// 支持配置文件中key的可以包含.符号
	newViper := viper.NewWithOptions(viper.KeyDelimiter("::"))
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		path, err := os.Getwd()
		if err != nil {
			return nil, xerrors.Errorf(": %w", err)
		}
		configPath = path + "/config"
	}
	newViper.AddConfigPath(configPath)
	var appConfig AppConfig
	appConfig.Server.Port = 8888
	appConfig.Level = "debug"
	appConfig.HostCluster = "kpanda-global-cluster"
	appConfig.HostNamespace = "auth-engine-system"
	if clusterFromEnv := os.Getenv("HOST_CLUSTER"); len(clusterFromEnv) != 0 {
		appConfig.HostCluster = clusterFromEnv
	}
	if namespaceFromEnv := os.Getenv("HOST_NAMESPACE"); len(namespaceFromEnv) != 0 {
		appConfig.HostNamespace = namespaceFromEnv
	}
	newViper.SetConfigName("config")
	newViper.SetConfigType("yaml")
	if err := newViper.ReadInConfig(); err != nil {
		return nil, xerrors.Errorf(": %w", err)
	}
	if err := newViper.Unmarshal(&appConfig); err != nil {
		return nil, xerrors.Errorf(": %w", err)
	}
	EnvConfs = appConfig.EnvConfs
	for _, envConf := range appConfig.EnvConfs {
		if envConf.IsDefault {
			CurrentEnvName = envConf.Name
			CurrentAlias = envConf.Alias
			break
		}
	}
	return &appConfig, nil
}
