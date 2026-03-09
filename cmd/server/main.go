package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

type ServiceManager struct {
	rootPath string
	cmds     []*exec.Cmd
}

func NewServiceManager(rootPath string) *ServiceManager {
	return &ServiceManager{
		rootPath: rootPath,
		cmds:     make([]*exec.Cmd, 0),
	}
}

var Service []string

func (sm *ServiceManager) startServices(dirNames []string, ext string) error {
	var dirPath string
	if len(dirNames) == 0 {
		dirPath = sm.rootPath
	} else {
		dirPath = filepath.Join(append([]string{sm.rootPath}, dirNames...)...)
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// 递归检查子目录
			if err := sm.startServices(append(dirNames, entry.Name()), ext); err != nil {
				return err
			}
			continue
		}

		// 检查是否是.go文件且不是测试文件
		if strings.HasSuffix(entry.Name(), ext) && !strings.HasSuffix(entry.Name(), "_test.go") {
			// 执行go run命令
			cmd := exec.Command("go", "run", filepath.Join(dirPath, entry.Name()))
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Start(); err != nil {
				log.Printf("Error starting service %s: %v", entry.Name(), err)
				continue
			}
			log.Printf("Started service: %s (PID: %d)", entry.Name(), cmd.Process.Pid)
			sm.cmds = append(sm.cmds, cmd)
		}
	}
	return nil
}

func (sm *ServiceManager) stopServices() {
	for _, cmd := range sm.cmds {
		if cmd.Process != nil {
			log.Printf("Stopping service PID: %d", cmd.Process.Pid)
			// 发送中断信号
			cmd.Process.Signal(syscall.SIGINT)

			// 等待进程结束
			done := make(chan error)
			go func() {
				done <- cmd.Wait()
			}()

			// 超时等待
			select {
			case err := <-done:
				if err != nil {
					log.Printf("Service exited with error: %v", err)
				} else {
					log.Printf("Service exited successfully")
				}
			case <-time.After(30 * time.Second):
				log.Printf("Service did not exit in time, force killing")
				cmd.Process.Kill()
			}
		}
	}
}

func main() {
	// 解析命令行参数
	var services string
	flag.StringVar(&services, "services", "", "Comma-separated list of services to start (optional)")
	var servicePath string
	flag.StringVar(&servicePath, "path", "./apis", "Path to services directory (default: ./apis)")
	flag.Parse()

	if services != "" {
		Service = strings.Split(services, ",")
	}

	// 创建服务管理器
	manager := NewServiceManager(servicePath)

	// 设置信号处理
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 启动服务
	if err := manager.startServices([]string{}, ".go"); err != nil {
		log.Fatalf("Failed to start services: %v", err)
	}

	log.Println("All services started successfully. Press Ctrl+C to stop.")

	// 等待停止信号
	<-quit

	log.Println("Shutting down services...")
	manager.stopServices()
	log.Println("All services stopped.")
}
