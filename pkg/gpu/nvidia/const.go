package nvidia

import (
	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1alpha"
)

// MemoryUnit describes GPU Memory, now only supports Gi, Mi
type MemoryUnit string

const (
	resourceName  = "shared-gpu/gpu-mem"
	resourceCount = "shared-gpu/gpu-count"
	serverSock    = pluginapi.DevicePluginPath + "gpushare.sock"

	OptimisticLockErrorMsg = "the object has been modified; please apply your changes to the latest version and try again"

	allHealthChecks             = "xids"
	containerTypeLabelKey       = "io.kubernetes.docker.type"
	containerTypeLabelSandbox   = "podsandbox"
	containerTypeLabelContainer = "container"
	containerLogPathLabelKey    = "io.kubernetes.container.logpath"
	sandboxIDLabelKey           = "io.kubernetes.sandbox.id"

	envNVGPU               = "NVIDIA_VISIBLE_DEVICES"
	EnvResourceIndex       = "SHARED_GPU_MEM_IDX"
	EnvResourceByPod       = "SHARED_GPU_MEM_POD"
	EnvResourceByContainer = "SHARED_GPU_MEM_CONTAINER"
	EnvResourceByDev       = "SHARED_GPU_MEM_DEV"
	EnvAssignedFlag        = "SHARED_GPU_MEM_ASSIGNED"
	EnvResourceAssumeTime  = "SHARED_GPU_MEM_ASSUME_TIME"
	EnvResourceAssignTime  = "SHARED_GPU_MEM_ASSIGN_TIME"

	GiBPrefix = MemoryUnit("GiB")
	MiBPrefix = MemoryUnit("MiB")
)
