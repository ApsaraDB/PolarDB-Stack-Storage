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


package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	smslog "polardb-sms/pkg/log"
	"polardb-sms/pkg/manager/application/service"
	"polardb-sms/pkg/manager/application/view"
	"polardb-sms/pkg/manager/config"
	"polardb-sms/pkg/manager/domain/k8spvc"
)

//k8s 使用静态pvc， 整个管理流程offload到 存储管理组件里面完成
type PvcController struct {
	pvcService *service.PvcService
}

// @Summary 创建 PVC
// @Tags PVC 接入管理
// @version 1.0
// @Description 用于创建 PVC
// @Accept  json
// @Produce  json
// @Param pvc body view.PvcCreateWithVolumeRequest true "请求参数"
// @Success 200 object view.WorkflowIdResponse 成功后返回值
// @Failure 400 object view.ErrorResult 参数异常返回值
// @Failure 500 object view.ErrorResult 服务异常返回值
// @Router /pvcs [post]
func (c *PvcController) CreatePvcWithVolume(ctx *gin.Context) {
	smslog.Infof("call CreatePvcWithVolume")

	var pvcCreateRequest = view.PvcCreateWithVolumeRequest{}
	if err := ParseParam(ctx, &pvcCreateRequest); err != nil {
		return
	}
	workflowResp, err := c.pvcService.CreatePvcWithVolume(GetTraceContextFromHeader(ctx), &pvcCreateRequest)
	if err != nil {
		smslog.Errorf("Could not create pvc %v: %v", pvcCreateRequest, err)
		ReturnError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, workflowResp)
}

// @Summary 使用 PVC
// @Tags PVC 接入管理
// @version 1.0
// @Description 创建集群的时候，指定使用 PVC, 默认需要Bind Pvc和已有的磁盘，更新pvc和volume的状态
// @Accept  json
// @Produce  json
// @Param pvc body view.PvcBindVolumeRequest true "请求参数"
// @Success 200 object view.WorkflowIdResponse 成功后返回值
// @Failure 400 object view.ErrorResult 参数异常返回值
// @Failure 500 object view.ErrorResult 服务异常返回值
// @Router /pvcs/use [post]
func (c *PvcController) UsePvc(ctx *gin.Context) {
	smslog.Infof("call UsePvc")
	var pvcBindVolumeRequest view.PvcBindVolumeRequest
	if err := ParseParam(ctx, &pvcBindVolumeRequest); err != nil {
		return
	}
	workflowIdResp, err := c.pvcService.UseAndFormatPvc(GetTraceContextFromHeader(ctx), &pvcBindVolumeRequest)
	if err != nil {
		smslog.Errorf("Could not ues pvc %v: %v", pvcBindVolumeRequest, err)
		ReturnError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, workflowIdResp)
}

// @Summary 删除 PVC
// @Tags PVC 接入管理
// @version 1.0
// @Description 用于删除 PVC
// @Accept  json
// @Produce  json
// @Param name query string true "pvc name"
// @Param namespace query string true "pvc namespace"
// @Success 200 object view.WorkflowIdResponse 成功后返回值
// @Failure 400 object view.ErrorResult 参数异常返回值
// @Failure 500 object view.ErrorResult 服务异常返回值
// @Router /pvcs/:name [delete]
func (c *PvcController) DeletePvc(ctx *gin.Context) {
	smslog.Debugf("call DeletePvc")
	pvcName := ctx.Params.ByName("name")
	namespace, ok := ctx.GetQuery("namespace")
	if !ok || pvcName == "" {
		smslog.Errorf("Could not find pvc or pvcName  namespace in params %v", ctx.Params)
		ReturnError(ctx, fmt.Errorf("miss pvc namespace in query param"))
		return
	}

	workflowResp, err := c.pvcService.DeletePvc(GetTraceContextFromHeader(ctx),
		&view.PvcRequest{
			Name:      pvcName,
			Namespace: namespace,
		})
	if err != nil {
		smslog.Errorf("Could not delete pvc %s: %v", pvcName, err)
		ReturnError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, workflowResp)
}

