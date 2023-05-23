package initialize

// 读配置信息
import (
	"encoding/json"
	"fmt"

	"github.com/nacos-group/nacos-sdk-go/clients"         // Nacos
	"github.com/nacos-group/nacos-sdk-go/common/constant" // Nacos
	"github.com/nacos-group/nacos-sdk-go/vo"              // Nacos
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"xrUncle/srvs/goods_srv/global"
)

func GetEnvInfo(env string) bool {
	viper.AutomaticEnv()
	return viper.GetBool(env)
}

func InitConfig() {

	// ***************** viper **********************

	// 从配置文件中读取对应的配置 - viper
	debug := GetEnvInfo("CHENBING_DEBUG") // 环境变量
	configFilePrefix := "config"
	configFileName := fmt.Sprintf("%s-pro.yaml", configFilePrefix)
	if debug { // 环境变量
		configFileName = fmt.Sprintf("%s-debug.yaml", configFilePrefix)
	}

	// 读取 config.yaml文件
	v := viper.New()
	//v.SetConfigFile("config-debug-Old.yaml") // go run 运行
	v.SetConfigFile(configFileName)
	err := v.ReadInConfig() // 读取
	if err != nil {
		panic(err)
	}

	if err := v.Unmarshal(&global.NacosConfig); err != nil {
		panic(err)
	}
	zap.S().Infof("配置信息:%v,", global.NacosConfig)

	// ***************** viper **********************

	// ***************** nacos **********************

	// Server 配置信息 Nacos
	sc := []constant.ServerConfig{
		{
			IpAddr: global.NacosConfig.Host,
			Port:   global.NacosConfig.Port,
		},
	}
	// Client 配置信息 Nacos
	cc := constant.ClientConfig{
		// We can create multiple clients with different namespaceId to support multiple namespace.
		// When namespace is public, fill in the blank string here.
		NamespaceId:         global.NacosConfig.Namespace,
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "tmp/nacos/log",   // 日志目录
		CacheDir:            "tmp/nacos/cache", // 缓存目录
		LogLevel:            "debug",           // 日志等级
	}

	// Another way of create config client for dynamic configuration (recommend)
	configClient, err := clients.NewConfigClient( // Nacos-Client
		vo.NacosClientParam{ // Nacos-vo
			ServerConfigs: sc,
			ClientConfig:  &cc,
		},
	)
	if err != nil {
		panic(err)
	}
	content, err := configClient.GetConfig(vo.ConfigParam{
		DataId: global.NacosConfig.DataId,
		Group:  global.NacosConfig.Group,
	})
	if err != nil {
		panic(err)
	}
	// 将得到的 json 数据转为 本地struct
	err = json.Unmarshal([]byte(content), &global.ServerConfig)
	if err != nil {
		zap.S().Fatalf("读取nacos配置失败：%s", err.Error())
	}
	//fmt.Println(&global.ServerConfig)
}

func InitConfigOld() {
	// 从配置文件中读取对应的配置
	debug := GetEnvInfo("CHENBING_DEBUG")
	configFilePrefix := "config"
	configFileName := fmt.Sprintf("goods_srv/%s-pro.yaml", configFilePrefix)
	if debug { // 环境变量
		configFileName = fmt.Sprintf("goods_srv/%s-debug.yaml", configFilePrefix)
	}

	v := viper.New()
	v.SetConfigFile(configFileName)
	//fmt.Println(configFileName == "user-srv/config-debug-old.yaml")
	if err := v.ReadInConfig(); err != nil {
		panic(err)
	}
	if err := v.Unmarshal(&global.ServerConfig); err != nil {
		panic(err)
	}
}
