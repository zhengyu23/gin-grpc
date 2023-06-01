package initialize

import "go.uber.org/zap"

// 初始化 logger
func InitLogger() {
	//logger, _:= zap.NewProduction()	// Level: InfoLevel
	logger, _ := zap.NewDevelopment() // Level: DebugLevel
	zap.ReplaceGlobals(logger)        // 生成全局 logger
}
