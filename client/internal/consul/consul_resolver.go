package consul

import (
	"errors"
	"fmt"
	"github.com/hashicorp/consul/api"
	"google.golang.org/grpc/resolver"
	"regexp"
	"sync"
)

const (
	defaultPort = "8500"
)

var (
	errMissingAddr = errors.New("consul resolver: missing address")

	errAddrMisMatch = errors.New("consul resolver: invalied uri")

	errEndsWithColon = errors.New("consul resolver: missing port after port-separator colon")

	regexConsul, _ = regexp.Compile("^([A-z0-9.]+)(:[0-9]{1,5})?/([A-z_]+)$")
)

func Init() {
	fmt.Printf("calling consul init\n")
	resolver.Register(NewBuilder())
}

type consulBuilder struct {
}

type consulResolver struct {
	address              string
	wg                   sync.WaitGroup
	cc                   resolver.ClientConn
	name                 string
	disableServiceConfig bool
	lastIndex            uint64
}

func NewBuilder() resolver.Builder {
	return &consulBuilder{}
}

func (cb *consulBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOption) (resolver.Resolver, error) {

	fmt.Printf("calling consul build\n")
	fmt.Printf("target: %v\n", target)
	host, port, name, err := parseTarget(fmt.Sprintf("%s/%s", target.Authority, target.Endpoint))
	if err != nil {
		return nil, err
	}

	cr := &consulResolver{
		address:              fmt.Sprintf("%s%s", host, port),
		name:                 name,
		cc:                   cc,
		disableServiceConfig: opts.DisableServiceConfig,
		lastIndex:            0,
	}

	cr.wg.Add(1)
	go cr.watcher()
	return cr, nil

}

func (cr *consulResolver) watcher() {
	fmt.Printf("calling consul watcher\n")
	config := api.DefaultConfig()
	config.Address = cr.address
	client, err := api.NewClient(config)
	if err != nil {
		fmt.Printf("error create consul client: %v\n", err)
		return
	}

	for {
		services, metainfo, err := client.Health().Service(cr.name, cr.name, true, &api.QueryOptions{WaitIndex: cr.lastIndex})
		if err != nil {
			fmt.Printf("error retrieving instances from Consul: %v", err)
		}

		cr.lastIndex = metainfo.LastIndex
		var newAddrs []resolver.Address
		for _, service := range services {
			addr := fmt.Sprintf("%v:%v", service.Service.Address, service.Service.Port)
			newAddrs = append(newAddrs, resolver.Address{Addr: addr})
		}
		fmt.Printf("adding service addrs\n")
		fmt.Printf("newAddrs: %v\n", newAddrs)
		cr.cc.NewAddress(newAddrs)
		cr.cc.NewServiceConfig(cr.name)
	}

}

func (cb *consulBuilder) Scheme() string {
	return "consul"
}

func (cr *consulResolver) ResolveNow(opt resolver.ResolveNowOption) {
}

func (cr *consulResolver) Close() {
}

func parseTarget(target string) (host, port, name string, err error) {

	fmt.Printf("target uri: %v\n", target)
	if target == "" {
		return "", "", "", errMissingAddr
	}

	if !regexConsul.MatchString(target) {
		return "", "", "", errAddrMisMatch
	}

	groups := regexConsul.FindStringSubmatch(target)
	host = groups[1]
	port = groups[2]
	name = groups[3]
	if port == "" {
		port = defaultPort
	}
	return host, port, name, nil
}
