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


package stage

import (
	"fmt"
	"polardb-sms/pkg/common"
	smslog "polardb-sms/pkg/log"
	"polardb-sms/pkg/manager/domain/k8spvc"
	"polardb-sms/pkg/manager/domain/lv"
	"reflect"
	"sync"
)

type DBPersistHandler struct {
	handlerMap map[string]reflect.Value
}

var _handler *DBPersistHandler
var _handlerOnce sync.Once

type DbMethod string

const (
	Create         DbMethod = "create"
	Update         DbMethod = "update"
	Delete         DbMethod = "delete"
	UpdateCapacity DbMethod = "update-capacity"
	UpdateStatus   DbMethod = "update-status"
	UpdatePr       DbMethod = "update-pr"
	UpdateUsed     DbMethod = "update-used"
)

const (
	TableLv        string = "lv"
	TablePv        string = "pv"
	TablePvc       string = "pvc"
	TablePvcStatus string = "pvc-status"
)

type DBPersistContext struct {
	Table    string   `json:"table"`
	DbMethod DbMethod `json:"db_method"`
	Param    string   `json:"param"`
}

//TODO fine tune this ugly
func GetDBPersistHandler() *DBPersistHandler {
	if _handler == nil {
		_handlerOnce.Do(func() {
			h := &DBPersistHandler{
				handlerMap: make(map[string]reflect.Value),
			}
			lvRepo := lv.GetLvRepository()
			h.handlerMap["lv-create"] = reflect.ValueOf(lvRepo.Create)
			h.handlerMap["lv-update"] = reflect.ValueOf(lvRepo.Save)
			h.handlerMap["lv-update-used"] = reflect.ValueOf(lvRepo.UpdateUsed)
			h.handlerMap["lv-update-pr"] = reflect.ValueOf(lvRepo.UpdatePr)
			h.handlerMap["lv-delete"] = reflect.ValueOf(lvRepo.Delete)

			pvcRepo := k8spvc.GetPvcRepository()
			h.handlerMap["pvc-update-pr"] = reflect.ValueOf(pvcRepo.UpdatePrKey)
			h.handlerMap["pvc-create"] = reflect.ValueOf(pvcRepo.Create)
			h.handlerMap["pvc-delete"] = reflect.ValueOf(pvcRepo.Delete)
			h.handlerMap["pvc-update-capacity"] = reflect.ValueOf(pvcRepo.UpdateCapacity)
			h.handlerMap["pvc-update-status"] = reflect.ValueOf(pvcRepo.UpdateStatus)

			pvRepo := k8spvc.GetPvRepository()
			h.handlerMap["pv-create"] = reflect.ValueOf(pvRepo.CreateOrUpdate)
			h.handlerMap["pv-delete"] = reflect.ValueOf(pvRepo.Delete)

			pvcStatusRepo := k8spvc.GetPvcStatusRepository()
			h.handlerMap["pvc-status-create"] = reflect.ValueOf(pvcStatusRepo.Create)

			_handler = h
		})
	}
	return _handler
}

//TODO fine tune this ugly
func newParam(table string) (interface{}, error) {
	switch table {
	case "lv":
		return &lv.LogicalVolumeEntity{}, nil
	case "pvc":
		return &k8spvc.PersistVolumeClaimEntity{}, nil
	case "pv":
		return &k8spvc.PersistVolumeEntity{}, nil
	default:
		return nil, fmt.Errorf("do not support table %s", table)
	}
}

//TODO refactor this ugly code
func (c *DBPersistHandler) Run(mCtx *DBPersistContext) *StageExecResult {
	param, err := newParam(mCtx.Table)
	if err != nil {
		smslog.Infof("newParam err %s", err.Error())
		return StageExecFail(err.Error())
	}
	err = common.BytesToStruct([]byte(mCtx.Param), param)
	if err != nil {
		smslog.Errorf("err when call Decode, %v", err)
		return StageExecFail(err.Error())
	}
	refValue, ok := c.handlerMap[mCtx.Table+"-"+string(mCtx.DbMethod)]
	if !ok {
		err := fmt.Errorf("can not find handler for %s", mCtx.Table)
		smslog.Infof(err.Error())
		return StageExecFail(err.Error())
	}

	callParam := reflect.ValueOf(param)
	ret := refValue.Call([]reflect.Value{callParam})
	if !ret[1].IsNil() {
		errRef := ret[1].Interface()
		err = errRef.(error)
		smslog.Infof("err when call %s, %v", mCtx.DbMethod, err)
		return StageExecFail(err.Error())
	}
	return StageExecSuccess(nil)
}

