package utils

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

// FileExist 判断文件是否存在
func FileExist(path string) bool {
	_, err := os.Lstat(path)
	return !os.IsNotExist(err)
}

// GetWorkDirectory 获取工作目录
func GetWorkDirectory() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	//兼容linux "\\"转"/"
	return strings.Replace(dir, "\\", "/", -1)
}

// GetFileDirectory 获取执行文件目录
func GetFileDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	//兼容linux "\\"转"/"
	return strings.Replace(dir, "\\", "/", -1)
}
