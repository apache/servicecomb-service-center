/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package local

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/apache/servicecomb-service-center/datasource"
	_ "github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	etcdsync "github.com/apache/servicecomb-service-center/datasource/etcd/sync"
	"github.com/apache/servicecomb-service-center/datasource/schema"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/openlog"
	"github.com/little-cui/etcdadpt"
	"io/fs"
	"os"
	pathutil "path"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
)

var mutexMap = make(map[string]*sync.Mutex)
var mutexLock = &sync.Mutex{}

func init() {
	schema.Install("local_with_embeded_etcd", NewSchemaDAO)
	schema.Install("local_with_embedded_etcd", NewSchemaDAO)
}

func NewSchemaDAO(opts schema.Options) (schema.DAO, error) {
	return &SchemaDAO{}, nil
}

type SchemaDAO struct{}

func ExistDir(path string) error {
	_, err := os.ReadDir(path)
	if err != nil {
		// create the dir if not exist
		if os.IsNotExist(err) {
			err = os.MkdirAll(path, fs.ModePerm)
			if err != nil {
				log.Error(fmt.Sprintf("failed to makr dir %s ", path), err)
				return err
			}
			return nil
		}
		if err != nil {
			log.Error(fmt.Sprintf("failed to read dir %s ", path), err)
		}
	}
	return err
}

func MoveDir(srcDir string, dstDir string) (err error) {
	var movedFiles []string
	files, err := os.ReadDir(srcDir)
	if err != nil {
		log.Error("move schema files failed ", err)
	}
	for _, file := range files {
		err = ExistDir(dstDir)
		if err != nil {
			return err
		}
		srcFile := filepath.Join(srcDir, file.Name())
		dstFile := filepath.Join(dstDir, file.Name())
		err = os.Rename(srcFile, dstFile)
		if err != nil {
			log.Error("move schema files failed ", err)
			break
		}
		movedFiles = append(movedFiles, file.Name())
	}

	if err != nil {
		log.Error("Occur error when move schema files, begain rollback... ", err)
		for _, fileName := range movedFiles {
			srcFile := filepath.Join(srcDir, fileName)
			dstFile := filepath.Join(dstDir, fileName)
			err = os.Rename(dstFile, srcFile)
			if err != nil {
				log.Error("Occur error when move schema rollback... ", err)
			}
		}
	}
	return err
}

func createOrUpdateFile(filepath string, content []byte, rollbackOperations *[]FileDoRecord) error {
	err := ExistDir(pathutil.Dir(filepath))
	if err != nil {
		log.Error(fmt.Sprintf("failed to build new schema file dir %s", filepath), err)
		return err
	}

	var fileExist = true
	_, err = os.Stat(filepath)
	if err != nil {
		fileExist = false
	}

	if fileExist {
		oldcontent, err := ReadFile(filepath)
		if err != nil {
			log.Error(fmt.Sprintf("failed to read content to file %s ", filepath), err)
			return err
		}
		*rollbackOperations = append(*rollbackOperations, FileDoRecord{filepath: filepath, content: oldcontent})
	} else {
		*rollbackOperations = append(*rollbackOperations, FileDoRecord{filepath: filepath, content: nil})
	}

	err = os.WriteFile(filepath, content, 0666)
	if err != nil {
		log.Error(fmt.Sprintf("failed to create file %s", filepath), err)
		return err
	}

	return nil
}

func deleteFile(filepath string, rollbackOperations *[]FileDoRecord) error {
	_, err := os.Stat(filepath)
	if err != nil {
		log.Error(fmt.Sprintf("file does not exist when deleting file %s ", filepath), err)
		return nil
	}

	oldcontent, err := ReadFile(filepath)
	if err != nil {
		log.Error(fmt.Sprintf("failed to read content to file %s ", filepath), err)
		return err
	}

	*rollbackOperations = append(*rollbackOperations, FileDoRecord{filepath: filepath, content: oldcontent})

	err = os.Remove(filepath)
	if err != nil {
		log.Error(fmt.Sprintf("failed to delete file %s ", filepath), err)
		return err
	}

	return nil
}

