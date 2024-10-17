package common

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func IsPlugin(path string) bool {
	return filepath.Ext(path) == ".so"
}

func IsGoFile(path string) bool {
	return filepath.Ext(path) == ".go"
}

func IsDirectory(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fileInfo.IsDir()
}

func CopyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func(source *os.File) {
		err := source.Close()
		if err != nil {
			return
		}
	}(source)
	_ = os.MkdirAll(filepath.Dir(dst), os.ModePerm)

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func(destination *os.File) {
		err := destination.Close()
		if err != nil {
			return
		}
	}(destination)

	_, err = io.Copy(destination, source)
	return err
}

func CompileGoFile(src, dst string) error {
	_, err := os.Stat(dst)
	if !os.IsNotExist(err) {
		err = os.Remove(dst)
		if err != nil {
			return err
		}
	}

	cmd := exec.Command("go", "build", "-buildmode=plugin", "-o", dst, src)
	err = cmd.Run()
	return err
}
