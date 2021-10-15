/* 
*Copyright (c) 2019-2021, Alibaba Group Holding Limited;
*Licensed under the Apache License, Version 2.0 (the "License");
*you may not use this file except in compliance with the License.
*You may obtain a copy of the License at

*   http://www.apache.org/licenses/LICENSE-2.0

*Unless required by applicable law or agreed to in writing, software
*distributed under the License is distributed on an "AS IS" BASIS,
*WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
*See the License for the specific language governing permissions and
*limitations under the License.
 */


package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"log"
	"math/rand"
	"polardb-sms/pkg/common"
	smslog "polardb-sms/pkg/log"
	"strings"
	"sync"
	"time"

	"github.com/Unknwon/goconfig"
)

const (
	ConfigPath             = "/var/lib/polardb-sms/manager/manager.conf"
	LogFile                = "/var/log/polardb-sms/manager/manager.log"
	DefaultPort            = "2002"
	ServerSection          = "server"
	AppPort                = "port"
	ServerId               = "id"
	LogSection             = "log"
	LogLevel               = "level"
	FastHA                 = "fastHA"
	DBSection              = "database"
	DBUser                 = "user"
	DBPassword             = "password"
	DBHost                 = "host"
	DBPort                 = "port"
	DBSchema               = "schema"
	DBSchemaValue          = "polardb_sms"
	LocalHost              = "localhost"
	Separator              = ":"
	DBConnStrFormat        = "%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local"
	ClusterSection         = "cluster"
	Mode                   = "mode"
	Vip                    = "vip"
	Nodes                  = "nodes"
	HeartbeatMissTolerance = 180 * time.Second
)

type DBConfig struct {
	Metabase struct {
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
	}
}

type ServerConfig struct {
	Port   string
	Id     string
	Ip     string
	Mode   string
}

type LogConfig struct {
	Level string
}

type Node struct {
	Name              string    `json:"name"`
	Ip                string    `json:"ip"`
	Port              string    `json:"port"`
	LastHeartbeatTime time.Time `json:"last_heartbeat_time"`
}

func (c *Node) Active() bool {
	now := time.Now()
	if now.After(c.LastHeartbeatTime.Add(HeartbeatMissTolerance)) {
		return false
	}
	return true
}

type DeployMode string

type ClusterConfig struct {
	Vip        string
	Nodes      map[string]Node
	DeployMode string
	NodeCnt    int
}

func GetNodeByIp(ip string) *Node {
	for _, node := range GetAvailableNodes() {
		if node.Ip == ip {
			return &node
		}
	}
	return nil
}

func GetNodeById(nodeId string) *Node {
	n, ok := GetAvailableNodes()[nodeId]
	if ok {
		return &n
	}
	return nil
}

func GetOneNode() *Node {
	nodes := GetAvailableNodes()
	idx := rand.Intn(len(nodes))
	if idx >= len(nodes) {
		idx = 0
	}
	var id = 0
	for _, val := range nodes {
		if id == idx {
			return &val
		}
		id++
	}
	return nil
}

func GetAvailableNodes() map[string]Node {
	var availableNodes = make(map[string]Node)
	for _, node := range ClusterConf.Nodes {
		if !node.Active() {
			smslog.Warnf("not available cluster agent %s, last heartbeat time %v", node.Name, node.LastHeartbeatTime)
			if ServerConf.Mode == FastHA {
				continue
			}
		}
		availableNodes[node.Name] = node
	}
	return availableNodes
}

func AvailableNodes() int {
	nodes := len(GetAvailableNodes())
	if nodes < 1 {
		return 1
	}
	return nodes
}

var (
	ServerConf     ServerConfig
	DBConf         DBConfig
	LogConf        LogConfig
	ClusterConf    ClusterConfig
	ClientSet      kubernetes.Interface
	processingLock sync.Mutex
)

func Init(confPath, metaDbConfStr, agentPort string, clientSet kubernetes.Interface) {
	ClientSet = clientSet
	parseFromPath(confPath)
	parseDbConfig(metaDbConfStr)
	parseClusterConfig(agentPort)
	smslog.Debugf("server config %v, cluster config %v", ServerConf, ClusterConf)
}