func CleanDir(dir string) error {
	rollbackOperations := []FileDoRecord{}
	_, err := os.Stat(dir)
	if err != nil {
		return nil
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		filepath := filepath.Join(dir, file.Name())
		err = deleteFile(filepath, &rollbackOperations)
		if err != nil {
			break
		}
	}

	if err != nil {
		log.Error("Occur error when create schema files, begain rollback... ", err)
		rollback(rollbackOperations)
		return err
	}

	err = os.Remove(dir)
	if err != nil {
		log.Error("OOccur error when remove service schema dir, begain rollback... ", err)
		rollback(rollbackOperations)
		return err
	}

	return nil
}

func ReadFile(filepath string) ([]byte, error) {
	// check the file is empty
	content, err := os.ReadFile(filepath)
	if err != nil {
		log.Error(fmt.Sprintf("failed to read content to file %s ", filepath), err)
		return nil, err
	}
	return content, nil
}

func ReadAllFiles(dir string) ([]string, [][]byte, error) {
	files := []string{}
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		files = append(files, path)
		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	var contentArray [][]byte
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			log.Error(fmt.Sprintf("failed to read content from schema file %s ", file), err)
			return nil, nil, err
		}
		contentArray = append(contentArray, content)
	}
	return files, contentArray, nil
}

func rollback(rollbackOperations []FileDoRecord) {
	var err error
	for _, fileOperation := range rollbackOperations {
		if fileOperation.content == nil {
			err = deleteFile(fileOperation.filepath, &[]FileDoRecord{})
		} else {
			err = createOrUpdateFile(fileOperation.filepath, fileOperation.content, &[]FileDoRecord{})
		}
		if err != nil {
			log.Error("Occur error when rolling back schema files:  ", err)
		}
	}
}

type FileDoRecord struct {
	filepath string
	content  []byte
}

func (s *SchemaDAO) GetRef(ctx context.Context, refRequest *schema.RefRequest) (*schema.Ref, error) {
	domainProject := util.ParseDomainProject(ctx)
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	serviceID := refRequest.ServiceID
	schemaID := refRequest.SchemaID

	servicepath := filepath.Join(schema.RootFilePath, domainProject, serviceID, schemaID+".json")

	// read file content
	content, err := ReadFile(servicepath)
	if err != nil {
		log.Error(fmt.Sprintf("read service[%s] schema content file [%s] failed ", serviceID, schemaID), err)
		return nil, err
	}

	var schemaContent schema.ContentItem
	err = json.Unmarshal(content, &schemaContent)

	if err != nil {
		log.Error(fmt.Sprintf("get service[%s] schema content file [%s] failed when unmarshal", serviceID, schemaID), err)
		return nil, err
	}

	return &schema.Ref{
		Domain:    domain,
		Project:   project,
		ServiceID: serviceID,
		SchemaID:  schemaID,
		Hash:      schemaContent.Hash,
		Summary:   schemaContent.Summary,
		Content:   schemaContent.Content,
	}, nil
}

func (s *SchemaDAO) ListRef(ctx context.Context, refRequest *schema.RefRequest) ([]*schema.Ref, error) {
	domainProject := util.ParseDomainProject(ctx)
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	serviceID := refRequest.ServiceID

	var dir = filepath.Join(schema.RootFilePath, domainProject, serviceID)
	schemaIDs, contents, err := ReadAllFiles(dir)

	if err != nil {
		log.Error(fmt.Sprintf("read service[%s] schema content files failed ", serviceID), err)
		return nil, err
	}

	schemas := make([]*schema.Ref, 0, len(contents))
	for i := 0; i < len(contents); i++ {
		content := contents[i]
		var schemaContent schema.ContentItem
		err = json.Unmarshal(content, &schemaContent)
		if err != nil {
			log.Error(fmt.Sprintf("failed to unmarshal schema content for service [%s] and schema [%s]", serviceID, schemaIDs[i]), err)
			return nil, err
		}

		schemaFileName := schemaIDs[i]
		baseName := filepath.Base(schemaFileName)
		extension := filepath.Ext(baseName)
		schemaID := strings.TrimSuffix(baseName, extension)

		schemas = append(schemas, &schema.Ref{
			Domain:    domain,
			Project:   project,
			ServiceID: serviceID,
			SchemaID:  schemaID,
			Hash:      schemaContent.Summary,
			Summary:   schemaContent.Summary, // may be empty
			Content:   schemaContent.Content,
		})
	}
	return schemas, nil
}

func removeStringFromSlice(slice []string, s string) []string {
	for i := 0; i < len(slice); i++ {
		if slice[i] == s {
			slice = append(slice[:i], slice[i+1:]...)
			i--
		}
	}
	return slice
}

