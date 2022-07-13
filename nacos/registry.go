package nacos

import (
	"net/url"
	"strconv"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
)

func Init() {

}

func CreateConfigClient(timeout uint64, namespaceId, username, password string, urls ...url.URL) (config_client.IConfigClient, error) {
	c, err := clients.CreateConfigClient(map[string]interface{}{
		"serverConfigs": CreateServerConfig(urls...),
		"clientConfig":  CreateClientConifg(timeout, namespaceId, username, password),
	})
	return c, err
}

func CreateNamingClient(timeout uint64, namespaceId, username, password string, urls ...url.URL) (naming_client.INamingClient, error) {
	c, err := clients.CreateNamingClient(map[string]interface{}{
		"serverConfigs": CreateServerConfig(urls...),
		"clientConfig":  CreateClientConifg(timeout, namespaceId, username, password),
	})
	return c, err
}

func CreateClientConifg(timeout uint64, namespaceId, username, password string) constant.ClientConfig {
	c := constant.ClientConfig{}
	return c
}

func CreateServerConfig(urls ...url.URL) []constant.ServerConfig {
	serverConfigs := make([]constant.ServerConfig, 0)
	for _, u := range urls {
		port, _ := strconv.ParseUint(u.Port(), 10, 64)
		serverConfig := constant.ServerConfig{
			Scheme: u.Scheme,
			Port:   port,
			IpAddr: u.Host,
			// todo
			ContextPath: "",
		}
		serverConfigs = append(serverConfigs, serverConfig)
	}
	return serverConfigs
}