func parseDbConfig(metaDbConfStr string) {
	smslog.Debugf("metaDbConf %s", metaDbConfStr)
	DBConf = DBConfig{}
	err := yaml.Unmarshal([]byte(metaDbConfStr), &DBConf)
	if err != nil {
		smslog.Fatalf("can not parse metadb conf %s", metaDbConfStr)
	}
	smslog.Infof("DBConfig %v", DBConf)
}

func parseClusterConfig(agentPort string) {
	ClusterConf = ClusterConfig{
		Vip:        "",
		Nodes:      map[string]Node{},
		DeployMode: "",
		NodeCnt:    0,
	}
	nodeList, err := common.GetNodes(ClientSet)
	if err != nil {
		log.Fatal("can not get nodes, panic")
		return
	}
	for _, n := range nodeList {
		node := Node{
			Name: n.Name,
			Port: agentPort,
		}
		for _, addr := range n.Status.Addresses {
			if addr.Type == corev1.NodeInternalIP {
				node.Ip = addr.Address
			}
		}
		ClusterConf.Nodes[n.Name] = node
	}
}

func parseFromPath(path string) {
	if path == "" {
		path = ConfigPath
	}
	config, err := goconfig.LoadConfigFile(ConfigPath)
	if err != nil {
		smslog.Errorf("load config file error from path: %s err %v", path, err)
		return
	}
	parse(config)
}

func ParseFromContent(contents string) {
	config, err := goconfig.LoadFromReader(strings.NewReader(contents))
	if err != nil {
		smslog.Errorf("load config file error from contents: %s err %v", contents, err)
		return
	}
	parse(config)
}

func DBConnStr() string {
	return fmt.Sprintf(DBConnStrFormat,
		DBConf.Metabase.User,
		DBConf.Metabase.Password,
		DBConf.Metabase.Host,
		DBConf.Metabase.Port,
		DBSchemaValue)
}

func parse(conf *goconfig.ConfigFile) {
	sConf, err := conf.GetSection(ServerSection)
	if err != nil {
		smslog.Errorf("Error when get server section: %v", err)
	} else {
		parseServerConf(sConf)
	}

	logConf, err := conf.GetSection(LogSection)
	if err != nil {
		smslog.Errorf("Error when get log section: %v", err)
	} else {
		parseLogConf(logConf)
	}
}

func parseServerConf(confMap map[string]string) {
	ServerConf = ServerConfig{}
	ServerConf.Port, _ = confMap[AppPort]
	ServerConf.Id, _ = confMap[ServerId]
	ServerConf.Mode, _ = confMap[Mode]
	smslog.Infof("Server config is %v", ServerConf)
}

func parseLogConf(confMap map[string]string) {
	//parse Log config
	LogConf = LogConfig{}
	LogConf.Level, _ = confMap[LogLevel]
	smslog.Infof("Log config is %v", LogConf)
}

func parseClusterConf(confMap map[string]string) {
	//parse cluster config
	ClusterConf = ClusterConfig{
		Nodes: make(map[string]Node),
	}
	ClusterConf.Vip, _ = confMap[Vip]
	nodeStr, ok := confMap[Nodes]
	if !ok {
		smslog.Infof("Error parse config conf %v", confMap)
	} else {
		temp := strings.Split(nodeStr, ",")
		for _, val := range temp {
			arr := strings.Split(val, "=")
			vals := strings.Split(arr[1], Separator)
			nodeName := strings.Trim(arr[0], "")
			ClusterConf.Nodes[nodeName] = Node{
				Ip:   vals[0],
				Port: vals[1],
				Name: nodeName,
			}
		}
	}
	ClusterConf.DeployMode, _ = confMap[Mode]
	ClusterConf.NodeCnt = len(ClusterConf.Nodes)

	smslog.Infof("Cluster config is %v", ClusterConf)
}

func UpdateAgentNodeHeartbeatTime(nodeId string) {
	if node, exist := ClusterConf.Nodes[nodeId]; exist {
		processingLock.Lock()
		node.LastHeartbeatTime = time.Now()
		ClusterConf.Nodes[nodeId] = node
		processingLock.Unlock()
	}
}