// @Summary 释放 PVC
// @Tags PVC 接入管理
// @version 1.0
// @Description 用于释放 PVC
// @Accept  json
// @Produce  json
// @Param pvc body view.PvcRequest true "请求参数"
// @Success 200 object view.WorkflowIdResponse 成功后返回值
// @Failure 400 object view.ErrorResult 参数异常返回值
// @Failure 500 object view.ErrorResult 服务异常返回值
// @Router /pvcs/release [post]
func (c *PvcController) ReleasePvc(ctx *gin.Context) {
	smslog.Infof("call ReleasePvc")
	var pvcRequest view.PvcRequest
	if err := ParseParam(ctx, &pvcRequest); err != nil {
		return
	}
	workflowResp, err := c.pvcService.ReleasePvc(GetTraceContextFromHeader(ctx),
		&pvcRequest)
	if err != nil {
		smslog.Errorf("Could not release pvc %v: %v", pvcRequest, err)
		ReturnError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, workflowResp)
}

// @Summary 查询 PVC是否ready
// @Tags PVC 接入管理
// @version 1.0
// @Description 用于创建集群时，查询PVC是否准备完，可以使用
// @Accept  json
// @Produce  json
// @Param name query string true "pvc name"
// @Param workflowId query string true "workflowId"
// @Success 200 object view.PvcIsReadyResponse 成功后返回值
// @Failure 400 object view.ErrorResult 参数异常返回值
// @Failure 500 object view.ErrorResult 服务异常返回值
// @Router /pvcs/ready [get]
func (c *PvcController) PvcIsReady(ctx *gin.Context) {
	smslog.Infof("call PvcIsReady")
	var pvcRequest view.PvcRequest
	pvcName := ctx.Query("name")
	if pvcName == "" {
		err := fmt.Errorf("request param not exist name")
		smslog.Errorf(err.Error())
		ReturnError(ctx, err)
		return
	}
	namespace, exist := ctx.GetQuery("namespace")
	if !exist {
		err := fmt.Errorf("request param not exist namespace")
		smslog.Errorf(err.Error())
		ReturnError(ctx, err)
		return
	}
	workflowId, exist := ctx.GetQuery("workflowId")
	if !exist {
		err := fmt.Errorf("request param not exist workflowId")
		smslog.Errorf(err.Error())
		ReturnError(ctx, err)
		return
	}
	isReady, err := c.pvcService.PvcIsReady(pvcName, namespace, workflowId)
	if err != nil {
		smslog.Errorf("Could not check pvc status %v: %v", pvcRequest, err)
		ReturnError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, isReady)
}

// @Summary PVC 添加写锁节点
// @Tags PVC 接入管理
// @version 1.0
// @Description 用于PVC 添加写锁节点, 集群中只有一个节点可以添加写锁
// @Accept  json
// @Produce  json
// @Param pvc body view.PvcWriteLockRequest true "请求参数"
// @Success 200 object view.WorkflowIdResponse 成功后返回值
// @Failure 400 object view.ErrorResult 参数异常返回值
// @Failure 500 object view.ErrorResult 服务异常返回值
// @Router /pvcs/lock [post]
func (c *PvcController) PvcSetWriteLock(ctx *gin.Context) {
	smslog.Infof("call PvcSetWriteLock")
	var pvcLockRequest view.PvcWriteLockRequest
	if err := ParseParam(ctx, &pvcLockRequest); err != nil {
		return
	}
	workflowResp, err := c.pvcService.SetVolumeWriteLock(GetTraceContextFromHeader(ctx),
		&pvcLockRequest)
	if err != nil {
		smslog.Errorf("Could not set pvc write lock %v: %v", pvcLockRequest, err)
		ReturnError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, workflowResp)
}

// @Summary PVC 强制format和lock
// @Tags PVC 接入管理
// @version 1.0
// @Description 用于PVC 强制format和lock, 添加写锁节点, 集群中只有一个节点可以添加写锁
// @Accept  json
// @Produce  json
// @Param pvc body view.PvcWriteLockRequest true "请求参数"
// @Success 200 object view.WorkflowIdResponse 成功后返回值
// @Failure 400 object view.ErrorResult 参数异常返回值
// @Failure 500 object view.ErrorResult 服务异常返回值
// @Router /pvcs/formatAndLock [post]
func (c *PvcController) FormatAndLock(ctx *gin.Context) {
	smslog.Infof("call FormatAndLock with params %v", ctx.Params)
	var pvcLockRequest view.PvcWriteLockRequest
	if err := ParseParam(ctx, &pvcLockRequest); err != nil {
		return
	}
	workflowResp, err := c.pvcService.ForceFormatAndLock(GetTraceContextFromHeader(ctx),
		&pvcLockRequest)
	if err != nil {
		smslog.Errorf("Could not set pvc write lock %v: %v", pvcLockRequest, err)
		ReturnError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, workflowResp)
}

