package utils

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

func FileExist(path string) bool {
	_, err := os.Lstat(path)
	return !os.IsNotExist(err)
}

// GetMainDirectory 获取项目根目录
func GetMainDirectory() string {
	dir, err := filepath.Abs("./")
	if err != nil {
		log.Fatal(err)
	}
	//兼容linux "\\"转"/"
	return strings.Replace(dir, "\\", "/", -1)
}
