package consul

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	"time"
)

type ConsulService struct {
	IP   string
	Port int
	Tag  []string
	Name string
}

func RegitserService(ca string, cs *ConsulService) {

	//register consul
	consulConfig := api.DefaultConfig()
	consulConfig.Address = ca
	client, err := api.NewClient(consulConfig)
	if err != nil {
		fmt.Printf("NewClient error\n%v", err)
		return
	}
	agent := client.Agent()
	interval := time.Duration(10) * time.Second
	deregister := time.Duration(1) * time.Minute

	reg := &api.AgentServiceRegistration{
		ID:      fmt.Sprintf("%v-%v-%v", cs.Name, cs.IP, cs.Port), // 服务节点的名称
		Name:    cs.Name,                                          // 服务名称
		Tags:    cs.Tag,                                           // tag，可以为空
		Port:    cs.Port,                                          // 服务端口
		Address: cs.IP,                                            // 服务 IP
		Check: &api.AgentServiceCheck{ // 健康检查
			Interval:                       interval.String(),                                // 健康检查间隔
			GRPC:                           fmt.Sprintf("%v:%v/%v", cs.IP, cs.Port, cs.Name), // grpc 支持，执行健康检查的地址，service 会传到 Health.Check 函数中
			DeregisterCriticalServiceAfter: deregister.String(),                              // 注销时间，相当于过期时间
		},
	}

	fmt.Printf("registing to %v\n", ca)
	if err := agent.ServiceRegister(reg); err != nil {
		fmt.Printf("Service Register error\n%v", err)
		return
	}

}
