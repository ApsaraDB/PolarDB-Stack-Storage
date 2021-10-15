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


package repository

type IdempotentRepo struct {
	*BaseDB
}

func (i *IdempotentRepo) Create(clusterLv *Idempotent) (int64, error) {
	affected, err := i.Engine.Insert(clusterLv)
	if err != nil {
		return 0, err
	}
	return affected, nil
}

func (i *IdempotentRepo) FindById(name string) (*Idempotent, error) {
	model := Idempotent{
		IdempotentId: name,
	}

	exist, err := i.Engine.Alias("a").Where("a.Idempotent_id=?", name).Get(&model)
	if err != nil {
		return nil, err
	}

	if !exist {
		return nil, nil
	}

	return &model, nil
}

func (i *IdempotentRepo) FindBySourceAndId(idempotentId, source string) (*Idempotent, error) {
	model := Idempotent{
		IdempotentId: idempotentId,
	}

	exist, err := i.Engine.Alias("a").Where("a.idempotent_id=? and a.source=?", idempotentId, source).Get(&model)
	if err != nil {
		return nil, err
	}

	if !exist {
		return nil, nil
	}

	return &model, nil
}

func (i *IdempotentRepo) FindBySource(source string) (*Idempotent, error) {
	model := Idempotent{
		Source: source,
	}

	exist, err := i.Engine.Alias("a").Where("a.source=?", source).Get(&model)
	if err != nil {
		return nil, err
	}

	if !exist {
		return nil, nil
	}

	return &model, nil
}

func (i *IdempotentRepo) Save(model *Idempotent) (int64, error) {
	result, err := i.Engine.Where("idempotent_id=?", model.IdempotentId).Update(model)
	if err != nil {
		return 0, err
	}

	return result, nil
}

func (i *IdempotentRepo) Delete(id uint64) (int64, error) {
	affected, err := i.Engine.ID(id).Delete(&Idempotent{})
	if err != nil {
		return 0, err
	}
	return affected, nil
}

func NewIdempotentRepo() *IdempotentRepo {
	return &IdempotentRepo{
		BaseDB: GetBaseDB(),
	}
}
