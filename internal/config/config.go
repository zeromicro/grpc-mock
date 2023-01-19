package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf

	ProxyPort string `json:",default=:9999"`

	BizRedis redis.RedisConf

	MockEnableKey     string `json:",default=mock"`
	MockEnableValue   string `json:",default=yes"`
	MockDisableValue  string `json:",default=no"`
	MockCaseKey       string `json:",default=case"`
	MockCustomCaseKey string `json:",default=custom_case"`
	TestedAppNameKey  string `json:",default=tested_app_name"`
}
