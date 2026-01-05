package utils

import (
	"os"
	"path/filepath"
)

// FindConfigFile 查找配置文件，尝试多个可能的路径
// 这是解决配置文件路径问题的通用方案，适用于从不同目录运行程序的情况
//
// 参数:
//   - configFileName: 配置文件名（如 "config_debug.yaml"）
//   - possiblePaths: 可能的配置文件路径列表（按优先级排序）
//
// 返回:
//   - 找到的配置文件绝对路径，如果都找不到则返回第一个路径（用于错误提示）
//
// 使用示例:
//
//	configPath := utils.FindConfigFile("config_debug.yaml", []string{
//	    "configs/config_debug.yaml",           // 从项目根目录运行
//	    "../../configs/config_debug.yaml",     // 从 cmd/server 目录运行
//	    "../configs/config_debug.yaml",        // 从 cmd 目录运行
//	})
func FindConfigFile(configFileName string, possiblePaths []string) string {
	// 如果 possiblePaths 为空，使用默认路径列表
	if len(possiblePaths) == 0 {
		possiblePaths = []string{
			"configs/" + configFileName,       // 从项目根目录运行
			"../../configs/" + configFileName, // 从 cmd/server 目录运行
			"../configs/" + configFileName,    // 从 cmd 目录运行
			"./configs/" + configFileName,     // 当前目录
		}
	}

	// 查找第一个存在的配置文件
	for _, path := range possiblePaths {
		// 尝试转换为绝对路径
		if absPath, err := filepath.Abs(path); err == nil {
			if _, err := os.Stat(absPath); err == nil {
				return absPath
			}
		}
		// 也尝试直接使用相对路径
		if _, err := os.Stat(path); err == nil {
			// 转换为绝对路径返回
			if absPath, err := filepath.Abs(path); err == nil {
				return absPath
			}
			return path
		}
	}

	// 如果所有路径都不存在，返回第一个路径（用于错误提示）
	if len(possiblePaths) > 0 {
		return possiblePaths[0]
	}
	return "configs/" + configFileName
}

// FindConfigFileWithMode 根据运行模式查找配置文件
// 这是一个便捷函数，自动根据 mode 选择对应的配置文件
//
// 参数:
//   - mode: 运行模式（"debug" 或 "release"）
//   - possibleBasePaths: 可能的配置文件基础路径列表（不包含文件名）
//
// 返回:
//   - 找到的配置文件绝对路径
//
// 使用示例:
//
//	configPath := utils.FindConfigFileWithMode("debug", []string{
//	    "configs",
//	    "../../configs",
//	    "../configs",
//	})
func FindConfigFileWithMode(mode string, possibleBasePaths []string) string {
	var configFileName string
	if mode == "release" {
		configFileName = "config_release.yaml"
	} else {
		configFileName = "config_debug.yaml"
	}

	// 构建可能的路径列表
	var possiblePaths []string
	if len(possibleBasePaths) == 0 {
		possibleBasePaths = []string{"configs", "../../configs", "../configs", "./configs"}
	}

	for _, basePath := range possibleBasePaths {
		possiblePaths = append(possiblePaths, filepath.Join(basePath, configFileName))
	}

	return FindConfigFile(configFileName, possiblePaths)
}