type DBPersistStageRunner struct {
	*Stage
}

func (s *DBPersistStageRunner) Run(ctx common.TraceContext) *StageExecResult {
	dbExecContext := s.Content.(*DBPersistContext)
	smslog.Debugf("DBPersistStageRunner: table[%s] method[%s]", dbExecContext.Table, dbExecContext.DbMethod)
	ret := GetDBPersistHandler().Run(dbExecContext)
	s.Result = ret
	return ret
}

func (s *DBPersistStageRunner) Rollback(ctx common.TraceContext) *StageExecResult {
	return StageExecFail(fmt.Errorf("umimplement fs expand rollback").Error())
}

func NewDBPersistStage(dbMethod DbMethod, tableName string, param interface{}) (*DBPersistStageRunner, error) {
	paramBytes, err := common.StructToBytes(param)
	if err != nil {
		return nil, err
	}
	return &DBPersistStageRunner{
		Stage: &Stage{
			Content: &DBPersistContext{
				DbMethod: dbMethod,
				Table:    tableName,
				Param:    string(paramBytes),
			},
			SType:     DBPersistStage,
			StartTime: 0,
			Result:    nil,
		},
	}, nil
}

func NewDBPersistPvcCreateStage(param interface{}) (*DBPersistStageRunner, error) {
	return NewDBPersistStage(Create, TablePvc, param)
}

func NewDBPersistPvcDeleteStage(param interface{}) (*DBPersistStageRunner, error) {
	return NewDBPersistStage(Delete, TablePvc, param)
}

func NewDBPersistPvcUpdatePrKeyStage(param interface{}) (*DBPersistStageRunner, error) {
	return NewDBPersistStage(UpdatePr, TablePvc, param)
}

func NewDBPersistPvcUpdateStatusStage(param interface{}) (*DBPersistStageRunner, error) {
	return NewDBPersistStage(UpdateStatus, TablePvc, param)
}

func NewDBPersistPvcUpdateCapacityStage(param interface{}) (*DBPersistStageRunner, error) {
	return NewDBPersistStage(UpdateCapacity, TablePvc, param)
}

func NewDBPvcStatusCreateStage(param interface{}) (*DBPersistStageRunner, error) {
	return NewDBPersistStage(Create, TablePvcStatus, param)
}

func NewDBPersistLvCreateStage(param interface{}) (*DBPersistStageRunner, error) {
	return NewDBPersistStage(Create, TableLv, param)
}

func NewDBPersistLvUpdateStage(param interface{}) (*DBPersistStageRunner, error) {
	return NewDBPersistStage(Update, TableLv, param)
}

func NewDBPersistLvUsedStage(param interface{}) (*DBPersistStageRunner, error) {
	return NewDBPersistStage(UpdateUsed, TableLv, param)
}

func NewDBPersistLvPrStage(param interface{}) (*DBPersistStageRunner, error) {
	return NewDBPersistStage(UpdatePr, TableLv, param)
}

func NewDBPersistLvDeleteStage(param interface{}) (*DBPersistStageRunner, error) {
	return NewDBPersistStage(Delete, TableLv, param)
}

func NewDBPersistK8sPvCreateStage(param interface{}) (*DBPersistStageRunner, error) {
	return NewDBPersistStage(Create, TablePv, param)
}

func NewDBPersistK8sPvDeleteStage(param interface{}) (*DBPersistStageRunner, error) {
	return NewDBPersistStage(Delete, TablePv, param)
}

type DBPersistStageConstructor struct {
}

func (c *DBPersistStageConstructor) Construct() interface{} {
	return &DBPersistStageRunner{
		Stage: &Stage{
			Content: &DBPersistContext{},
		},
	}
}
