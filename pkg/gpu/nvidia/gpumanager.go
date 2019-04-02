package nvidia

import (
	"fmt"
	"syscall"
	"time"

	"github.com/NVIDIA/gpu-monitoring-tools/bindings/go/nvml"
	log "github.com/astaxie/beego/logs"
	"github.com/fsnotify/fsnotify"
	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1alpha"
)

type sharedGPUManager struct {
	enableMPS   bool
	healthCheck bool
}

func NewSharedGPUManager(enableMPS, healthCheck bool, bp MemoryUnit) *sharedGPUManager {
	metric = bp
	return &sharedGPUManager{
		enableMPS:   enableMPS,
		healthCheck: healthCheck,
	}
}

func (ngm *sharedGPUManager) Run() error {
	log.Info("Loading NVML")

	if err := nvml.Init(); err != nil {
		log.Warning("Failed to initialize NVML: %s.", err)
		log.Warning("If this is a GPU node, did you set the docker default runtime to `nvidia`?")
		select {}
	}
	defer func() { log.Info("Shutdown of NVML returned:", nvml.Shutdown()) }()

	log.Info("Fetching devices.")
	if getDeviceCount() == uint(0) {
		log.Info("No devices found. Waiting indefinitely.")
		select {}
	}

	log.Info("Starting FS watcher.")
	watcher, err := newFSWatcher(pluginapi.DevicePluginPath)
	if err != nil {
		log.Warning("Failed to created FS watcher.")
		return err
	}
	defer watcher.Close()

	log.Info("Starting OS watcher.")
	sigs := newOSWatcher(syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	restart := true
	var devicePlugin *NvidiaDevicePlugin

L:
	for {
		if restart {
			if devicePlugin != nil {
				devicePlugin.Stop()
			}

			devicePlugin = NewNvidiaDevicePlugin(ngm.enableMPS, ngm.healthCheck)
			if err := devicePlugin.Serve(); err != nil {
				log.Warning("Failed to start device plugin due to %v", err)
			} else {
				restart = false
			}
		}

		select {
		case event := <-watcher.Events:
			if event.Name == pluginapi.KubeletSocket && event.Op&fsnotify.Create == fsnotify.Create {
				log.Info("inotify: %s created, restarting.", pluginapi.KubeletSocket)
				restart = true
			}

		case err := <-watcher.Errors:
			log.Warning("inotify: %s", err)

		case s := <-sigs:
			switch s {
			case syscall.SIGHUP:
				log.Info("Received SIGHUP, restarting.")
				restart = true
			case syscall.SIGQUIT:
				t := time.Now()
				timestamp := fmt.Sprint(t.Format("20060102150405"))
				log.Info("generate core dump")
				coreDump("/etc/kubernetes/go_" + timestamp + ".txt")
			default:
				log.Info("Received signal \"%v\", shutting down.", s)
				devicePlugin.Stop()
				break L
			}
		}
	}

	return nil
}
