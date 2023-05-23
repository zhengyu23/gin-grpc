package initialize

import (
	"encoding/json"
	"fmt"
	"xrUncle/webs/goods-web/global"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func GetEnvInfo(env string) bool {
	viper.AutomaticEnv()
	return viper.GetBool(env)
}

func InitConfig() {

	// ***************** viper **********************

	// 判断运行环境为debug/pro,以此使用不同yaml文件
	debug := GetEnvInfo("CHENBING_DEBUG")
	configFilePrefix := "config"
	configFileName := fmt.Sprintf("%s-pro.yaml", configFilePrefix)
	if debug { // 环境变量
		configFileName = fmt.Sprintf("%s-debug.yaml", configFilePrefix)
	}

	// 读取 config.yaml文件
	v := viper.New()
	v.SetConfigFile(configFileName) // 设置 yaml文件目录
	err := v.ReadInConfig()         // 读取 yaml文件内容
	if err != nil {
		panic(err)
	}
	if err := v.Unmarshal(global.NacosConfig); err != nil {
		panic(err)
	} // 解码 yaml文件内容
	zap.S().Infof("配置信息:%v,", global.NacosConfig)

	// ***************** viper **********************

	// ***************** nacos **********************

	// 1. 从配置文件中读取 nacos的服务器配置
	sc := []constant.ServerConfig{
		{
			IpAddr: global.NacosConfig.Host,
			Port:   global.NacosConfig.Port,
		},
	} // server

	// 2. 从配置文件中读取发往nacos的客户端配置
	cc := constant.ClientConfig{
		NamespaceId:         global.NacosConfig.Namespace, //we can create multiple clients with different namespaceId to support multiple namespace.When namespace is public, fill in the blank string here.
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "tmp/nacos/log", // 需要注意路径不要写成 /tmp/nacos/log
		CacheDir:            "tmp/nacos/cache",
		LogLevel:            "debug",
	}

	// 3. 连接nacos的客户端配置
	// Another way of create config client for dynamic configuration (recommend)
	configClient, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ServerConfigs: sc,
			ClientConfig:  &cc,
		},
	)
	if err != nil {
		panic(err)
	}
	// 4. 客户端从nacos中获取配置信息
	content, err := configClient.GetConfig(vo.ConfigParam{
		DataId: global.NacosConfig.DataId,
		Group:  global.NacosConfig.Group,
	})
	if err != nil {
		panic(err)
	}

	// 5. 将获取到的信息解码
	err = json.Unmarshal([]byte(content), &global.ServerConfig) // json 转 struct
	if err != nil {
		zap.S().Fatalf("读取nacos配置失败：%s", err.Error())
	}
	fmt.Println(&global.ServerConfig)
}
