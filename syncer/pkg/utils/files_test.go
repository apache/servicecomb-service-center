package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExist(t *testing.T) {
	if IsDirExist("./test/file"){
		t.Error("dir exist failed")
	}

	if IsFileExist("./test/file"){
		t.Error("file exist failed")
	}
}

func TestOpenFile(t *testing.T) {
	fileName := "file"
	f, err := OpenFile(fileName)
	if err != nil {
		t.Errorf("open file failed: %s", err)
	}
	f.Close()

	f, err = OpenFile(fileName)
	if err != nil {
		t.Errorf("open file failed: %s", err)
	}
	f.Close()
	os.RemoveAll(fileName)

	fileName = "./test/file"
	f, err = OpenFile(fileName)
	if err != nil {
		t.Errorf("open file failed: %s", err)
	}
	f.Close()
	os.RemoveAll(filepath.Dir(fileName))
}