// @Summary PVC format
// @Tags PVC 接入管理
// @version 1.0
// @Description 用于PVC format
// @Accept  json
// @Produce  json
// @Param pvc body view.PvcFormatRequest true "请求参数"
// @Success 200 object view.WorkflowIdResponse 成功后返回值
// @Failure 400 object view.ErrorResult 参数异常返回值
// @Failure 500 object view.ErrorResult 服务异常返回值
// @Router /pvcs/format [post]
func (c *PvcController) Format(ctx *gin.Context) {
	smslog.Infof("call FormatAndLock with params %v", ctx.Params)
	var pvcFormatRequest view.PvcFormatRequest
	if err := ParseParam(ctx, &pvcFormatRequest); err != nil {
		return
	}
	workflowResp, err := c.pvcService.PvcFsFormat(GetTraceContextFromHeader(ctx),
		&pvcFormatRequest)
	if err != nil {
		smslog.Errorf("Could not format pvc %v: %v", pvcFormatRequest, err)
		ReturnError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, workflowResp)
}

// @Summary 查询 PVC后端的磁盘读写权限
// @Tags PVC 接入管理
// @version 1.0
// @Description 用于集群校验读写拓扑，查询磁盘在集群各个节点的读写权限
// @Accept  json
// @Produce  json
// @Param name query string true "pvc name"
// @Param workflowId query string true "workflowId"
// @Param namespace query string true "pvc namespace"
// @Success 200 object view.PvcVolumePermissionTopoResponse 成功后返回值
// @Failure 400 object view.ErrorResult 参数异常返回值
// @Failure 500 object view.ErrorResult 服务异常返回值
// @Router /pvcs/topo [get]
func (c *PvcController) QueryVolumePermissionTopo(ctx *gin.Context) {
	smslog.Infof("call QueryVolumePermissionTopo")
	pvcName := ctx.Query("name")
	if pvcName == "" {
		err := fmt.Errorf("request param not exist name")
		smslog.Errorf(err.Error())
		ReturnError(ctx, err)
		return
	}
	workflowId, exist := ctx.GetQuery("workflowId")
	if !exist {
		err := fmt.Errorf("request param not exist workflowId")
		smslog.Errorf(err.Error())
		ReturnError(ctx, err)
		return
	}
	namespace, exist := ctx.GetQuery("namespace")
	if !exist {
		err := fmt.Errorf("request param not exist namespace")
		smslog.Errorf(err.Error())
		ReturnError(ctx, err)
		return
	}

	retView, err := c.pvcService.QueryVolumePermissionTopo(pvcName, namespace, workflowId)
	if err != nil {
		smslog.Errorf("Could not query volume permission: %v", err)
		ReturnError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, retView)
}

// @Summary 查询pvc
// @Tags PVC 接入管理
// @version 1.0
// @Description 用于查询指定的pvc
// @Accept  json
// @Produce  json
// @Param name query string true "pvc name"
// @Param namespace query string true "pvc namespace"
// @Success 200 object view.PvcResponse 成功后返回值
// @Failure 400 object view.ErrorResult 参数异常返回值
// @Failure 500 object view.ErrorResult 服务异常返回值
// @Router /pvcs/:name [get]
func (c *PvcController) QueryPvc(ctx *gin.Context) {
	smslog.Infof("call QueryPvc request: %+v, params: %v", ctx.Params, ctx.Request)

	pvcName, exist := ctx.Params.Get("pvcDynamicName")
	if !exist {
		err := fmt.Errorf("request param not exist name")
		smslog.Errorf(err.Error())
		ReturnError(ctx, err)
		return
	}

	namespace, exist := ctx.GetQuery("namespace")
	if !exist {
		err := fmt.Errorf("request param not exist namespace")
		smslog.Errorf(err.Error())
		ReturnError(ctx, err)
		return
	}
	retView, err := c.pvcService.QueryPvc(pvcName, namespace)
	if err != nil {
		smslog.Errorf("Could not query pvc %s/%s: %v", pvcName, namespace, err)
		ReturnError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, retView)
}

