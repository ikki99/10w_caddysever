package config

import (
	"os"
	"path/filepath"
)

var (
	AppDir       string
	DataDir      string
	CaddyDir     string
	CaddyBin     string
	CaddyConfig  string
	CaddyLogFile string
	DatabasePath string
)

func Init() error {
	// 获取程序目录
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	AppDir = filepath.Dir(exe)

	// 设置数据目录
	DataDir = filepath.Join(AppDir, "data")
	CaddyDir = filepath.Join(DataDir, "caddy")
	CaddyBin = filepath.Join(CaddyDir, "caddy.exe")
	CaddyConfig = filepath.Join(CaddyDir, "Caddyfile")
	CaddyLogFile = filepath.Join(CaddyDir, "caddy.log")
	DatabasePath = filepath.Join(DataDir, "caddy-manager.db")

	// 创建必要的目录
	dirs := []string{DataDir, CaddyDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}