func (s *SchemaDAO) DeleteRef(ctx context.Context, refRequest *schema.RefRequest) error {
	rollbackOperations := []FileDoRecord{}
	domainProject := util.ParseDomainProject(ctx)
	serviceID := refRequest.ServiceID
	schemaID := refRequest.SchemaID
	schemaPath := filepath.Join(schema.RootFilePath, domainProject, serviceID, schemaID+".json")

	// get the mutex lock
	servicepath := filepath.Join(schema.RootFilePath, domainProject, serviceID)

	mutexLock.Lock()
	mutex, ok := mutexMap[servicepath]
	if !ok {
		mutex = &sync.Mutex{}
		mutexMap[servicepath] = mutex
	}
	mutexLock.Unlock()

	// add lock
	mutex.Lock()
	defer mutex.Unlock()

	err := deleteFile(schemaPath, &rollbackOperations)

	if err != nil {
		log.Error("Occur error when delete schema file, begain rollback... ", err)
		rollback(rollbackOperations)
		return err
	}

	// update schemas in service
	service, err := datasource.GetMetadataManager().GetService(ctx, &discovery.GetServiceRequest{
		ServiceId: serviceID,
	})
	if err != nil {
		log.Error(fmt.Sprintf("get service[%s] failed", serviceID), err)
		rollback(rollbackOperations)
		return err
	}

	service.Schemas = removeStringFromSlice(service.Schemas, schemaID)

	err = updateServiceSchema(ctx, serviceID, service)
	if err != nil {
		rollback(rollbackOperations)
		return err
	}
	return nil
}

func (s *SchemaDAO) GetContent(ctx context.Context, contentRequest *schema.ContentRequest) (*schema.Content, error) {
	// no usage, should not be called
	log.Error("Occur error when call SchemaDAO.GetContent, this method should not be called in any condition", schema.ErrSchemaNotFound)
	return nil, schema.ErrSchemaNotFound
}

func (s *SchemaDAO) PutContent(ctx context.Context, contentRequest *schema.PutContentRequest) error {
	rollbackOperations := []FileDoRecord{}
	domainProject := util.ParseDomainProject(ctx)
	serviceID := contentRequest.ServiceID
	servicepath := filepath.Join(schema.RootFilePath, domainProject, serviceID)
	mutexLock.Lock()
	mutex, ok := mutexMap[servicepath]
	if !ok {
		mutex = &sync.Mutex{}
		mutexMap[servicepath] = mutex
	}
	mutexLock.Unlock()

	// add lock
	mutex.Lock()
	defer mutex.Unlock()

	var err error

	defer func() {
		if err != nil {
			rollback(rollbackOperations)
		}
	}()

	if err != nil {
		log.Error("Occur error when clean schema files before update schemas, begain rollback... ", err)
		return err
	}

	// update file
	schemaBytes, marshalErr := json.Marshal(contentRequest.Content)
	err = marshalErr
	if err != nil {
		openlog.Error("fail to marshal kv " + err.Error())
		return err
	}

	err = createOrUpdateFile(filepath.Join(servicepath, contentRequest.SchemaID+".json"), schemaBytes, &rollbackOperations)
	if err != nil {
		log.Error("Occur error when create schema files when update schemas, begain rollback... ", err)
		return err
	}

	// update service schema
	service, serviceErr := datasource.GetMetadataManager().GetService(ctx, &discovery.GetServiceRequest{
		ServiceId: serviceID,
	})
	err = serviceErr
	if err != nil {
		log.Error(fmt.Sprintf("get service[%s] failed when update schemas", serviceID), err)
		return err
	}

	var schemaIdValid = false
	for _, serviceSchemaId := range service.Schemas {
		if serviceSchemaId == contentRequest.SchemaID {
			schemaIdValid = true
		}
	}
	if !schemaIdValid {
		err = schema.ErrSchemaNotFound
		log.Error(fmt.Sprintf("update service[%s] failed when valide schema id", serviceID), err)
		return err
	}

	err = updateServiceSchema(ctx, serviceID, service)
	if err != nil {
		log.Error(fmt.Sprintf("update service[%s] failed when update schemas", serviceID), err)
		return err
	}
	return nil
}

