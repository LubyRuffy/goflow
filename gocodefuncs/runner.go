package gocodefuncs

import (
	"context"
	"github.com/sirupsen/logrus"
)

type Runner interface {
	GetObject(name string) (interface{}, bool)                   // 查询全局注册的对象，用于内部调用，比如fofacli
	SetObject(name string, value interface{})                    // 设置全局变量
	GetLastFile() string                                         // GetLastFile 获取最后一次生成的文件
	GetContext() context.Context                                 // GetContext 获取ctx
	Debugf(format string, args ...interface{})                   // 打印调试信息
	Warnf(format string, args ...interface{})                    // 打印警告信息
	Logf(level logrus.Level, format string, args ...interface{}) // 打印日志信息
	SetProgress(p float64)                                       // 设置进度
}

// Artifact 过程中生成的文件
type Artifact struct {
	FilePath string `json:"filePath,omitempty"` // 文件路径
	FileName string `json:"fileName,omitempty"` // 文件路径
	FileSize int    `json:"fileSize,omitempty"` // 文件大小
	FileType string `json:"fileType,omitempty"` // 文件类型
	Memo     string `json:"memo,omitempty"`     // 备注，比如URL等
}

// FuncResult 返回的结构
type FuncResult struct {
	OutFile   string      `json:"outFile,omitempty"`   // 往后传递的文件
	Artifacts []*Artifact `json:"artifacts,omitempty"` // 中间文件
}
