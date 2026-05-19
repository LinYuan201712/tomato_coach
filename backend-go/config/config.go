package config

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config 应用全局配置 (Updated)
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Log      LogConfig      `mapstructure:"log"`
	Upload   UploadConfig   `mapstructure:"upload"`
	ARK      ARKConfig      `mapstructure:"ark"`
	Aliyun   AliyunConfig   `mapstructure:"aliyun"`
	RAG      RAGConfig      `mapstructure:"rag"`
	Elastic  ElasticConfig  `mapstructure:"elasticsearch"`
	Channels ChannelsConfig `mapstructure:"channels"`
	Langfuse LangfuseConfig `mapstructure:"langfuse"`
	MinerU   MinerUConfig   `mapstructure:"mineru"`
}

// LangfuseConfig Langfuse 可观测性配置
type LangfuseConfig struct {
	Enabled   bool   `mapstructure:"enabled"`
	SecretKey string `mapstructure:"secret_key"`
	PublicKey string `mapstructure:"public_key"`
	BaseURL   string `mapstructure:"base_url"`
}

// MinerUConfig MinerU PDF 解析配置
type MinerUConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	Token       string `mapstructure:"token"`
	PollTimeout string `mapstructure:"poll_timeout"`
}

// ChannelsConfig 渠道配置
type ChannelsConfig struct {
	Feishu FeishuChannelConfig `mapstructure:"feishu"`
}

// FeishuChannelConfig 飞书渠道配置
type FeishuChannelConfig struct {
	Enabled           bool     `mapstructure:"enabled"`
	AppID             string   `mapstructure:"app_id"`
	AppSecret         string   `mapstructure:"app_secret"`
	VerificationToken string   `mapstructure:"verification_token"`
	EncryptKey        string   `mapstructure:"encrypt_key"`
	Domain            string   `mapstructure:"domain"`
	AllowedIDs        []string `mapstructure:"allowed_ids"`
	DMPolicy          string   `mapstructure:"dm_policy"`
	CronOutputChatID string   `mapstructure:"cron_output_chat_id"`
}

// AliyunConfig 阿里云 DashScope 配置
type AliyunConfig struct {
	APIKey         string `mapstructure:"api_key"`
	ChatModel      string `mapstructure:"chat_model"`
	CheapModel     string `mapstructure:"cheap_model"`
	EmbeddingModel string `mapstructure:"embedding_model"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port        int    `mapstructure:"port"`
	ContextPath string `mapstructure:"context_path"`
	Environment string `mapstructure:"environment"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	Database        string `mapstructure:"database"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
	LogLevel        string `mapstructure:"log_level"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret     string `mapstructure:"secret"`
	Expiration int64  `mapstructure:"expiration"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	OutputPath string `mapstructure:"output_path"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
}

// UploadConfig 上传文件配置
type UploadConfig struct {
	BasePath string `mapstructure:"base_path"`
	MaxSize  int64  `mapstructure:"max_size"`
}

// ARKConfig 火山引擎 ARK 模型配置
type ARKConfig struct {
	APIKey         string `mapstructure:"api_key"`
	ChatModel      string `mapstructure:"chat_model"`
	CheapModel     string `mapstructure:"cheap_model"`
	EmbeddingModel string `mapstructure:"embedding_model"`
	BaseURL        string `mapstructure:"base_url"`
}

type RAGConfig struct {
	TopK           int     `mapstructure:"top_k"`
	ScoreThreshold float64 `mapstructure:"score_threshold"`
	HybridWeight   float64 `mapstructure:"hybrid_weight"` // 向量搜索权重
}

// ElasticConfig Elasticsearch 配置
type ElasticConfig struct {
	Addresses []string `mapstructure:"addresses"`
	Username  string   `mapstructure:"username"`
	Password  string   `mapstructure:"password"`
	Index     string   `mapstructure:"index"`
}

// LoadConfig 从配置文件加载配置
func LoadConfig(configPath string) (*Config, error) {
	// 加载 .env 文件（如果存在）
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	v := viper.New()

	// 设置配置文件参数
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(configPath)

	// 读取环境变量
	v.AutomaticEnv()

	// 绑定环境变量前缀
	v.SetEnvPrefix("TOMATO")

	// 绑定特定的环境变量
	bindEnvVars(v)

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		log.Printf("Failed to read config file: %v, using default config", err)
	}

	// 反序列化到Config结构体
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 设置默认值
	setDefaults(&cfg)

	return &cfg, nil
}

