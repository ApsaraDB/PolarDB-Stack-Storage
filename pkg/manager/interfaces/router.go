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


package interfaces

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	smslog "polardb-sms/pkg/log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
	"polardb-sms/pkg/manager/interfaces/controller"
)

var router *gin.Engine
var srv *http.Server

func Init(ip, port string) {
	gin.SetMode(gin.ReleaseMode)
	router = gin.Default()
	router.Use(newLogger())
	//router.Use(service.IdempotentCheck)
	registerRoutes()
	srv = &http.Server{
		Addr:    ip + ":" + port,
		Handler: router,
	}
	go Start()
}

func Start() {
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		smslog.Fatalf("listen: %s\n", err)
	}
}

func Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		smslog.Fatal("Server forced to shutdown:", err)
		return
	}

	smslog.Infof("Server exiting")
}

func registerRoutes() {
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"welcome": "Test"})
	})
	// use ginSwagger middleware to
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	clusterLunController := controller.NewClusterLunController()
	router.POST("/cluster-luns", clusterLunController.CreateClusterLun)
	router.GET("/cluster-luns", clusterLunController.QueryClusterLuns)
	router.GET("/cluster-luns/:wwid", clusterLunController.SearchClusterLun)
	router.POST("/cluster-luns/format", clusterLunController.FormatClusterLun)
	router.POST("/cluster-luns/expand", clusterLunController.ExpandClusterLun)
	router.POST("/cluster-luns/same-san", clusterLunController.LunsFromSameSan)
	router.DELETE("/cluster-luns/:wwid", clusterLunController.DeleteClusterLun)

	clusterLvController := controller.NewClusterLvController()
	router.POST("/cluster-lvs", clusterLvController.CreateClusterLv)
	router.GET("/cluster-lvs", clusterLvController.QueryClusterLvs)
	router.GET("/cluster-lvs/:name", clusterLvController.SearchClusterLv)
	router.POST("/cluster-lvs/format", clusterLvController.FormatClusterLv)
	router.POST("/cluster-lvs/expand", clusterLvController.ExpandClusterLv)
	router.POST("/cluster-lvs/fs-expand", clusterLvController.ExpandClusterLvForFs)
	router.DELETE("/cluster-lvs/:name", clusterLvController.DeleteClusterLv)

	eventController := controller.NewEventController()
	router.POST("/events", eventController.Upload)
	router.POST("/events/batch", eventController.BatchUpload)

	volumeController := controller.NewVolumeController()
	router.POST("/scsi/rescan", volumeController.Rescan)

	clusterTaskController := controller.NewClusterTaskController()
	router.GET("/workflows/:workflowId", clusterTaskController.FindWorkflowById)

	prCheckController := controller.NewPrCheckController()
	router.POST("/pr-check/overall", prCheckController.CheckOverallCapabilities)
	router.POST("/pr-check/detail", prCheckController.CheckDetailCapabilities)
	router.GET("/pr-check/detail", prCheckController.QueryDetailCapabilities)

	pvcController := controller.NewPvcController()
	router.GET("/pvcs", pvcController.QueryPvcs)
	router.GET("/pvcs/:pvcDynamicName", func(c *gin.Context) {
		if strings.HasPrefix(c.Request.RequestURI, "/pvcs/ready") {
			pvcController.PvcIsReady(c)
			return
		}
		if strings.HasPrefix(c.Request.RequestURI, "/pvcs/topo") {
			pvcController.QueryVolumePermissionTopo(c)
			return
		}
		if strings.HasPrefix(c.Request.RequestURI, "/pvcs/expand") {
			pvcController.PvcExpandedFs(c)
			return
		}
		pvcController.QueryPvc(c)
	})
	router.DELETE("/pvcs/:name", pvcController.DeletePvc)
	router.POST("/pvcs", pvcController.CreatePvcWithVolume)
	router.POST("/pvcs/use", pvcController.UsePvc)
	router.POST("/pvcs/lock", pvcController.PvcSetWriteLock)
	router.GET("/hack/pvcs", pvcController.ModifyPvc)
	router.POST("/pvcs/expand", pvcController.PvcExpandFs)
	router.POST("/pvcs/release", pvcController.ReleasePvc)
	router.POST("/pvcs/formatAndLock", pvcController.FormatAndLock)
	router.POST("/pvcs/format", pvcController.Format)

	agentController := controller.NewAgentController()
	router.POST("/agent/heartbeat", agentController.Heartbeat)

	dmController := controller.NewDeviceMapperController()
	router.POST("/cluster-lvs/dm-conf", dmController.GeneratePreviewConf)
	//Test
	router.GET("/test", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"hello": "hello world!"})
	})
}

func newLogger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 开始时间
		startTime := time.Now()
		// body
		body, _ := ctx.GetRawData()
		ctx.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		// 处理请求
		ctx.Next()
		// 结束时间
		endTime := time.Now()
		// 执行时间
		latencyTime := endTime.Sub(startTime)
		// 请求方式
		reqMethod := ctx.Request.Method
		// 请求路由
		reqUri := ctx.Request.RequestURI
		// 状态码
		statusCode := ctx.Writer.Status()
		// 请求IP
		clientIP := ctx.ClientIP()
		// param
		params := ctx.Params

		//日志格式
		smslog.Debugf("| %3d | %13v | %15s | %s | %s | %v| %s |",
			statusCode,
			latencyTime,
			clientIP,
			reqMethod,
			reqUri,
			params,
			string(body),
		)
	}
}
