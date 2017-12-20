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
package traceutils

import (
	"archive/zip"
	"fmt"
	"github.com/ServiceComb/service-center/pkg/util"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

var pathReplacer *strings.Replacer

func EscapPath(msg string) string {
	return pathReplacer.Replace(msg)
}

func removeFile(path string) error {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return err
	}
	if fileInfo.IsDir() {
		return nil
	}
	err = os.Remove(path)
	if err != nil {
		return err
	}
	return nil
}

func removeExceededFiles(path string, baseFileName string,
	maxKeptCount int, rotateStage string) {
	if maxKeptCount < 0 {
		return
	}
	fileList := make([]string, 0, 2*maxKeptCount)
	var pat string
	if rotateStage == "rollover" {
		//rotated file, svc.log.20060102150405000
		pat = fmt.Sprintf(`%s\.[0-9]{1,17}$`, baseFileName)
	} else if rotateStage == "backup" {
		//backup compressed file, svc.log.20060102150405000.zip
		pat = fmt.Sprintf(`%s\.[0-9]{17}\.zip$`, baseFileName)
	} else {
		return
	}
	fileList, err := FilterFileList(path, pat)
	if err != nil {
		util.Logger().Error("filepath.Walk() "+EscapPath(path)+" failed", err)
		return
	}
	sort.Strings(fileList)
	if len(fileList) <= maxKeptCount {
		return
	}
	//remove exceeded files, keep file count below maxBackupCount
	for len(fileList) > maxKeptCount {
		filePath := fileList[0]
		util.Logger().Warn("remove "+EscapPath(filePath), nil)
		err := removeFile(filePath)
		if err != nil {
			util.Logger().Error("remove "+EscapPath(filePath)+" failed", err)
			break
		}
		//remove the first element of a list
		fileList = append(fileList[:0], fileList[1:]...)
	}
}

//filePath: file full path, like ${_APP_LOG_DIR}/svc.log.1
//fileBaseName: rollover file base name, like svc.log
//replaceTimestamp: whether or not to replace the num. of a rolled file
func compressFile(filePath, fileBaseName string, replaceTimestamp bool) error {
	ifp, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer ifp.Close()

	var zipFilePath string
	if replaceTimestamp {
		//svc.log.1 -> svc.log.20060102150405000.zip
		zipFileBase := fileBaseName + "." + getTimeStamp() + "." + "zip"
		zipFilePath = filepath.Dir(filePath) + "/" + zipFileBase
	} else {
		zipFilePath = filePath + ".zip"
	}
	zipFile, err := os.OpenFile(zipFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0440)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	ofp, err := zipWriter.Create(filepath.Base(filePath))
	if err != nil {
		return err
	}

	_, err = io.Copy(ofp, ifp)
	if err != nil {
		return err
	}

	return nil
}

func shouldRollover(fPath string, MaxFileSize int) bool {
	if MaxFileSize <= 0 {
		return false
	}

	fileInfo, err := os.Stat(fPath)
	if err != nil {
		util.Logger().Error("state "+EscapPath(fPath)+" failed", err)
		return false
	}

	if fileInfo.Size() > int64(MaxFileSize*1024*1024) {
		return true
	}
	return false
}

func doRollover(fPath string, MaxFileSize int, MaxBackupCount int) {
	if !shouldRollover(fPath, MaxFileSize) {
		return
	}

	timeStamp := getTimeStamp()
	//absolute path
	rotateFile := fPath + "." + timeStamp
	err := CopyFile(fPath, rotateFile)
	if err != nil {
		util.Logger().Error("copy "+EscapPath(fPath)+" failed", err)
	}

	//truncate the file
	f, err := os.OpenFile(fPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		util.Logger().Error("truncate "+EscapPath(fPath)+" failed", err)
		return
	}
	f.Close()

	//remove exceeded rotate files
	removeExceededFiles(filepath.Dir(fPath), filepath.Base(fPath), MaxBackupCount, "rollover")
}

