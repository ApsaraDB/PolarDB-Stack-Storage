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

package main

import (
	"flag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"os"
	_ "polardb-sms/docs"
	smslog "polardb-sms/pkg/log"
	"polardb-sms/pkg/manager/application/service"
	"polardb-sms/pkg/manager/cluster/ink8s"
	"polardb-sms/pkg/manager/config"
	"polardb-sms/pkg/manager/interfaces"
	"polardb-sms/pkg/manager/interfaces/controller/anticorrosion"
	"polardb-sms/pkg/manager/msgserver"
	"polardb-sms/pkg/version"

	"github.com/gin-gonic/gin"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/sample-controller/pkg/signals"
)

const (
	LockName         = "polardb-sms-manager-leader"
	DefaultNamespace = "kube-system"
	MetaDBConfigMap  = "metabase-config"
	LogFile          = "polardb-sms-manager"
)

var (
	Vip        = flag.String("vip", "198.19.61.1", "vip")
	NodeId     = flag.String("node_id", "", "node id")
	AgentPort  = flag.String("agent_port", "18888", "agent port")
	NodeIp     = flag.String("node_ip", "", "node ip")
	servePort  = flag.String("serve_port", "2002", "serve port for gin")
	master     = flag.String("master", "", "Master URL to build a client config from. Either this or kubernetes config needs to be set if the provisioner is being run out of cluster.")
	kubeConfig = flag.String("kube_config", "", "Absolute path to the kubernetes config file. Either this or master needs to be set if the provisioner is being run out of cluster.")
	logDir     = flag.String("logDir", "/var/log/alicloud", "sms log dir")
	logLevel   = flag.String("log_level", "DEBUG", "log level: DEBUG|INFO|ERROR|WARN")
	clientSet  kubernetes.Interface
)

func initKubeConfig(master, kubeConfig string) kubernetes.Interface {
	var (
		err       error
		restConf  *rest.Config
		clientSet kubernetes.Interface
	)

	// get the KUBECONFIG from env if specified (useful for local/debug cluster)
	kubeConfigEnv := os.Getenv("KUBECONFIG")
	if kubeConfigEnv != "" {
		smslog.Info("Found KUBECONFIG environment variable set, using that..")
		kubeConfig = kubeConfigEnv
	}

	if master != "" || kubeConfig != "" {
		smslog.Info("Either master or kubeConfig specified. building kube restConf from that..")
		restConf, err = clientcmd.BuildConfigFromFlags(master, kubeConfig)
	} else {
		smslog.Info("Building kube configs for running in cluster...")
		restConf, err = rest.InClusterConfig()
	}

	if err != nil {
		smslog.Fatalf("Failed to create restConf: %v", err)
	}

	if clientSet, err = kubernetes.NewForConfig(restConf); err != nil {
		smslog.Fatalf("Failed to create client: %v", err)
	}

	return clientSet
}

func initLog() {
	// print client-go balancer_ha log
	klog.InitFlags(nil)
	gin.DisableConsoleColor()
	smslog.InitLogger(*logDir, LogFile, smslog.LogLevel(*logLevel))
}

func init() {
	flag.Parse()
	initLog()
	clientSet = initKubeConfig(*master, *kubeConfig)
	dbCm, err := clientSet.CoreV1().ConfigMaps(DefaultNamespace).Get(MetaDBConfigMap, metav1.GetOptions{})
	if err != nil || dbCm == nil {
		panic("cannot find the metabase-config config map")
	}
	config.Init("", dbCm.Data["metabase.yml"], *AgentPort, clientSet)
	version.LogVersion()
}

// @title PolarDB Storage Manage System
// @version 1.0
// @description v1.8 存储管理系统

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host 127.0.0.1:2002
// @BasePath /v2
func main() {
	defer smslog.LogPanic()
	stopCh := signals.SetupSignalHandler()
	interfaces.Init(*NodeIp, *servePort)
	// Connect Agent Server
	msgServer := msgserver.NewMessageServer(config.ClusterConf.Nodes)
	go msgServer.Run()

	// Leader Election Server
	server := ink8s.NewServer(&ink8s.InK8sServerConfig{
		Vip:             *Vip,
		DistributedLock: LockName,
		ClientSet:       clientSet,
		Ip:              *NodeIp,
		Id:              *NodeId,
		StopCh:          stopCh,
	})

	//Add workflowEngine
	server.AddRunner(service.GetWorkflowEngine())

	// PureSoft CSI Server
	cfg := anticorrosion.NewControllerConfig(stopCh, *NodeId, *NodeIp, clientSet)
	csiServer := anticorrosion.NewStorageController(cfg)
	server.AddRunner(csiServer)

	go server.Run()
	<-stopCh

	interfaces.Stop()
	defer smslog.Flush()
	os.Exit(0)
}
