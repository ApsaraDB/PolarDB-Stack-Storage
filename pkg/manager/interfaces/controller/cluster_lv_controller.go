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
	"polardb-sms/pkg/manager/application/assembler"
	"polardb-sms/pkg/manager/application/service"
	"polardb-sms/pkg/manager/application/view"
	"polardb-sms/pkg/manager/domain/lv"
)

type ClusterLvController struct {
	lvRepo lv.LvRepository
	as     assembler.ClusterLvAssembler
	cs     *service.ClusterLvService
}

// @Summary 创建 Cluster LV
// @Tags LV 管理
// @version 1.0
// @Description 用于创建 Cluster LV
// @Accept  json
// @Produce  json
// @Param clusterLv body view.ClusterLvCreateRequest true "请求参数"
// @Success 200 object view.WorkflowIdResponse 成功后返回值
// @Failure 400 object view.ErrorResult 参数异常返回值
// @Failure 500 object view.ErrorResult 服务异常返回值
// @Router /cluster-lvs [post]
func (controller *ClusterLvController) CreateClusterLv(ctx *gin.Context) {
	smslog.Infof("call CreateClusterLv")

	var createRequest view.ClusterLvCreateRequest
	if err := ParseParam(ctx, &createRequest); err != nil {
		smslog.Errorf("Could not create cluster lv %v: %v", createRequest, err)
		ReturnError(ctx, err)
		return
	}
	wflResp, err := controller.cs.Create(GetTraceContextFromHeader(ctx), &createRequest)
	if err != nil {
		smslog.Errorf("Could not create cluster lv: %v", err)
		ReturnError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, wflResp)
}

// @Summary 列表 Cluster LV
// @Tags LV 管理
// @version 1.0
// @Description 用于列表 Cluster LV
// @Accept  json
// @Produce  json
// @Success 200 array view.ClusterLvResponse 成功后返回值
// @Failure 500 object view.ErrorResult 服务异常返回值
// @Router /cluster-lvs [get]
func (controller *ClusterLvController) QueryClusterLvs(ctx *gin.Context) {
	smslog.Infof("call QueryClusterLvs")

	responses, err := controller.cs.QueryAllDmVolumes()
	if err != nil {
		smslog.Errorf("Could not query cluster lvs: %v", err)
		ReturnError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, responses)
}

// @Summary 查询 Cluster LV
// @Tags LV 管理
// @version 1.0
// @Description 用于查询 Cluster LV
// @Accept  json
// @Produce  json
// @Param name query string true "LV name"
// @Success 200 object view.ClusterLvResponse 成功后返回值
// @Failure 400 object view.ErrorResult 参数异常返回值
// @Failure 500 object view.ErrorResult 服务异常返回值
// @Router /cluster-lvs/:name [get]
func (controller *ClusterLvController) SearchClusterLv(ctx *gin.Context) {
	smslog.Infof("call SearchClusterLv")
	name, exist := ctx.Params.Get("name")
	if !exist {
		err := fmt.Errorf("request param not exist lv name")
		smslog.Errorf(err.Error())
		ReturnError(ctx, err)
		return
	}

	lvEntity, err := controller.lvRepo.FindByVolumeId(name)
	if err != nil {
		smslog.Errorf("Could not search cluster lv: %v", err)
		ReturnError(ctx, err)
		return
	}
	response := controller.as.ToClusterLvView(lvEntity, assembler.GetLvPvcMap())
	ctx.JSON(http.StatusOK, response)
}

// @Summary 格式化 Cluster LV
// @Tags LV 管理
// @version 1.0
// @Description 用于格式化指定 Cluster LV
// @Accept  json
// @Produce  json
// @Param clusterLv body view.ClusterLvFormatRequest true "请求参数"
// @Success 200 object view.WorkflowIdResponse 成功后返回值
// @Failure 400 object view.ErrorResult 参数异常返回值
// @Failure 500 object view.ErrorResult 服务异常返回值
// @Router /cluster-lvs/format [post]
func (controller *ClusterLvController) FormatClusterLv(ctx *gin.Context) {
	smslog.Info("call FormatClusterLv")
	var request view.ClusterLvFormatRequest
	if err := ParseParam(ctx, &request); err != nil {
		smslog.Errorf("Cloud not parse cluster lun format request %v: %v", request, err)
		ReturnError(ctx, err)
		return
	}
	wflResp, err := controller.cs.Format(GetTraceContextFromHeader(ctx), &request)
	if err != nil {
		smslog.Errorf("Could not format cluster lv %v: %v", request, err)
		ReturnError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, wflResp)
}

