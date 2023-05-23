package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"xrUncle/webs/goods-web/global"
	"xrUncle/webs/goods-web/initialize"
	"xrUncle/webs/goods-web/utils"
	"xrUncle/webs/goods-web/utils/register/consul"

	uuid "github.com/satori/go.uuid"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func main() {
	// 初始化日志库 - zap
	initialize.InitLogger()
	// 初始化配置文件 - viper+nacos
	initialize.InitConfig()
	// 初始化 Routers - gin
	Router := initialize.Routers()
	// 初始化validator翻译 - gin
	if err := initialize.InitTrans("zh"); err != nil {
		panic(err)
	}
	// 初始化Srv的连接 - gRPC
	initialize.InitSrvConn()
	// 初始化sentinel
	initialize.InitSentinel()

	// 获取动态端口
	// viper 获取环境变量
	viper.AutomaticEnv()
	//如果是本地开发环境端口号固定，线上环境启动获取端口号
	debug := viper.GetBool("CHENBING_DEBUG")
	//debug = false
	if !debug { // default of debug is true
		port, err := utils.GetFreePort()
		if err == nil {
			global.ServerConfig.Port = port
		}
	}

	// 本服务的 consul 注册
	register_client := consul.NewRegistryClient(global.ServerConfig.ConsulInfo.Host, global.ServerConfig.ConsulInfo.Port)
	serviceId := fmt.Sprintf("%s", uuid.NewV4())
	err := register_client.Register(global.ServerConfig.Host, global.ServerConfig.Port, global.ServerConfig.Name, global.ServerConfig.Tags, serviceId)
	if err != nil {
		zap.S().Panic("服务注册失败:", err.Error())
	}

	// 协程 启动服务器
	zap.S().Debugf("启动服务器, 端口： %d", global.ServerConfig.Port)
	go func() {
		if err := Router.Run(fmt.Sprintf(":%d", global.ServerConfig.Port)); err != nil {
			zap.S().Panic("启动失败:", err.Error())
		}
	}()

	// 接收终止信号 - 优雅退出
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // 阻塞
	// 优雅退出后续 , 注销 consul
	if err = register_client.DeRegister(serviceId); err != nil {
		zap.S().Info("注销失败:", err.Error())
	} else {
		zap.S().Info("注销成功:")
	}
}
