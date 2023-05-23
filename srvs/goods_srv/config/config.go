package config

// 配置文件模板

// NacosConfig Nacos模板
type NacosConfig struct {
	Host      string `mapstructure:"host"`
	Port      uint64 `mapstructure:"port"`
	Namespace string `mapstructure:"namespace"`
	User      string `mapstructure:"user"`
	Password  string `mapstructure:"password"`
	DataId    string `mapstructure:"dataid"`
	Group     string `mapstructure:"group"`
}

// MysqlConfig 数据库模板
type MysqlConfig struct {
	Host     string `mapstructure:"host" json:"host"`
	Port     int    `mapstructure:"port" json:"port"`
	Name     string `mapstructure:"db" json:"db"`
	User     string `mapstructure:"user" json:"user"`
	Password string `mapstructure:"password" json:"password"`
}

// EsConfig Elasticsearch模板
type EsConfig struct {
	Host string `mapstructure:"host" json:"host"`
	Port int    `mapstructure:"port" json:"port"`
}

type ConsulConfig struct {
	Host string `mapstructure:"host" json:"host"`
	Port int    `mapstructure:"port" json:"port"`
}

// ServerConfig 服务器模板
type ServerConfig struct {
	Name        string       `mapstructure:"name" json:"name"`
	Host        string       `mapstructure:"host" json:"host"`
	MysqlConfig MysqlConfig  `mapstructure:"mysql" json:"mysql"`
	ConsulInfo  ConsulConfig `mapstructure:"consul" json:"consul"`
	Tags        []string     `mapstructure:"name" json:"tags"`
	EsInfo      EsConfig     `mapstructure:"es" json:"es"`
}
