package config

import (
	"fmt"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

var (
	once     sync.Once
	instance *Config
)

// Config 全局配置结构体
type Config struct {
	App       AppConfig       `mapstructure:"app"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Redis     RedisConfig     `mapstructure:"redis"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	Logger    LoggerConfig    `mapstructure:"logger"`
	WeChat    WeChatConfig    `mapstructure:"wechat"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
}

type AppConfig struct {
	Name               string   `mapstructure:"name"`
	Version            string   `mapstructure:"version"`
	Mode               string   `mapstructure:"mode"`
	Port               int      `mapstructure:"port"`
	SecretKey          string   `mapstructure:"secret_key"`
	CORSAllowedOrigins []string `mapstructure:"cors_allowed_origins"`
}

type DatabaseConfig struct {
	Driver           string `mapstructure:"driver"`
	Host             string `mapstructure:"host"`
	Port             int    `mapstructure:"port"`
	Username         string `mapstructure:"username"`
	Password         string `mapstructure:"password"`
	DBName           string `mapstructure:"dbname"`
	Charset          string `mapstructure:"charset"`
	MaxOpenConns     int    `mapstructure:"max_open_conns"`
	MaxIdleConns     int    `mapstructure:"max_idle_conns"`
	MaxLifetimeHours int    `mapstructure:"max_lifetime_hours"`
	LogLevel         string `mapstructure:"log_level"`
}

// DSN 生成MySQL数据源名称
func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		d.Username, d.Password, d.Host, d.Port, d.DBName, d.Charset)
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

// Addr 返回Redis连接地址
func (r *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

type JWTConfig struct {
	Secret             string `mapstructure:"secret"`
	ExpireHours        int    `mapstructure:"expire_hours"`
	RefreshExpireHours int    `mapstructure:"refresh_expire_hours"`
}

type LoggerConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	FilePath   string `mapstructure:"file_path"`
	MaxSizeMB  int    `mapstructure:"max_size_mb"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAgeDays int    `mapstructure:"max_age_days"`
	Compress   bool   `mapstructure:"compress"`
}

type WeChatConfig struct {
	MiniProgram     WeChatMiniProgramConfig     `mapstructure:"mini_program"`
	OfficialAccount WeChatOfficialAccountConfig `mapstructure:"official_account"`
}

type WeChatMiniProgramConfig struct {
	AppID     string `mapstructure:"app_id"`
	AppSecret string `mapstructure:"app_secret"`
}

type WeChatOfficialAccountConfig struct {
	AppID     string `mapstructure:"app_id"`
	AppSecret string `mapstructure:"app_secret"`
}

type RateLimitConfig struct {
	Enabled           bool    `mapstructure:"enabled"`
	RequestsPerSecond float64 `mapstructure:"requests_per_second"`
	Burst             int     `mapstructure:"burst"`
}

// Load 加载配置（单例模式）
func Load(configPath string) (*Config, error) {
	var err error
	once.Do(func() {
		instance, err = load(configPath)
	})
	return instance, err
}

// Get 获取全局配置实例
func Get() *Config {
	return instance
}

func load(configPath string) (*Config, error) {
	v := viper.New()

	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("./configs")
		v.AddConfigPath(".")
	}

	// 支持环境变量覆盖，格式：APP_PORT -> app.port
	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return cfg, nil
}
