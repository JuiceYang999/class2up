package Config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type MqttConfig struct {
	IP             string `yaml:"IP"`
	Port           uint16 `yaml:"Port"`
	UserName       string `yaml:"UserName"`
	Password       string `yaml:"Password"`
	AutoReconnect  bool   `yaml:"AutoReconnect"`
	ConnectTimeout uint32 `yaml:"ConnectTimeout"`
}
type NCLinkConfig struct {
	QueryResponseQoS        byte `yaml:"QueryResponseQoS"`
	QueryRequestQoS         byte `yaml:"QueryRequestQoS"`
	SetResponseQoS          byte `yaml:"SetResponseQoS"`
	SetRequestQoS           byte `yaml:"SetRequestQoS"`
	ProbeQueryRequestQoS    byte `yaml:"ProbeQueryRequestQoS"`
	ProbeSetRequestQoS      byte `yaml:"ProbeSetRequestQoS"`
	ProbeQueryResponseQoS   byte `yaml:"ProbeQueryResponseQoS"`
	ProbeSetResponseQoS     byte `yaml:"ProbeSetResponseQoS"`
	RegisterRequestQoS      byte `yaml:"RegisterRequestQoS"`
	RegisterResponseQoS     byte `yaml:"RegisterResponseQoS"`
	ProbeVersionQoS         byte `yaml:"ProbeVersionQoS"`
	ProbeVersionResponseQoS byte `yaml:"ProbeVersionResponseQoS"`
}
type Log struct {
	File string `yaml:"File""`
}
type AppConfig struct {
	Mqtt           MqttConfig   `yaml:"Mqtt"`
	NCLink         NCLinkConfig `yaml:"NCLink"`
	Log            Log          `yaml:"Log""`
}

var appConfig *AppConfig = nil

func init() {
	configText, err := ioutil.ReadFile("./config.yaml")
	if err != nil {
		fmt.Printf("ioutil.ReadFile:%v\n", err)
		return
	}
	config := &AppConfig{}
	err = yaml.Unmarshal(configText, config)
	if err != nil {
		fmt.Printf("yaml.Unmarshal:%v\n", err.Error())
	} else {
		appConfig = config
	}
}
func GetAppConfig() *AppConfig {
	return appConfig
}
