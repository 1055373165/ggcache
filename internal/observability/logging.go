package observability

import (
    "github.com/1055373165/ggcache/pkg/common/logger"
    "github.com/sirupsen/logrus"
)

// SetupLogging configures global logger with desired level & json toggle.
func SetupLogging(level logrus.Level, json bool) {
    logger.Init(level, json)
}
