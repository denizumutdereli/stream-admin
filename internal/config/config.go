package config

import (
	"log"
	"os"

	"github.com/RackSec/srslog"
	"github.com/denizumutdereli/stream-admin/internal/database"
	"github.com/denizumutdereli/stream-admin/internal/prefix"
	"github.com/go-playground/validator"
	"github.com/mattn/go-colorable"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// TODO: DSN -> .env -> I am keeping here for testing purposes. Production should be with .env and keyspace
type DatabaseDSN struct {
	Host   string
	Port   string
	User   string
	Pass   string
	DBName string
}

type PolicyRules struct {
	RolesPoliciesMin     int `mapstructure:"roles_policies_min"`
	RolesPoliciesMax     int `mapstructure:"roles_policies_max"`
	DashboardIdleTimeout int `mapstructure:"dashboard_id_timeout_in_minutes"`
}

type Config struct {
	AppName                         string             `mapstructure:"APP_NAME" validate:"required"`
	GoServicePort                   string             `mapstructure:"GO_SERVICE_PORT" validate:"required"`
	WsServerPort                    string             `mapstructure:"WSSERVER_PORT" validate:"required"`
	GoGrpcPort                      string             `mapstructure:"GO_GRPC_PORT" validate:"required"`
	NodeUUID                        string             `json:"-"`
	EtcdUrl                         string             `mapstructure:"ETCD_URL" validate:"required" json:"-"`
	SysLog                          string             `mapstructure:"SYSLOG" validate:"required" json:"-"`
	Https                           string             `mapstructure:"HTTPS" validate:"required" json:"-"`
	CorsWhitelist                   string             `mapstructure:"CORS_WHITELIST" validate:"required" json:"-"`
	Channels                        []string           `mapstructure:"CHANNELS"`
	AdminActionLogTopics            []string           `mapstructure:"ADMIN_ACTION_LOG_TOPICS"`
	AllowedServices                 []string           `mapstructure:"ALLOWED_SERVICES" validate:"required"`
	PerRequestLimit                 string             `mapstructure:"PER_REQUEST_LIMIT" validate:"required"`
	AllowAllOrigins                 bool               `mapstructure:"ALLOW_ALL_ORIGINS" json:"-"`
	AllowedOrigins                  []string           `mapstructure:"ALLOWED_ORIGINS" validate:"required" json:"-"`
	AllowedRestMethods              []string           `mapstructure:"ALLOWED_REST_METHODS" validate:"required"`
	AllowedRestHeaders              []string           `mapstructure:"ALLOWED_REST_HEADERS" validate:"required"`
	AllowedSpecificIps              []string           `mapstructure:"ALLOWED_SPECIFIC_IPS" validate:"required"`
	AllowedIpRanges                 []string           `mapstructure:"ALLOWED_IP_RANGES" validate:"required"`
	KafkaBrokers                    []string           `mapstructure:"KAFKA_BROKERS" validate:"required" json:"-"`
	KafkaConsumerGroup              string             `mapstructure:"KAFKA_CONSUMER_GROUP" validate:"required" json:"-"`
	KafkaConsumeTopics              []string           `mapstructure:"KAFKA_CONSUME_TOPICS" validate:"required" json:"-"`
	KafkaProduceTopic               string             `mapstructure:"KAFKA_PRODUCE_TOPIC" validate:"required" json:"-"`
	NatsURL                         []string           `mapstructure:"NATS_URL" validate:"required"`
	RedisURL                        string             `mapstructure:"REDIS_URL" json:"-"`
	RedistPort                      int                `mapstructure:"REDIS_PORT" json:"-"`
	RedisSecretKey                  string             `mapstructure:"REDIS_SECRET_KEY" json:"-"`
	AssetsApi                       string             `mapstructure:"ASSETS_API"`
	MarkupApi                       string             `mapstructure:"MARKUP_API"`
	OTPServiceApi                   string             `mapstructure:"OTP_SERVICE_API"`
	OTPServiceKey                   string             `mapstructure:"OTP_SERVICE_KEY" json:"-"`
	OTPCodesInMinutes               int                `mapstructure:"OTP_CODES_IN_MINUTES"`
	OTPCodesMaxTry                  int                `mapstructure:"OTP_CODES_MAX_TRY"`
	OtpCodesMaxTryInARowInMinutes   int                `mapstructure:"OTP_CODES_MAX_TRY_IN_A_ROW_IN_MINUTES"`
	DefaultPanelLockPeriodInMinutes int                `mapstructure:"DEFAULT_PANEL_LOCK_PERIOD_IN_MINUTES"`
	DefaultCacheQueryTimeInSeconds  int                `mapstructure:"DEFAULT_CACHE_QUERY_TIME_IN_SECONDS"`
	DefaultFuncsTimeOutInSeconds    int                `mapstructure:"DEFAULT_FUNCS_TIMEOUT_IN_SECONDS"`
	DefaultPanelIdleSessionTimeOut  int                `mapstructure:"DEFAULT_PANEL_IDLE_SESSION_TIMEOUT_IN_MINUTES"`
	DefaultPanelAccessTokenTimeOut  int                `mapstructure:"DEFAULT_PANEL_ACCESS_TOKEN_TIMEOUT_IN_MINUTES"`
	DefaultPanelRefreshTokenTimeOut int                `mapstructure:"DEFAULT_PANEL_REFRESH_TOKEN_TIMEOUT_IN_MINUTES"`
	DefaultTickerInterval           int                `mapstructure:"DEFAULT_TICKER_INTERVAL" validate:"required"`
	MaxConnections                  int                `mapstructure:"MAX_CONNECTIONS" validate:"required"`
	MaxAppErrors                    int                `mapstructure:"MAX_APP_ERRORS"`
	MaxRetry                        int                `mapstructure:"MAX_RETRY"`
	MaxWait                         int                `mapstructure:"MAX_WAIT"`
	IgnoredAsssets                  []string           `mapstructure:"IGNORED_ASSETS"`
	SettingsApi                     string             `mapstructure:"SETTINGS_API" validate:"required"`
	Test                            string             `mapstructure:"TEST" validate:"required"`
	SecretRefreshToken              string             `mapstructure:"SECRET_REFRESH_TOKEN" validate:"required" json:"-"`
	SecretJWTToken                  string             `mapstructure:"SECRET_JWT_TOKEN" validate:"required" json:"-"`
	EtcdNodes                       int                `json:"-"`
	IsLeader                        chan bool          `json:"-"`
	Logger                          *zap.Logger        `json:"-"`
	LoggerSys                       *srslog.Writer     `json:"-"`
	MaxSpreadPct                    float64            `mapstructure:"MAX_SPREAD_PCT"  json:"-"`
	ServiceName                     string             `json:"ServiceName"`
	Database                        *database.CitusDSN `mapstructure:"DATABASE" validate:"required" json:"-"`
	PolicyRules                     *PolicyRules       `mapstructure:"POLICY_RULES" validate:"required"`
	PrefixService                   *prefix.Prefix     `json:"-"`
}

var config = &Config{}

func init() {
	viper.SetConfigName("config")
	viper.AddConfigPath("./internal/config/")
	viper.SetConfigType("json")

	viper.SetDefault("REDIS_PORT", 6379)
	viper.SetDefault("MAX_RETRY", 5)
	viper.SetDefault("MAX_WAIT", 2000)

	log.Println("Reading config...")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Fatal error config: %v", err)
	}

	log.Println("Unmarshalling config...")
	err = viper.Unmarshal(&config)
	if err != nil {
		log.Fatalf("Unable to decode into struct, %v", err)
	}

	if config.SysLog == "true" {

		config.LoggerSys, err = srslog.Dial("", "", srslog.LOG_INFO, "CEF0")
		if err != nil {
			log.Println("Error setting up syslog:", err)
			os.Exit(1)
		}

	}

	config.PrefixService = prefix.NewPrefixService()

	logs := zap.NewDevelopmentEncoderConfig() //zap.NewProduction()
	logs.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.Logger = zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(logs),
		zapcore.AddSync(colorable.NewColorableStdout()),
		zapcore.DebugLevel,
	))
	defer config.Logger.Sync()

	validate := validator.New()
	err = validate.Struct(config)
	if err != nil {
		log.Fatalf("Config validation failed, %v", err)
	}
}

func GetConfig() (*Config, error) {
	return config, nil
}