// @Summary 扩容 Cluster LV
// @Tags LV 管理
// @version 1.0
// @Description 用于扩容指定 Cluster LV
// @Accept  json
// @Produce  json
// @Param clusterLv body view.ClusterLvCreateRequest true "请求参数"
// @Success 200 object view.WorkflowIdResponse 成功后返回值
// @Failure 400 object view.ErrorResult 参数异常返回值
// @Failure 500 object view.ErrorResult 服务异常返回值
// @Router /cluster-lvs/expand [post]
func (controller *ClusterLvController) ExpandClusterLv(ctx *gin.Context) {
	smslog.Info("call ExpandClusterLv")
	var expandRequest view.ClusterLvExpandRequest
	if err := ParseParam(ctx, &expandRequest); err != nil {
		smslog.Errorf("Cloud not parse cluster lun format request %v: %v", expandRequest, err)
		ReturnError(ctx, err)
		return
	}
	wflResp, err := controller.cs.Expand(GetTraceContextFromHeader(ctx), &expandRequest)
	if err != nil {
		smslog.Errorf("Could not expand cluster lv %v: %v", expandRequest, err)
		ReturnError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, wflResp)
}

// @Summary 扩容 Cluster LV For Filesystem
// @Tags LV 管理
// @version 1.0
// @Description 用于扩容指定 Cluster LV Filesystem
// @Accept  json
// @Produce  json
// @Param clusterLv body view.ClusterLvFsExpandRequest true "请求参数"
// @Success 200 object view.WorkflowIdResponse 成功后返回值
// @Failure 400 object view.ErrorResult 参数异常返回值
// @Failure 500 object view.ErrorResult 服务异常返回值
// @Router /cluster-lvs/fs-expand [post]
func (controller *ClusterLvController) ExpandClusterLvForFs(ctx *gin.Context) {
	smslog.Info("call ExpandClusterLv")
	var expandRequest view.ClusterLvFsExpandRequest

	if err := ParseParam(ctx, &expandRequest); err != nil {
		smslog.Errorf("Cloud not parse cluster lun expand request %v: %v", expandRequest, err)
		ReturnError(ctx, err)
		return
	}

	wflResp, err := controller.cs.FsExpand(GetTraceContextFromHeader(ctx), &expandRequest)
	if err != nil {
		smslog.Errorf("Could not do fs expand cluster lv %v: %v", expandRequest, err)
		ReturnError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, wflResp)
}

// @Summary 删除 Cluster LV
// @Tags LV 管理
// @version 1.0
// @Description 用于删除指定 Cluster LV
// @Accept  json
// @Produce  json
// @Param name query string true "lv name"
// @Success 200 object string 成功后返回值
// @Failure 400 object view.ErrorResult 参数异常返回值
// @Failure 500 object view.ErrorResult 服务异常返回值
// @Router /cluster-lvs/:name [delete]
func (controller *ClusterLvController) DeleteClusterLv(ctx *gin.Context) {
	smslog.Info("call DeleteClusterLv")
	volumeId, exist := ctx.Params.Get("name")
	if !exist {
		err := fmt.Errorf("request param not exist lv name")
		smslog.Errorf(err.Error())
		ReturnError(ctx, err)
		return
	}
	//todo fix this and add a workflow do this
	if _, err := controller.cs.Delete(GetTraceContextFromHeader(ctx), volumeId); err != nil {
		smslog.Errorf("Could not delete cluster lv by  volume Id %s: %v", volumeId, err)
		ReturnError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"tst": 0})
}

func NewClusterLvController() *ClusterLvController {
	clusterLvController := &ClusterLvController{
		lvRepo: lv.GetLvRepository(),
		as:     assembler.NewClusterLvAssembler(),
		cs:     service.NewClusterLvService(),
	}
	return clusterLvController
}
