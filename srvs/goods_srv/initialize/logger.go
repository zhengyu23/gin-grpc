package initialize

import (
	"go.uber.org/zap"
)

// InitLogger 初始化 zap全局Logger
func InitLogger() {
	//logger, _:= zap.NewProduction()	// Level: InfoLevel
	logger, _ := zap.NewDevelopment() // Level: DebugLevel
	zap.ReplaceGlobals(logger)        // 生成全局 logger
}
