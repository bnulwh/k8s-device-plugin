// Copyright (c) 2017, NVIDIA CORPORATION. All rights reserved.

package nvidia

import (
	"fmt"
	log "github.com/astaxie/beego/logs"
	"strings"

	"github.com/NVIDIA/gpu-monitoring-tools/bindings/go/nvml"

	"golang.org/x/net/context"
	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1alpha"
)

var (
	gpuMemory uint
	metric    MemoryUnit
)

func check(err error) {
	if err != nil {
		log.Critical("Fatal:", err)
	}
}

func generateFakeDeviceID(realID string, fakeCounter uint) string {
	return fmt.Sprintf("%s-_-%d", realID, fakeCounter)
}

func extractRealDeviceID(fakeDeviceID string) string {
	return strings.Split(fakeDeviceID, "-_-")[0]
}

func setGPUMemory(raw uint) {
	v := raw
	if metric == GiBPrefix {
		v = raw / 1024
	}
	gpuMemory = v
	log.Info("set gpu memory: %d", gpuMemory)
}

func getGPUMemory() uint {
	return gpuMemory
}

func getDeviceCount() uint {
	n, err := nvml.GetDeviceCount()
	check(err)
	return n
}

func displayDevice(d *nvml.Device) {
	log.Info("======= device : =============")
	log.Info("Path: %v", d.Path)
	log.Info("UUID: %v", d.UUID)
	log.Info("CPUAffinity: %v", d.CPUAffinity)
	log.Info("Memory: %v", d.Memory)
	log.Info("Power: %v", d.Power)
	log.Info("Model: %v", d.Model)
	log.Info("Clocks: %v MHz", d.Clocks)
	log.Info("PCI:", d.PCI)
}
func displayDeviceStatus(device *nvml.Device) {
	st, err := device.Status()
	if err != nil {
		log.Critical("Error getting device %d status: %v\n", device.Path, err)
	}
	log.Info("-------------------------------------")
	log.Info("path  Power Temp GPU%  Mem%  Enc   Dec   MemC  Cores")
	log.Info("%5d %5d %5d %5d %5d %5d %5d %5d %5d",
		device.Path, *st.Power, *st.Temperature, *st.Utilization.GPU, *st.Utilization.Memory,
		*st.Utilization.Encoder, *st.Utilization.Decoder, *st.Clocks.Memory, *st.Clocks.Cores)
	log.Info("-------------------------------------")
}
func displayProcessInfo(d *nvml.Device) {
	pInfo, err := d.GetAllRunningProcesses()
	if err != nil {
		log.Warning("Error getting device %s processes: %v\n", d.Path, err)
	}
	log.Info("----------------------------")
	log.Info("Path  PID   Type  Mem   Name")
	if len(pInfo) == 0 {
		log.Info("%5v %5s %5s %5s %-5s\n", d.Path, "-", "-", "-", "-")
	}
	for j := range pInfo {
		log.Info("%5v %5v %5v %5v %-5v\n",
			d.Path, pInfo[j].PID, pInfo[j].Type, pInfo[j].MemoryUsed, pInfo[j].Name)
	}
	log.Info("----------------------------")
}

func getDevices() ([]*pluginapi.Device, map[string]uint) {
	n, err := nvml.GetDeviceCount()
	check(err)

	var devs []*pluginapi.Device
	realDevNames := map[string]uint{}
	for i := uint(0); i < n; i++ {
		d, err := nvml.NewDevice(i)
		check(err)
		displayDevice(d)
		displayDeviceStatus(d)
		displayProcessInfo(d)
		// realDevNames = append(realDevNames, d.UUID)
		var id uint
		log.Info("Deivce %s's Path is %s", d.UUID, d.Path)
		_, err = fmt.Sscanf(d.Path, "/dev/nvidia%d", &id)
		check(err)
		realDevNames[d.UUID] = id
		// var KiB uint64 = 1024
		log.Info("# device %s's Memory: %d", d.UUID, uint(*d.Memory))
		if getGPUMemory() == uint(0) {
			setGPUMemory(uint(*d.Memory))
		}
		gmem := getGPUMemory()
		for j := uint(0); j < gmem; j++ {
			fakeID := generateFakeDeviceID(d.UUID, j)
			if j == 0 {
				log.Info("# Add first device ID: " + fakeID)
			}
			if j == gmem-1 {
				log.Info("# Add last device ID: " + fakeID)
			}
			devs = append(devs, &pluginapi.Device{
				ID:     fakeID,
				Health: pluginapi.Healthy,
			})
		}
	}

	return devs, realDevNames
}

func deviceExists(devs []*pluginapi.Device, id string) bool {
	for _, d := range devs {
		if d.ID == id {
			return true
		}
	}
	return false
}

func watchXIDs(ctx context.Context, devs []*pluginapi.Device, xids chan<- *pluginapi.Device) {
	eventSet := nvml.NewEventSet()
	defer nvml.DeleteEventSet(eventSet)

	for _, d := range devs {
		realDeviceID := extractRealDeviceID(d.ID)
		err := nvml.RegisterEventForDevice(eventSet, nvml.XidCriticalError, realDeviceID)
		if err != nil && strings.HasSuffix(err.Error(), "Not Supported") {
			log.Warning("Warning: %s (%s) is too old to support healthchecking: %s. Marking it unhealthy.", realDeviceID, d.ID, err)

			xids <- d
			continue
		}

		if err != nil {
			log.Critical("Fatal:", err)
		}
		log.Info("register event for device %s ok", d.ID)
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		e, err := nvml.WaitForEvent(eventSet, 5000)
		if err != nil && e.Etype != nvml.XidCriticalError {
			continue
		}

		// FIXME: formalize the full list and document it.
		// http://docs.nvidia.com/deploy/xid-errors/index.html#topic_4
		// Application errors: the GPU should still be healthy
		if e.Edata == 31 || e.Edata == 43 || e.Edata == 45 {
			continue
		}

		if e.UUID == nil || len(*e.UUID) == 0 {
			// All devices are unhealthy
			for _, d := range devs {
				xids <- d
			}
			continue
		}

		for _, d := range devs {
			if extractRealDeviceID(d.ID) == *e.UUID {
				xids <- d
			}
		}
	}
}
