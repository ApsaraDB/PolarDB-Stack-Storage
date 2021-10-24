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

//兼容lun, 其实是lv类型的 multipath
type ClusterLunController struct {
	lvRepo             lv.LvRepository
	as                 assembler.ClusterLunAssembler
	lvForOldLunService *service.LvForOldLunService
}

// @Summary 创建 Cluster LUN
// @Tags LUN 管理
// @version 1.0
// @Description 用于创建 Cluster LUN
// @Accept  json
// @Produce  json
// @Param clusterLun body view.ClusterLunCreateRequest true "请求参数"
// @Success 200 object view.WorkflowIdResponse 成功后返回值
// @Failure 400 object view.ErrorResult 参数异常返回值
// @Failure 500 object view.ErrorResult 服务异常返回值
// @Router /cluster-luns [post]
func (controller *ClusterLunController) CreateClusterLun(ctx *gin.Context) {
	smslog.Info("call CreateClusterLun")

	var createRequest view.ClusterLunCreateRequest
	if err := ParseParam(ctx, &createRequest); err != nil {
		smslog.Errorf("Could not create cluster lun %v: %v", createRequest, err)
		ReturnError(ctx, err)
		return
	}

	workflowResp, err := controller.lvForOldLunService.Create(GetTraceContextFromHeader(ctx), &createRequest)
	if err != nil {
		smslog.Errorf("Could not create cluster lun %v: %v", createRequest, err)
		ReturnError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, workflowResp)
}

// @Summary 列表 Cluster LUN
// @Tags LUN 管理
// @version 1.0
// @Description 用于查询所有 Cluster LUN
// @Accept  json
// @Produce  json
// @Success 200 array view.ClusterLunResponse 成功后返回值
// @Failure 500 object view.ErrorResult 服务异常返回值
// @Router /cluster-luns [get]
func (controller *ClusterLunController) QueryClusterLuns(ctx *gin.Context) {
	smslog.Info("call QueryClusterLuns")

	responses, err := controller.lvForOldLunService.QueryAllMultipathVolumes()
	if err != nil {
		smslog.Errorf("Could not query cluster luns: %v", err)
		ReturnError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, responses)
}

// @Summary 查询 Cluster LUN
// @Tags LUN 管理
// @version 1.0
// @Description 用于查询指定 Cluster LUN
// @Accept  json
// @Produce  json
// @Param wwid query string true "LUN WWID"
// @Success 200 object view.ClusterLunResponse 成功后返回值
// @Failure 400 object view.ErrorResult 参数异常返回值
// @Failure 500 object view.ErrorResult 服务异常返回值
// @Router /cluster-luns/:wwid [get]
func (controller *ClusterLunController) SearchClusterLun(ctx *gin.Context) {
	smslog.Info("call SearchClusterLun")

	wwid, exist := ctx.Params.Get("wwid")
	if !exist {
		err := fmt.Errorf("request param not exist wwid")
		smslog.Errorf(err.Error())
		ReturnError(ctx, err)
		return
	}

	lvEntity, err := controller.lvRepo.FindByVolumeId(wwid)
	if err != nil {
		smslog.Errorf("Could not find cluster lun by wwid %s: %v", wwid, err)
		ReturnError(ctx, err)
		return
	}
	response := controller.as.ToClusterLunView(lvEntity, assembler.GetLvPvcMap())
	ctx.JSON(http.StatusOK, response)
}

// @Summary 格式化 Cluster LUN
// @Tags LUN 管理
// @version 1.0
// @Description 用于格式化指定 Cluster LUN
// @Accept  json
// @Produce  json
// @Param clusterLun body view.ClusterLunFormatRequest true "请求参数"
// @Success 200 object view.WorkflowIdResponse 成功后返回值
// @Failure 400 object view.ErrorResult 参数异常返回值
// @Failure 500 object view.ErrorResult 服务异常返回值
// @Router /cluster-luns/format [post]
func (controller *ClusterLunController) FormatClusterLun(ctx *gin.Context) {
	smslog.Info("call FormatClusterLun")

	var formatRequest view.ClusterLunFormatRequest
	if err := ParseParam(ctx, &formatRequest); err != nil {
		smslog.Errorf("Cloud not parse cluster lun format request %v: %v", formatRequest, err)
		ReturnError(ctx, err)
		return
	}

	wflResp, err := controller.lvForOldLunService.Format(GetTraceContextFromHeader(ctx), &formatRequest)
	if err != nil {
		smslog.Errorf("Could not format cluster lun %v: %v", formatRequest, err)
		ReturnError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, wflResp)
}