// @Summary 查询pvc 列表
// @Tags PVC 接入管理
// @version 1.0
// @Description 用于查询指定pvc列表
// @Accept  json
// @Produce  json
// @Param volumeType query string true "pvc used volumeType"
// @Success 200 {array} view.PvcResponse  成功后返回值
// @Failure 400 object view.ErrorResult 参数异常返回值
// @Failure 500 object view.ErrorResult 服务异常返回值
// @Router /pvcs [get]
func (c *PvcController) QueryPvcs(ctx *gin.Context) {
	smslog.Infof("call QueryPvcs")
	volumeTypeStr, exist := ctx.GetQuery("volumeType")
	if !exist {
		volumeTypeStr = ""
	}
	smslog.Infof("Volume type %s", volumeTypeStr)
	retView, err := c.pvcService.QueryPvcsByType(volumeTypeStr)
	if err != nil {
		smslog.Errorf("Could not query pvcs by type %s: %v", volumeTypeStr, err)
		ReturnError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, retView)
}

// @Summary PVC扩容
// @Tags PVC 接入管理
// @version 1.0
// @Description  PVC对应的数据库集群扩容
// @Accept  json
// @Produce  json
// @Param pvc body view.PvcExpandFsRequest true "请求参数"
// @Success 200 object view.WorkflowIdResponse 成功后返回值
// @Failure 400 object view.ErrorResult 参数异常返回值
// @Failure 500 object view.ErrorResult 服务异常返回值
// @Router /pvcs/expand [post]
func (c *PvcController) PvcExpandFs(ctx *gin.Context) {
	smslog.Infof("call PvcExpandFs")

	var pvcExpandRequest = view.PvcExpandFsRequest{}
	if err := ParseParam(ctx, &pvcExpandRequest); err != nil {
		return
	}
	workflowResp, err := c.pvcService.PvcExpandFs(GetTraceContextFromHeader(ctx), &pvcExpandRequest)
	if err != nil {
		smslog.Errorf("Could not expand pvc %v: %v", pvcExpandRequest, err)
		ReturnError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, workflowResp)
}

// @Summary PVC扩容查询
// @Tags PVC 接入管理
// @version 1.0
// @Description  PVC对应的数据库集群扩容是否结束
// @Accept  json
// @Produce  json
// @Param workflowId query string true "workflowId"
// @Success 200 object view.PvcExpandFsResponse 成功后返回值
// @Failure 400 object view.ErrorResult 参数异常返回值
// @Failure 500 object view.ErrorResult 服务异常返回值
// @Router /pvcs/expand [get]
func (c *PvcController) PvcExpandedFs(ctx *gin.Context) {
	smslog.Infof("call PvcExpandFs")

	var pvcExpandRequest = view.PvcExpandFsRequest{}
	if err := ParseParam(ctx, &pvcExpandRequest); err != nil {
		return
	}

	workflowResp, err := c.pvcService.PvcExpandFs(GetTraceContextFromHeader(ctx), &pvcExpandRequest)
	if err != nil {
		smslog.Errorf("Could not expand pvc %v: %v", pvcExpandRequest, err)
		ReturnError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, workflowResp)
}

func (c *PvcController) ModifyPvc(ctx *gin.Context) {
	hackToken, exist := ctx.GetQuery("hack")
	if !exist || hackToken != "sg-private" {
		ReturnError(ctx, fmt.Errorf("wrong hack token"))
	}
	pvcName, exist := ctx.GetQuery("pvc")
	if !exist {
		ReturnError(ctx, fmt.Errorf("miss pvc name"))
	}
	pvcNamespace, exist := ctx.GetQuery("namespace")
	if !exist {
		pvcNamespace = "default"
	}
	modify, exist := ctx.GetQuery("modify")
	if !exist {
		ReturnError(ctx, fmt.Errorf("miss modify"))
	}

	value, exist := ctx.GetQuery("value")
	if !exist {
		ReturnError(ctx, fmt.Errorf("miss modify"))
	}
	switch modify {
	case "cap":
		_ = k8spvc.UpdatePvcCapacity(config.ClientSet, pvcName, pvcNamespace, value)
	case "status":
		_ = k8spvc.UpdatePvcStatus(config.ClientSet, pvcName, pvcNamespace, value)
	}
	ctx.JSON(http.StatusOK, "successful")
}

func NewPvcController() *PvcController {
	return &PvcController{
		pvcService: service.NewPvcService(),
	}
}
