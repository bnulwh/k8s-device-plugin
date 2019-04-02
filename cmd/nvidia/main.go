// Copyright (c) 2017, NVIDIA CORPORATION. All rights reserved.

package main

import (
	"flag"
	"fmt"
	log "github.com/astaxie/beego/logs"
	"github.com/bnulwh/k8s-device-plugin/pkg/gpu/nvidia"
	"os"
)

const logPath = "/var/log"

var (
	mps         = flag.Bool("mps", false, "Enable or Disable MPS")
	healthCheck = flag.Bool("health-check", false, "Enable or disable Health check")
	memoryUnit  = flag.String("memory-unit", "MiB", "Set memoryUnit of the GPU Memroy, support 'GiB' and 'MiB'")
)

func init() {
	beegoInit()
}

func beegoInit() {
	log.EnableFuncCallDepth(true)
	log.SetLogFuncCallDepth(3)
	if !pathExists(logPath) {
		fmt.Printf("dir: %s not found.", logPath)
		err := os.MkdirAll(logPath, 0711)
		if err != nil {
			fmt.Printf("mkdir %s failed: %v", logPath, err)
		}
	}
	err := log.SetLogger(log.AdapterMultiFile, `{"filename":"/var/log/nvidia.log","separate":["emergency", "alert", 
			"critical", "error", "warning", "notice", "info", "debug"]}`)
	if err != nil {
		fmt.Println(err)
	}
	err = log.SetLogger(log.AdapterConsole, `{"level":6}`)
	if err != nil {
		fmt.Println(err)
	}

}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func main() {
	flag.Parse()
	log.Info("Start gpushare device plugin")
	log.Info("mps: %v", mps)
	log.Info("healthCheck: %v", healthCheck)
	log.Info("memoryUnit: %v", memoryUnit)
	ngm := nvidia.NewSharedGPUManager(*mps, *healthCheck, translateMemoryUnits(*memoryUnit))
	err := ngm.Run()
	if err != nil {
		log.Critical("Failed due to %v", err)
	}
}

func translateMemoryUnits(value string) nvidia.MemoryUnit {
	memoryUnit := nvidia.MemoryUnit(value)
	switch memoryUnit {
	case nvidia.MiBPrefix:
	case nvidia.GiBPrefix:
	default:
		log.Warning("Unsupported memory unit: %s, use memoryUnit Gi as default", value)
		memoryUnit = nvidia.GiBPrefix
	}

	return memoryUnit
}