func updateServiceSchema(ctx context.Context, serviceID string, service *discovery.MicroService) error {
	// update schemas in service
	domainProject := util.ParseDomainProject(ctx)
	body, err := json.Marshal(service)
	if err != nil {
		log.Error("marshal service failed", err)
		return err
	}

	var options []etcdadpt.OpOptions
	serviceKey := path.GenerateServiceKey(domainProject, serviceID)
	options = append(options, etcdadpt.OpPut(etcdadpt.WithStrKey(serviceKey), etcdadpt.WithValue(body)))

	// update service task
	serviceOpts, err := etcdsync.GenUpdateOpts(ctx, datasource.ResourceKV, body, etcdsync.WithOpts(map[string]string{"key": serviceKey}))
	if err != nil {
		log.Error("fail to create update opts", err)
	}
	options = append(options, serviceOpts...)
	err = etcdadpt.Txn(ctx, options)

	return err
}

func (s *SchemaDAO) PutManyContent(ctx context.Context, contentRequest *schema.PutManyContentRequest) error {
	rollbackOperations := []FileDoRecord{}
	domainProject := util.ParseDomainProject(ctx)
	serviceID := contentRequest.ServiceID

	if len(contentRequest.SchemaIDs) != len(contentRequest.Contents) {
		log.Error(fmt.Sprintf("service[%s] contents request invalid", serviceID), nil)
		return discovery.NewError(discovery.ErrInvalidParams, "contents request invalid")
	}

	// get the mutex lock
	servicepath := filepath.Join(schema.RootFilePath, domainProject, serviceID)

	mutexLock.Lock()
	mutex, ok := mutexMap[servicepath]
	if !ok {
		mutex = &sync.Mutex{}
		mutexMap[servicepath] = mutex
	}
	mutexLock.Unlock()

	// add lock
	mutex.Lock()
	defer mutex.Unlock()

	var err error

	defer func() {
		if err != nil {
			rollback(rollbackOperations)
		}
	}()

	// get all the files under this dir
	existedFiles, readErr := os.ReadDir(servicepath)
	err = readErr
	if err != nil && !errors.Is(err, syscall.ERROR_FILE_NOT_FOUND) && !errors.Is(err, syscall.ENOTDIR) {
		return err
	}
	err = nil

	// clean existed files
	for _, file := range existedFiles {
		if file.IsDir() {
			continue
		}
		filepath := servicepath + "/" + file.Name()
		err = deleteFile(filepath, &rollbackOperations)
		if err != nil {
			break
		}
	}
	if err != nil {
		log.Error("Occur error when clean schema files before update schemas, begain rollback... ", err)
		return err
	}

	// create or update files
	for i := 0; i < len(contentRequest.SchemaIDs); i++ {
		schemaId := contentRequest.SchemaIDs[i]
		schema := contentRequest.Contents[i]

		schemaBytes, marshalErr := json.Marshal(schema)
		err = marshalErr
		if err != nil {
			openlog.Error("fail to marshal kv " + err.Error())
			return err
		}
		err = createOrUpdateFile(servicepath+"/"+schemaId+".json", schemaBytes, &rollbackOperations)
		if err != nil {
			break
		}
	}

	if err != nil {
		log.Error("Occur error when create schema files when update schemas, begain rollback... ", err)
		return err
	}

	// update service schema
	if contentRequest.Init {
		return nil
	}
	service, err := datasource.GetMetadataManager().GetService(ctx, &discovery.GetServiceRequest{
		ServiceId: serviceID,
	})
	if err != nil {
		log.Error(fmt.Sprintf("get service[%s] failed when update schemas", serviceID), err)
		return err
	}

	service.Schemas = contentRequest.SchemaIDs

	err = updateServiceSchema(ctx, serviceID, service)
	if err != nil {
		log.Error(fmt.Sprintf("update service[%s] failed when update schemas", serviceID), err)
	}
	return err
}

func (s *SchemaDAO) DeleteContent(ctx context.Context, contentRequest *schema.ContentRequest) error {
	// no usage, should not be called
	log.Error("Occur error when call SchemaDAO.DeleteContent, this method should not be called in any condition", schema.ErrSchemaContentNotFound)
	return schema.ErrSchemaContentNotFound
}

func (s *SchemaDAO) DeleteNoRefContents(ctx context.Context) (int, error) {
	// no usage, should not be called
	log.Error("Occur error when call SchemaDAO.DeleteNoRefContents, this method should not be called in any condition", schema.ErrSchemaNotFound)
	return 0, schema.ErrSchemaNotFound
}
