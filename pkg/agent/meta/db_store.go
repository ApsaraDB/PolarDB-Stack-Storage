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

package meta

import (
	"fmt"
	bolt "go.etcd.io/bbolt"
)

const (
	DbName  = "logical-volume.db"
	DataKey = "data"
)

type DBStore struct {
	*bolt.DB
}

//todo version control
func (s *DBStore) Put(record *DMTableRecord) error {
	return s.Update(func(tx *bolt.Tx) error {
		var err error
		bucket := tx.Bucket([]byte(record.Name))
		if bucket == nil {
			bucket, err = tx.CreateBucket([]byte(record.Name))
			if err != nil {
				return err
			}
		}
		err = bucket.Put([]byte(DataKey), []byte(record.Data))
		if err != nil {
			return err
		}
		return nil
	})

}

func (s *DBStore) Get(name string) (*DMTableRecord, error) {
	var ret *DMTableRecord
	err := s.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(name))
		if bucket == nil {
			return fmt.Errorf("can not find record for %s", name)
		}
		data := bucket.Get([]byte(DataKey))
		if data == nil {
			return fmt.Errorf("can not find record data for %s", name)
		}
		ret = &DMTableRecord{
			Name: name,
			Data: string(data),
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (s *DBStore) Delete(name string) error {
	return s.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(name))
		if bucket != nil {
			return tx.DeleteBucket([]byte(name))
		}
		return nil
	})
}

func (s *DBStore) List() (map[string]*DMTableRecord, error) {
	var ret = map[string]*DMTableRecord{}
	err := s.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, b *bolt.Bucket) error {
			data := b.Get([]byte(DataKey))
			if data == nil {
				return fmt.Errorf("can not find record data for %s", name)
			}
			ret[string(name)] = &DMTableRecord{
				Name: string(name),
				Data: string(data),
			}
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (s *DBStore) Versions(name string) ([]*DMTableRecord, error) {
	panic("implement me")
}

func NewDBStore(dataDir string) (*DBStore, error) {
	db, err := bolt.Open(dataDir+DbName, 0600, nil)
	if err != nil {
		return nil, err
	}
	return &DBStore{
		DB: db,
	}, nil
}
