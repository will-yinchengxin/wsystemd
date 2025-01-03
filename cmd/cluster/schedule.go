package cluster

import "wsystemd/cmd/http/core"

var (
	// taskCount/cpuUsage/memUsage/loadUsage
	ScheduleTaskCount = "taskCount"
	ScheduleCpuUsage  = "cpuUsage"
	ScheduleMemUsage  = "memUsage"
	ScheduleLoadUsage = "loadUsage"
)

func GetWorkNode() (string, error) {
	base, err := WkMg.getWorkBase()
	if err != nil {
		return "", err
	}
	schedule := core.CoreConfig["schedule"]
	if schedule == ScheduleTaskCount {
		return FindLeastTasksNode(base), nil
	}
	if schedule == ScheduleMemUsage {
		return FindLeastMemoryNode(base), nil
	}
	if schedule == ScheduleCpuUsage {
		return FindLeastCPUNode(base), nil
	}
	if schedule == ScheduleLoadUsage {
		return FindLeastLoadNode(base), nil
	}
	return "", nil
}

// FindLeastTasksNode 基于任务数进行调度
func FindLeastTasksNode(nodeStats map[string]ResourceInfo) string {
	var targetNode string
	minTasks := int(^uint64(0) >> 1)

	for node, count := range nodeStats {
		if count.TaskCount < minTasks {
			minTasks = count.TaskCount
			targetNode = node
		}
	}

	return targetNode
}

// FindLeastCPUNode 基于CPU使用率进行调度
func FindLeastCPUNode(nodeStats map[string]ResourceInfo) string {
	var targetNode string
	minCPU := float64(^uint64(0) >> 1)

	for node, cpu := range nodeStats {
		if cpu.CPUUsage < minCPU {
			minCPU = cpu.CPUUsage
			targetNode = node
		}
	}

	return targetNode
}

// FindLeastLoadNode 基于Load进行调度
func FindLeastLoadNode(nodeStats map[string]ResourceInfo) string {
	var targetNode string
	minLoad := float64(^uint64(0) >> 1)
	for node, load := range nodeStats {
		if load.LoadUsage < minLoad {
			minLoad = load.LoadUsage
			targetNode = node
		}
	}

	return targetNode
}

// FindLeastMemoryNode 基于 MemoryUsage 进行调度
func FindLeastMemoryNode(nodeStats map[string]ResourceInfo) string {
	var targetNode string
	minMemory := float64(^uint64(0) >> 1)
	for node, memory := range nodeStats {
		if memory.MemoryUsage < minMemory {
			minMemory = memory.MemoryUsage
			targetNode = node
		}
	}
	return targetNode
}
