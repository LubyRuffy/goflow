package gocodefuncs

import (
	"fmt"
	"github.com/LubyRuffy/goflow/utils"
	"github.com/Ullaakut/nmap/v2"
	"github.com/mitchellh/mapstructure"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type ScanPortParam struct {
	Targets string `json:"targets"` // 扫描目标
	Ports   string `json:"ports"`   // 扫描端口
	NmapDir string `json:"nmapDir"` // 扫描端口
}

// ScanPort 扫描端口
// 参数: hosts/ports
// 输出格式：{"ip":"117.161.125.154","port":80,"base_protocol":"tcp","service":"http","hostnames":"fofa.info"}
func ScanPort(p Runner, params map[string]interface{}) *FuncResult {
	var err error
	var options ScanPortParam
	if err = mapstructure.Decode(params, &options); err != nil {
		panic(fmt.Errorf("screenShot failed: %w", err))
	}

	if options.Targets == "" {
		options.Targets = "127.0.0.1"
	}
	if options.Ports == "" {
		options.Ports = "22,80,443,1080,3389,8080,8443"
	}

	opts := []nmap.Option{
		nmap.WithTargets(options.Targets),
		nmap.WithPorts(options.Ports),
		nmap.WithOpenOnly(),
	}

	var searchFilePath []string
	// 用户指定的路径优先
	if len(options.NmapDir) > 0 {
		searchFilePath = append(searchFilePath, []string{
			filepath.Join(options.NmapDir, "nmap.exe"),
			filepath.Join(options.NmapDir, "Nmap", "nmap.exe"),
		}...)

		defaultNmapPath := utils.LoadFirstExistsFile(searchFilePath)
		if len(defaultNmapPath) > 0 {
			opts = append(opts, nmap.WithBinaryPath(defaultNmapPath))
			log.Println("use user defined nmap path:", defaultNmapPath)
		}
	}

	scanner, err := nmap.NewScanner(
		opts...,
	)
	if err != nil {
		if err != nmap.ErrNmapNotInstalled {
			panic(fmt.Errorf("ScanPort error: %w", err))
		}

		searchFilePath = append(searchFilePath, []string{
			"nmap.exe",
			filepath.Join(os.Getenv("PROGRAMFILES"), "Nmap", "nmap.exe"),
			filepath.Join(os.Getenv("PROGRAMFILES(X86)"), "Nmap", "nmap.exe"),
		}...)

		defaultNmapPath := utils.LoadFirstExistsFile(searchFilePath)
		if len(defaultNmapPath) != 0 {
			opts = append(opts, nmap.WithBinaryPath(defaultNmapPath))
			scanner, err = nmap.NewScanner(
				opts...,
			)
		}
		if err != nil {
			panic(fmt.Errorf("ScanPort error: %w", err))
		}
	}

	progress := make(chan float32, 1)
	// Function to listen and print the progress
	go func() {
		for percent := range progress {
			//p.Debugf("ScanPort Progress: %v %%\n", percent)
			//fmt.Printf("Progress: %v %%\n", percent)
			p.SetProgress(float64(percent) / 100)
		}
	}()

	result, _, err := scanner.RunWithProgress(progress)
	if err != nil {
		panic(fmt.Errorf("ScanPort error: %w", err))
	}

	var fn string
	fn, err = utils.WriteTempFile(".json", func(f *os.File) error {
		// 遍历host
		for _, host := range result.Hosts {
			if len(host.Ports) == 0 || len(host.Addresses) == 0 {
				continue
			}
			var hostnames []string
			for i := range host.Hostnames {
				hostnames = append(hostnames, host.Hostnames[i].Name)
			}
			for _, addr := range host.Addresses {
				for _, port := range host.Ports {
					_, err := f.WriteString(fmt.Sprintf(`{"ip":"%s","port":%d,"base_protocol":"%s","service":"%s","hostnames":"%s"}`+"\n",
						addr, port.ID, port.Protocol, port.Service.Name, strings.Join(hostnames, ",")))
					if err != nil {
						return err
					}
				}
			}
		}

		return err
	})
	if err != nil {
		panic(fmt.Errorf("ScanPort error: %w", err))
	}

	return &FuncResult{
		OutFile: fn,
	}
}