func doBackup(fPath string, MaxBackupCount int) {
	if MaxBackupCount <= 0 {
		return
	}
	pat := fmt.Sprintf(`%s\.[0-9]{1,17}$`, filepath.Base(fPath))
	rotateFileList, err := FilterFileList(filepath.Dir(fPath), pat)
	if err != nil {
		util.Logger().Error("walk"+EscapPath(fPath)+" failed", err)
		return
	}

	for _, file := range rotateFileList {
		var err error
		p := fmt.Sprintf(`%s\.[0-9]{17}$`, filepath.Base(fPath))
		if ret, _ := regexp.MatchString(p, file); ret {
			//svc.log.20060102150405000, not replace Timestamp
			err = compressFile(file, filepath.Base(fPath), false)
		} else {
			//svc.log.1, replace Timestamp
			err = compressFile(file, filepath.Base(fPath), true)
		}
		if err != nil {
			util.Logger().Error("compress"+EscapPath(file)+" failed", err)
			continue
		}
		err = removeFile(file)
		if err != nil {
			util.Logger().Error("remove"+EscapPath(file)+" failed", err)
		}
	}

	//remove exceeded backup files
	removeExceededFiles(filepath.Dir(fPath), filepath.Base(fPath), MaxBackupCount, "backup")
}

func LogRotateFile(file string, MaxFileSize int, MaxBackupCount int) {
	defer func() {
		if e := recover(); e != nil {
			util.Logger().Errorf(nil, "LogRotate file %s catch an exception, err: %v.", EscapPath(file), e)
		}
	}()

	doRollover(file, MaxFileSize, MaxBackupCount)
	doBackup(file, MaxBackupCount)
}

//path:			where log files need rollover
//MaxFileSize: 		MaxSize of a file before rotate. By M Bytes.
//MaxBackupCount: 	Max counts to keep of a log's backup files.
func LogRotate(path string, MaxFileSize int, MaxBackupCount int) {
	//filter .log .trace files
	defer func() {
		if e := recover(); e != nil {
			util.Logger().Errorf(nil, "LogRotate catch an exception, err: %v.", e)
		}
	}()

	pat := `.(\.log|\.trace|\.out)$`
	fileList, err := FilterFileList(path, pat)
	if err != nil {
		util.Logger().Error("filepath.Walk() "+EscapPath(path)+" failed", err)
		return
	}

	for _, file := range fileList {
		LogRotateFile(file, MaxFileSize, MaxBackupCount)
	}
}

//path : where the file will be filtered
//pat  : regexp pattern to filter the matched file
func FilterFileList(path, pat string) ([]string, error) {
	capacity := 10
	//initialize a fileName slice, len=0, cap=10
	fileList := make([]string, 0, capacity)
	err := filepath.Walk(path,
		func(pathName string, f os.FileInfo, e error) error {
			if f == nil {
				return e
			}
			if f.IsDir() {
				return nil
			}
			if pat != "" {
				ret, _ := regexp.MatchString(pat, f.Name())
				if !ret {
					return nil
				}
			}
			fileList = append(fileList, pathName)
			return nil
		})
	return fileList, err
}

func getTimeStamp() string {
	now := time.Now().Format("2006.01.02.15.04.05.000")
	timeSlot := strings.Replace(now, ".", "", -1)
	return timeSlot
}

func CopyFile(srcFile, destFile string) error {
	file, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer file.Close()
	dest, err := os.Create(destFile)
	if err != nil {
		return err
	}
	defer dest.Close()
	_, err = io.Copy(dest, file)
	return err
}

type LogRotateConfig struct {
	Dir         string
	Period      time.Duration
	MaxFileSize int
	BackupCount int
}

func RunLogRotate(cfg *LogRotateConfig) {
	util.Go(func(stopCh <-chan struct{}) {
		for {
			select {
			case <-stopCh:
				return
			case <-time.After(cfg.Period):
				LogRotate(cfg.Dir, cfg.MaxFileSize, cfg.BackupCount)
			}
		}
	})
}

func init() {
	pathReplacer = strings.NewReplacer(
		os.ExpandEnv("${APP_ROOT}"), "APP_ROOT",
		os.ExpandEnv("${_APP_SHARE_DIR}"), "_APP_SHARE_DIR",
		os.ExpandEnv("${_APP_TMP_DIR}"), "_APP_TMP_DIR",
		os.ExpandEnv("${SSL_ROOT}"), "SSL_ROOT",
		os.ExpandEnv("${CIPHER_ROOT}"), "CIPHER_ROOT",
		os.ExpandEnv("${_APP_LOG_DIR}"), "_APP_LOG_DIR",
		os.ExpandEnv("${INSTALL_ROOT}"), "INSTALL_ROOT")
}
