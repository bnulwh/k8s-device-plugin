package nvidia

import (
	"fmt"
	"strings"
	"time"

	log "github.com/astaxie/beego/logs"
	"golang.org/x/net/context"
	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1alpha"
)

var (
	clientTimeout    = 30 * time.Second
	lastAllocateTime time.Time
)

//// create docker client
//func init() {
//	kubeInit()
//}

//func buildErrResponse(reqs *pluginapi.AllocateRequest, podReqGPU uint) *pluginapi.AllocateResponse {
//
//	responses := pluginapi.AllocateResponse{
//
//	}
//	for _, req := range reqs.DevicesIDs {
//		response := pluginapi.ContainerAllocateResponse{
//			Envs: map[string]string{
//				envNVGPU:               fmt.Sprintf("no-gpu-has-%dMiB-to-run", podReqGPU),
//				EnvResourceIndex:       fmt.Sprintf("-1"),
//				EnvResourceByPod:       fmt.Sprintf("%d", podReqGPU),
//				EnvResourceByContainer: fmt.Sprintf("%d", uint(len(req.DevicesIDs))),
//				EnvResourceByDev:       fmt.Sprintf("%d", getGPUMemory()),
//			},
//		}
//		responses.ContainerResponses = append(responses.ContainerResponses, &response)
//	}
//	return &responses
//}

// Allocate which return list of devices.
func (m *NvidiaDevicePlugin) Allocate(ctx context.Context, r *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	devs := m.devs
	log.Info("devs: %v", devs)
	log.Info("device ids: %v", r.DevicesIDs)
	response := pluginapi.AllocateResponse{
		Envs: map[string]string{
			"NVIDIA_VISIBLE_DEVICES": strings.Join(r.DevicesIDs, ","),
		},
	}

	for _, id := range r.DevicesIDs {
		if !deviceExists(devs, id) {
			return nil, fmt.Errorf("invalid allocation request: unknown device: %s", id)
		}
	}

	return &response, nil
}