// @Summary Lun from same san
// @Tags LUN 管理
// @version 1.0
// @Description 用于查询Cluster LUN是否来自于同一个san存储
// @Accept  json
// @Produce  json
// @Param clusterLun body view.ClusterLunSameSanRequest true "请求参数"
// @Success 200 object view.BoolResult 成功后返回值
// @Failure 400 object view.ErrorResult 参数异常返回值
// @Failure 500 object view.ErrorResult 服务异常返回值
// @Router /cluster-luns/same-san [post]
func (controller *ClusterLunController) LunsFromSameSan(ctx *gin.Context) {
	smslog.Info("call LunsFromSameSan")

	var request view.ClusterLunSameSanRequest
	if err := ParseParam(ctx, &request); err != nil {
		smslog.Errorf("Cloud not parse cluster lun samesan request %v: %v", request, err)
		ReturnError(ctx, err)
		return
	}

	resp, err := controller.lvForOldLunService.LunsFromSameSan(&request)
	if err != nil {
		smslog.Errorf("Could not samesan cluster lun %v: %v", request, err)
		ReturnError(ctx, err)
		return
	}
	ReturnBool(ctx, resp)
}

// @Summary 扩容 Cluster LUN
// @Tags LUN 管理
// @version 1.0
// @Description 用于扩容指定 Cluster LUN
// @Accept  json
// @Produce  json
// @Param clusterLun body view.ClusterLunCreateRequest true "请求参数"
// @Success 200 object view.WorkflowIdResponse 成功后返回值
// @Failure 400 object view.ErrorResult 参数异常返回值
// @Failure 500 object view.ErrorResult 服务异常返回值
// @Router /cluster-luns/expand [post]
func (controller *ClusterLunController) ExpandClusterLun(ctx *gin.Context) {
	smslog.Info("call ExpandClusterLun")

	var expandRequest view.ClusterLunCreateRequest
	if err := ParseParam(ctx, &expandRequest); err != nil {
		smslog.Errorf("Cloud not parse cluster lun expand request %v: %v", expandRequest, err)
		ReturnError(ctx, err)
		return
	}

	wflResp, err := controller.lvForOldLunService.Expand(GetTraceContextFromHeader(ctx), &expandRequest)
	if err != nil {
		smslog.Errorf("Could not expand cluster lun %v: %v", expandRequest, err)
		ReturnError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, wflResp)
}

// @Summary 扩容 Cluster LUN For Filesystem
// @Tags LUN 管理
// @version 1.0
// @Description 用于扩容指定 Cluster LUN Filesystem
// @Accept  json
// @Produce  json
// @Param clusterLun body view.ClusterLunFsExpandRequest true "请求参数"
// @Success 200 object view.WorkflowIdResponse 成功后返回值
// @Failure 400 object view.ErrorResult 参数异常返回值
// @Failure 500 object view.ErrorResult 服务异常返回值
// @Router /cluster-luns/fs-expand [post]
func (controller *ClusterLunController) ExpandClusterLunForFs(ctx *gin.Context) {
	smslog.Info("call ExpandClusterLunForFs")

	var expandFsRequest view.ClusterLunFsExpandRequest
	if err := ParseParam(ctx, &expandFsRequest); err != nil {
		smslog.Errorf("Cloud not parse cluster lun expand fs request %v: %v", expandFsRequest, err)
		ReturnError(ctx, err)
		return
	}

	wflResp, err := controller.lvForOldLunService.FsExpand(GetTraceContextFromHeader(ctx), &expandFsRequest)
	if err != nil {
		smslog.Errorf("Could not expand cluster lun %v: %v", expandFsRequest, err)
		ReturnError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, wflResp)
}

// @Summary 删除 Cluster LUN
// @Tags LUN 管理
// @version 1.0
// @Description 用于删除指定 Cluster LUN
// @Accept  json
// @Produce  json
// @Param wwid query string true "LUN WWID"
// @Success 200 object string 成功后返回值
// @Failure 400 object view.ErrorResult 参数异常返回值
// @Failure 500 object view.ErrorResult 服务异常返回值
// @Router /cluster-luns/:wwid [delete]
func (controller *ClusterLunController) DeleteClusterLun(ctx *gin.Context) {
	smslog.Info("call DeleteClusterLun")

	wwid, exist := ctx.Params.Get("wwid")
	if !exist {
		err := fmt.Errorf("request param not exist wwid")
		smslog.Errorf(err.Error())
		ReturnError(ctx, err)
		return
	}

	//todo add workflow support
	if _, err := controller.lvRepo.DeleteByVolumeId(wwid); err != nil {
		smslog.Errorf("Could not delete cluster lun by wwid %s: %v", wwid, err)
		ReturnError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"tst": 0})
}

func NewClusterLunController() *ClusterLunController {
	clusterLunController := &ClusterLunController{
		lvRepo:             lv.GetLvRepository(),
		as:                 assembler.NewClusterLunAssembler(),
		lvForOldLunService: service.NewLvForOldLunService(),
	}
	return clusterLunController
}
