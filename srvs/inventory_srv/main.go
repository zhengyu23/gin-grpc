package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"xrUncle/srvs/inventory_srv/handler"
	"xrUncle/srvs/inventory_srv/utils/register/consul"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	"xrUncle/srvs/inventory_srv/global"
	"xrUncle/srvs/inventory_srv/utils"

	"xrUncle/srvs/inventory_srv/initialize"
	"xrUncle/srvs/inventory_srv/proto"

	uuid "github.com/satori/go.uuid"
)

func main() {
	IP := flag.String("ip", "0.0.0.0", "ip地址")
	Port := flag.Int("port", 0, "端口号")

	// 初始化
	initialize.InitLogger()
	initialize.InitConfig()
	initialize.InitDB()
	zap.S().Info(global.ServerConfig)

	flag.Parse()
	zap.S().Info("ip: ", *IP)
	if *Port == 0 { // 初始化 port
		*Port, _ = utils.GetFreePort()
	}
	zap.S().Info("port: ", *Port)

	// 1.实例化 server对象
	server := grpc.NewServer()
	// 2. 注册服务 tcp方式监听端口
	proto.RegisterInventoryServer(server, &handler.InventoryServer{})
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *IP, *Port))
	if err != nil {
		panic("failed to listen:" + err.Error())
	}

	// 健康检查
	register_client := consul.NewRegistryClient(global.ServerConfig.ConsulInfo.Host, global.ServerConfig.ConsulInfo.Port)
	serviceId := fmt.Sprintf("%s", uuid.NewV4())
	err = register_client.Register(global.ServerConfig.Host, *Port, global.ServerConfig.Name, global.ServerConfig.Tags, serviceId)
	if err != nil {
		zap.S().Panic("服务注册失败:", err.Error())
	}
	zap.S().Debugf("启动服务器, 端口： %d", *Port)

	// 注册服务健康检查 -> Check
	grpc_health_v1.RegisterHealthServer(server, health.NewServer())

	// 启动服务
	go func() {
		err = server.Serve(lis)
		if err != nil {
			panic("failed to start grpc:" + err.Error())
		}
	}()

	// 订单和库存 数据一致性板块
	// 1、监听库存归还topic
	// 2、监听到后的操作
	// 3、重复归还问题？ —— 接口确保幂等性(确保重复发送的消息不会导致订单库存归还多次；确保没有扣减的库存不要归还)
	//- 确保幂等性 —— 设计一张表，记录着详细订单扣减细节和归还细节
	c, _ := rocketmq.NewPushConsumer(
		consumer.WithNameServer([]string{"192.168.31.142:9876"}),
		consumer.WithGroupName("chenbing"),
	)

	if err := c.Subscribe("order_reback",
		consumer.MessageSelector{}, handler.AutoReback); err != nil {
		fmt.Println("读取消息失败")
	}
	_ = c.Start()

	// 接收终止信号
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	_ = c.Shutdown()
	if err = register_client.DeRegister(serviceId); err != nil {
		zap.S().Info("注销失败:", err.Error())
	} else {
		zap.S().Info("注销成功:")
	}
}
