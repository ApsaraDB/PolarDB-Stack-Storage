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


package service

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/ngaut/log"
)

const (
	VersionInvalid = int64(-1)
)

type MetaServer struct {
	DataDir string
}

func NewMetaServer(dir string) *MetaServer {
	return &MetaServer{DataDir: dir}
}

func (m *MetaServer) StageTable(name, table string) (*TableFile, error) {
	fullPath := path.Join(m.DataDir, name)
	if err := ioutil.WriteFile(fullPath, []byte(table), 0x700); err != nil {
		return nil, err
	}
	return NewTableFile(fullPath)
}

// 新的table文件默认没有版本号, 修改新的table文件版本号为原来版本号+1
func (m *MetaServer) CommitFile(file *TableFile) error {
	var (
		version = int64(0)
	)

	if file.version != VersionInvalid {
		return fmt.Errorf("commit file version must be %d", VersionInvalid)
	}

	files, err := m.LoadFiles()
	if err != nil {
		return err
	}

	if f2, ok := files[file.name]; ok {
		version = f2.version + 1
	}
	return os.Rename(file.path, path.Join(m.DataDir, fmt.Sprintf("%s.%d", file.name, version)))
}

func (m *MetaServer) LoadFiles() (map[string]*TableFile, error) {
	result := make(map[string]*TableFile)

	tableFiles, err := ioutil.ReadDir(m.DataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read data dir %s, err: %s", m.DataDir, err)
	}
	for _, info := range tableFiles {
		f1, err := NewTableFile(info.Name())
		if err != nil {
			log.Errorf("invalid table file %s", info)
			continue
		}

		if f2, ok := result[f1.name]; ok {
			if f1.Compare(f2) < 0 {
				result[f1.name] = f2
			}
		} else {
			result[f1.name] = f1
		}
	}
	return result, nil
}

type TableFile struct {
	path    string
	dir     string
	name    string
	version int64
}

func NewTableFile(filePath string) (*TableFile, error) {
	version := VersionInvalid
	dirname, basename := path.Split(filePath)
	parts := strings.SplitN(basename, ".", 2)
	if len(parts) == 2 {
		i, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return nil, err
		}
		version = i
	}
	return &TableFile{path: filePath, name: parts[0], version: version, dir: dirname}, nil
}

func (t *TableFile) Compare(o *TableFile) int {
	if t.version > o.version {
		return 1
	} else if t.version < o.version {
		return -1
	}
	return 0
}
