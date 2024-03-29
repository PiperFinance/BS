package conf

import (
	"fmt"
	"net/url"
	"time"

	"github.com/caarlos0/env/v8"
	"go.uber.org/zap/zapcore"
)

type config struct {
	MongoUrl                     url.URL       `env:"MONGO_URL"`
	MongoDBName                  string        `env:"MONGO_DBNAME"`
	RedisUrl                     url.URL       `env:"REDIS_URL"`
	MongoSlowLoading             time.Duration `env:"MONGO_SLOW_LOAD" envDefault:"10s"`
	MongoMaxPoolSize             uint64        `env:"MONGO_MAX_POOL_SIZE" envDefault:"340"`
	MongoMaxTimeout              time.Duration `env:"MONGO_MAX_TIMEOUT" envDefault:"5m"`
	RedisDB                      int           `env:"REDIS_DB"`
	RedisMongoSlowLoading        time.Duration `env:"REDIS_SLOW_LOAD" envDefault:"3s"`
	MaxConcurrency               int           `env:"MAX_CONCURRENT_WORKER"`
	MaxTaskTimeOut               time.Duration `env:"MaxTaskTimeOut" envDefault:"10m"`
	AsynqMonUrl                  string        `env:"ASYNQ_MON_URL" envDefault:":7654"`
	ApiUrl                       string        `env:"API_URL" envDefault:":1300"`
	StartingBlockNumber          uint64        `env:"STARTING_BLOCK_NUMBER" envDefault:"3"`
	BlockHeadDelay               uint64        `env:"BLOCK_HEAD_DELAY" envDefault:"30"`
	SilenceRRCErrs               bool          `env:"RPC_ERROR_SILENCE" envDefault:"false"`
	SilenceParseErrs             bool          `env:"PARSE_ERROR_SILENCE" envDefault:"false"`
	SilenceMulticallErrs         bool          `env:"MULTICALL_ERROR_SILENCE" envDefault:"false"`
	SaveAllTransferLogs          bool          `env:"SAVE_ALL_TRANSFER_LOGS" envDefault:"false"`
	SupportedChains              []int64       `env:"SUPPORTED_CHAINS" envSeparator:","`
	MultiCallTimeout             time.Duration `env:"MULTICALL_TIMEOUT" envDefault:"1m"`
	ParseBlockTimeout            time.Duration `env:"PARSE_BLOCK_TIMEOUT" envDefault:"2m"`
	FetchBlockTimeout            time.Duration `env:"FETCH_BLOCK_TIMEOUT" envDefault:"5m"`
	UserBalUpdateTimeout         time.Duration `env:"USER_BAL_UPDATE_TIMEOUT" envDefault:"5m"`
	VaccumLogsTaskTimeout        time.Duration `env:"VACCUME_LOGS_TIMEOUT" envDefault:"5m"`
	TestTimeout                  time.Duration `env:"TEST_RPC_CONNECTION_TIMEOUT" envDefault:"15s"`
	ScanTaskTimeout              time.Duration `env:"SCAN_TASK_TIMEOUT" envDefault:"25s"`
	ProcessBlockTimeout          time.Duration `env:"PROCESS_BLOCK_TASK_TIMEOUT" envDefault:"2m"`
	UpdateOnlineUsersTaskTimeout time.Duration `env:"ONLINE_USERS_TASK_TIMEOUT" envDefault:"25s"`
	ReentrancyCheckTTL           time.Duration `env:"REENTRANCY_CHECK_TTL" envDefault:"24h"`
	LogLevel                     string        `env:"LOG_LEVEL" envDefault:"warn"`
	LogDir                       string        `env:"LOG_DIR" envDefault:"/var/bs/log"`
	MainnetDir                   string        `env:"MAINNET_DIR" envDefault:"/data/mainnets.json"`
	DEV                          bool          `env:"DEV_DEBUG" envDefault:"false"`
	LimitUsers                   bool          `env:"LIMIT_USERS" envDefault:"false"`
	UserAppURL                   url.URL       `env:"UA_URL" envDefault:"http://154.49.243.32:6003"`
	BlockScannerURL              url.URL       `env:"BS_URL" envDefault:"http://154.49.243.32:6001"`
	PortfolioScannerURL          url.URL       `env:"PS_URL" envDefault:"http://154.49.243.32:6002"`
	TokenPriceURL                url.URL       `env:"TP_URL" envDefault:"http://154.49.243.32:6004"`
	TokenListUrl                 url.URL       `env:"TOKEN_LIST_URL" envDefault:"https://github.com/PiperFinance/CD/blob/main/tokens/outVerified/all_tokens.json?raw=true"`
	TokensDir                    string        `env:"TOKEN_LIST_DIR" envDefault:"data/all_tokens.json"`
	IsLocal                      bool          `env:"IS_LOCAL" envDefault:"false"`
	IsPlayground                 bool          `env:"IS_PLAYGROUND" envDefault:"false"`
	ZapLogLevel                  zapcore.Level
}

var Config config

func LoadConfig() {
	if err := env.Parse(&Config); err != nil {
		panic(fmt.Sprintf("%+v", err))
	}
	if lvl, err := zapcore.ParseLevel(Config.LogLevel); err != nil {
		panic(fmt.Sprintf("%+v", err))
	} else {
		Config.ZapLogLevel = lvl
	}
}
