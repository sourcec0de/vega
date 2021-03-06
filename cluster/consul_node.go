package cluster

import (
	"fmt"

	"github.com/vektra/vega"
)

type ConsulClusterNode struct {
	*clusterNode

	Config *ConsulNodeConfig

	routes  *consulRoutingTable
	service *vega.Service
}

const DefaultClusterPort = 8476

const DefaultPath = "/var/lib/vega"

type ConsulNodeConfig struct {
	AdvertiseAddr string
	ListenPort    int
	DataPath      string
	ConsulToken   string
	RoutingPrefix string
}

func (cn *ConsulNodeConfig) Normalize() error {
	if cn.ListenPort == 0 {
		cn.ListenPort = vega.DefaultPort
	}

	if cn.AdvertiseAddr == "" {
		ip, err := vega.GetPrivateIP()
		if err != nil {
			cn.AdvertiseAddr = "127.0.0.1"
		} else {
			cn.AdvertiseAddr = ip.String()
		}
	}

	if cn.DataPath == "" {
		cn.DataPath = DefaultPath
	}

	if cn.RoutingPrefix == "" {
		cn.RoutingPrefix = DefaultRoutingPrefix
	}

	return nil
}

func (cn *ConsulNodeConfig) ListenAddr() string {
	return fmt.Sprintf(":%d", cn.ListenPort)
}

func (cn *ConsulNodeConfig) AdvertiseID() string {
	return fmt.Sprintf("%s:%d", cn.AdvertiseAddr, cn.ListenPort)
}

func NewConsulClusterNode(config *ConsulNodeConfig) (*ConsulClusterNode, error) {
	if config == nil {
		config = &ConsulNodeConfig{}
	}

	err := config.Normalize()
	if err != nil {
		return nil, err
	}

	consul := NewConsulClient(config.ConsulToken)

	ct, err := NewConsulRoutingTable(config.RoutingPrefix, config.AdvertiseID(), consul)
	if err != nil {
		return nil, err
	}

	cn, err := NewClusterNode(config.DataPath, vega.NewRouter(ct))
	if err != nil {
		return nil, err
	}

	serv, err := vega.NewService(config.ListenAddr(), cn)
	if err != nil {
		cn.Close()
		return nil, err
	}

	ccn := &ConsulClusterNode{
		clusterNode: cn,
		Config:      config,
		routes:      ct,
		service:     serv,
	}

	for _, name := range cn.disk.MailboxNames() {
		ccn.Declare(name)
	}

	return ccn, nil
}

func (cn *ConsulClusterNode) Cleanup() error {
	return cn.routes.Cleanup()
}

func (cn *ConsulClusterNode) Accept() error {
	return cn.service.Accept()
}

func (cn *ConsulClusterNode) Close() error {
	cn.clusterNode.Close()
	return cn.service.Close()
}
