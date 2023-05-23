package consul

import (
	"fmt"

	"github.com/hashicorp/consul/api"
)

type Register struct {
	Host string
	Port int
}

// RegistryClienter 注册客户端接口
type RegistryClienter interface {
	Register(address string, port int, name string, tags []string, id string) error
	DeRegister(serviceId string) error
}

// NewRegistryClient 生成注册客户端
func NewRegistryClient(host string, port int) RegistryClienter {
	return &Register{
		Host: host,
		Port: port,
	}
}

func (r *Register) Register(address string, port int, name string, tags []string, id string) error {
	cfg := api.DefaultConfig()
	cfg.Address = fmt.Sprintf("%s:%d", r.Host, r.Port)

	client, err := api.NewClient(cfg)
	if err != nil {
		panic(err)
	}
	// 健康检查对象
	check := &api.AgentServiceCheck{
		HTTP:                           fmt.Sprintf("http://%s:%d/health", address, port),
		Timeout:                        "5s",
		Interval:                       "60s",
		DeregisterCriticalServiceAfter: "10s",
	}

	// 注册对象
	registration := new(api.AgentServiceRegistration)
	registration.Name = name
	registration.ID = id
	registration.Port = port
	registration.Tags = tags
	registration.Address = address
	registration.Check = check // 添加健康检查到对象

	err = client.Agent().ServiceRegister(registration)
	if err != nil {
		panic(err)
	}
	return nil
}

func (r *Register) DeRegister(serviceId string) error {
	cfg := api.DefaultConfig()
	cfg.Address = fmt.Sprintf("%s:%d", r.Host, r.Port)

	client, err := api.NewClient(cfg)
	if err != nil {
		return err
	}
	err = client.Agent().ServiceDeregister(serviceId)
	return err
}