// bindEnvVars 绑定环境变量
func bindEnvVars(v *viper.Viper) {
	// 数据库配置
	_ = v.BindEnv("database.host", "TOMATO_DB_HOST")
	_ = v.BindEnv("database.port", "TOMATO_DB_PORT")
	_ = v.BindEnv("database.user", "TOMATO_DB_USER")
	_ = v.BindEnv("database.password", "TOMATO_DB_PASSWORD")
	_ = v.BindEnv("database.database", "TOMATO_DB_NAME")

	// JWT配置
	_ = v.BindEnv("jwt.secret", "TOMATO_JWT_SECRET")
	_ = v.BindEnv("jwt.expiration", "TOMATO_JWT_EXPIRATION")

	// 服务器配置
	_ = v.BindEnv("server.port", "TOMATO_SERVER_PORT")
	_ = v.BindEnv("server.environment", "TOMATO_ENVIRONMENT")

	_ = v.BindEnv("ark.embedding_model", "ARK_EMBEDDING_MODEL")
	_ = v.BindEnv("ark.api_key", "ARK_API_KEY")

	_ = v.BindEnv("aliyun.embedding_model", "ALIYUN_EMBEDDING_MODEL")
	_ = v.BindEnv("aliyun.api_key", "ALIYUN_API_KEY")

	// 渠道配置
	_ = v.BindEnv("channels.feishu.app_id", "FEISHU_APP_ID")
	_ = v.BindEnv("channels.feishu.app_secret", "FEISHU_APP_SECRET")

	// Langfuse 配置
	_ = v.BindEnv("langfuse.secret_key", "LANGFUSE_SECRET_KEY")
	_ = v.BindEnv("langfuse.public_key", "LANGFUSE_PUBLIC_KEY")
	_ = v.BindEnv("langfuse.base_url", "LANGFUSE_BASE_URL")

	// MinerU 配置
	_ = v.BindEnv("mineru.enabled", "MINERU_ENABLED")
	_ = v.BindEnv("mineru.token", "MINERU_TOKEN")
}

// setDefaults 设置配置默认值
func setDefaults(cfg *Config) {
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8090
	}
	if cfg.Server.ContextPath == "" {
		cfg.Server.ContextPath = "/api"
	}
	if cfg.Server.Environment == "" {
		cfg.Server.Environment = "dev"
	}

	if cfg.Database.Host == "" {
		cfg.Database.Host = "localhost"
	}
	if cfg.Database.Port == 0 {
		cfg.Database.Port = 3306
	}
	if cfg.Database.Database == "" {
		cfg.Database.Database = "tomato_study_room"
	}
	if cfg.Database.MaxOpenConns == 0 {
		cfg.Database.MaxOpenConns = 25
	}
	if cfg.Database.MaxIdleConns == 0 {
		cfg.Database.MaxIdleConns = 5
	}
	if cfg.Database.ConnMaxLifetime == 0 {
		cfg.Database.ConnMaxLifetime = 5 * 60 // 5分钟
	}

	if cfg.JWT.Secret == "" {
		cfg.JWT.Secret = "tomato_secret_key_change_in_production"
	}
	if cfg.JWT.Expiration == 0 {
		cfg.JWT.Expiration = 3600 // 1小时
	}

	if cfg.Log.Level == "" {
		cfg.Log.Level = "info"
	}
	if cfg.Log.Format == "" {
		cfg.Log.Format = "json"
	}
	if cfg.Log.OutputPath == "" {
		cfg.Log.OutputPath = "stdout"
	}
	if cfg.Log.MaxSize == 0 {
		cfg.Log.MaxSize = 100 // MB
	}
	if cfg.Log.MaxBackups == 0 {
		cfg.Log.MaxBackups = 3
	}
	if cfg.Log.MaxAge == 0 {
		cfg.Log.MaxAge = 7 // 天
	}

	if cfg.Upload.BasePath == "" {
		cfg.Upload.BasePath = "./uploads"
	}
	if cfg.Upload.MaxSize == 0 {
		cfg.Upload.MaxSize = 10 * 1024 * 1024 // 10MB
	}

	// ARK 默认值
	if cfg.ARK.ChatModel == "" {
		cfg.ARK.ChatModel = "doubao-pro-32k"
	}
	if cfg.ARK.CheapModel == "" {
		cfg.ARK.CheapModel = "doubao-lite-4k"
	}
	if cfg.ARK.EmbeddingModel == "" {
		cfg.ARK.EmbeddingModel = "doubao-embedding"
	}
	if cfg.ARK.BaseURL == "" {
		cfg.ARK.BaseURL = "https://ark.cn-beijing.volces.com/api/v3"
	}

	// Aliyun 默认值
	if cfg.Aliyun.ChatModel == "" {
		cfg.Aliyun.ChatModel = "qwen-plus"
	}
	if cfg.Aliyun.CheapModel == "" {
		cfg.Aliyun.CheapModel = "qwen-turbo"
	}
	if cfg.Aliyun.EmbeddingModel == "" {
		cfg.Aliyun.EmbeddingModel = "text-embedding-v3"
	}

	// RAG 默认值
	if cfg.RAG.TopK == 0 {
		cfg.RAG.TopK = 5
	}
	if cfg.RAG.ScoreThreshold == 0 {
		cfg.RAG.ScoreThreshold = 0.5
	}
	if cfg.RAG.HybridWeight == 0 {
		cfg.RAG.HybridWeight = 0.7
	}

	// Elasticsearch 默认值
	if len(cfg.Elastic.Addresses) == 0 {
		cfg.Elastic.Addresses = []string{"http://localhost:9200"}
	}
	if cfg.Elastic.Index == "" {
		cfg.Elastic.Index = "tomato_rag"
	}

	// 渠道默认值
	if cfg.Channels.Feishu.Domain == "" {
		cfg.Channels.Feishu.Domain = "feishu"
	}

	// Langfuse 默认值
	if cfg.Langfuse.BaseURL == "" {
		cfg.Langfuse.BaseURL = "https://cloud.langfuse.com"
	}

	// MinerU 默认值
	if cfg.MinerU.PollTimeout == "" {
		cfg.MinerU.PollTimeout = "15m"
	}
}
