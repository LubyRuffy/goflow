package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var (
	lastCheckDockerTime        time.Time // 最后检查docker路径的时间
	defaultDockerPath          = "docker"
	defaultCheckDockerDuration = 5 * time.Minute
	globalDockerOK             = false
)

// DockerRun 运行docker，解决Windows找不到的问题
// 注意：exec.ExitError 错误会被忽略，我们只关心所有的字符串返回，不关注进程的错误代码
func DockerRun(args ...string) ([]byte, error) {
	// 缓存5分钟
	if time.Now().Sub(lastCheckDockerTime) < defaultCheckDockerDuration {
		if globalDockerOK {
			return RunCmdNoExitError(exec.Command(defaultDockerPath, args...).CombinedOutput())
		} else {
			return nil, fmt.Errorf("docker status is not ok")
		}
	}
	lastCheckDockerTime = time.Now()

	d, err := RunCmdNoExitError(exec.Command(defaultDockerPath, "version").CombinedOutput())
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			err = nil
		} else {
			// 可能路径不在PATH环境变量，需要自己找，主要是windows
			// https://docs.microsoft.com/en-us/windows/deployment/usmt/usmt-recognized-environment-variables
			defaultDockerPath = LoadFirstExistsFile([]string{
				"docker.exe",
				filepath.Join(os.Getenv("PROGRAMFILES"), "Docker", "Docker", "resources", "bin", "docker.exe"),
				filepath.Join(os.Getenv("PROGRAMFILES(X86)"), "Docker", "Docker", "resources", "bin", "docker.exe"),
			})
			if len(defaultDockerPath) == 0 {
				return nil, fmt.Errorf("could not find docker")
			}
			d, err = RunCmdNoExitError(exec.Command(defaultDockerPath, "version").CombinedOutput())
		}
	}
	if err == nil {
		if strings.Contains(string(d), "API version") {
			globalDockerOK = true
			return RunCmdNoExitError(exec.Command(defaultDockerPath, args...).CombinedOutput())
		} else {
			err = fmt.Errorf("docker is invalid")
		}
	}

	return nil, err
}

// DockerStatusOk 检查是否安装
func DockerStatusOk() bool {
	data, err := DockerRun("images")
	if err != nil {
		return false
	}
	return strings.Contains(string(data), "REPOSITORY")
}
