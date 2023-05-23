package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	uuid "github.com/satori/go.uuid"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	"xrUncle/srvs/goods_srv/global"
	"xrUncle/srvs/goods_srv/utils"
	"xrUncle/srvs/goods_srv/utils/register/consul"

	"xrUncle/srvs/goods_srv/handler"
	"xrUncle/srvs/goods_srv/initialize"
	"xrUncle/srvs/goods_srv/proto"
)

func main() {
	// 初始化
	initialize.InitLogger() // 初始化全局Logger - zap
	initialize.InitConfig() // 初始化配置文件 - viper nacos
	initialize.InitDB()     // 初始化数据库 - MySQL
	initialize.InitEs()     // 初始化数据库 - ES

	zap.S().Info(global.ServerConfig) // 输出当前服务器地址

	// 接收命令行参数
	IP := flag.String("ip", global.ServerConfig.Host, "ip地址")
	Port := flag.Int("port", 0, "端口号")
	flag.Parse()
	zap.S().Info("ip: ", *IP)
	if *Port == 0 { // 初始化 port
		*Port, _ = utils.GetFreePort() // 动态获取端口号
	}
	zap.S().Info("port: ", *Port)

	// 1. 初始化服务对象 - gRPC
	server := grpc.NewServer()
	// 2. 把GoodsServer注册进服务对象 - gRPC
	proto.RegisterGoodsServer(server, &handler.GoodsServer{})
	// 3. tcp方式开启端口 - gRPC
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *IP, *Port))
	if err != nil {
		panic("failed to listen:" + err.Error())
	}

	// ***************** consul **********************

	// 本服务器的consul注册
	register_client := consul.NewRegistryClient(global.ServerConfig.ConsulInfo.Host, global.ServerConfig.ConsulInfo.Port)
	serviceId := fmt.Sprintf("%s", uuid.NewV4())
	err = register_client.Register(global.ServerConfig.Host, *Port, global.ServerConfig.Name, global.ServerConfig.Tags, serviceId)
	if err != nil {
		zap.S().Panic("服务注册失败:", err.Error())
	}
	zap.S().Debugf("启动服务器, 端口： %d", *Port)

	// 注册服务健康检查 - gRPC
	grpc_health_v1.RegisterHealthServer(server, health.NewServer())

	// ***************** consul **********************

	go func() {
		// 4. 启动服务 - gRPC
		err = server.Serve(lis)
		if err != nil {
			panic("failed to start grpc:" + err.Error())
		}
	}()

	// 接收终止信号
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	if err = register_client.DeRegister(serviceId); err != nil {
		zap.S().Info("注销失败:", err.Error())
	} else {
		zap.S().Info("注销成功:")
	}
}
