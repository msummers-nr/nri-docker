package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/docker/docker/api/types"
)

func getContainerInfo() {
	ctx := context.Background()
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{All: true, Since: "24 hours ago"})
	if err != nil {
		errorLogToInsights(err)
	}

	var wg sync.WaitGroup
	wg.Add(len(containers))
	for _, container := range containers {
		go func(container types.Container) {
			defer wg.Done()
			fetchStats(ctx, container)
		}(container)
	}
	wg.Wait()
}

func fetchStats(ctx context.Context, container types.Container) {
	metricSet := newMetricSet("dockerContainerSample")
	setMetric(metricSet, "hostname", hostname)

	// docker container stats
	stats, _ := cli.ContainerStats(ctx, container.ID, false)
	var containerStats types.StatsJSON
	json.NewDecoder(stats.Body).Decode(&containerStats)

	setMetric(metricSet, "containerID", container.ID)
	setMetric(metricSet, "IDShort", container.ID[0:12])
	setMetric(metricSet, "image", container.Image)
	img := strings.Split(container.Image, "@")
	setMetric(metricSet, "imageShort", img[0])
	setMetric(metricSet, "imageID", container.ImageID)
	setMetric(metricSet, "state", container.State)
	setMetric(metricSet, "status", container.Status)
	setMetric(metricSet, "command", container.Command)
	setMetric(metricSet, "created", container.Created)
	setMetric(metricSet, "duration", makeTimestamp()-(container.Created*1000))

	setMetric(metricSet, "networkMode", container.HostConfig.NetworkMode)
	setMetric(metricSet, "sizeRw", container.SizeRw)
	setMetric(metricSet, "sizeRootFs", container.SizeRootFs)

	for key, val := range container.Labels {
		setMetric(metricSet, key, val)
	}

	b := new(bytes.Buffer)
	fmt.Fprintf(b, "%v,", container.Ports)
	setMetric(metricSet, "ports", b.String())

	b = new(bytes.Buffer)
	fmt.Fprintf(b, "%v,", container.Mounts)
	setMetric(metricSet, "mounts", b.String())

	netRx, netTx, netRxErrors, netTxErrors, netRxDropped, netTxDropped, netRxPackets, netTxPackets := calculateNetwork(containerStats.Networks)

	setMetric(metricSet, "netRx", netRx)
	setMetric(metricSet, "netTx", netTx)
	setMetric(metricSet, "netRxErrors", netRxErrors)
	setMetric(metricSet, "netTxErrors", netTxErrors)
	setMetric(metricSet, "netRxDropped", netRxDropped)
	setMetric(metricSet, "netTxDropped", netTxDropped)
	setMetric(metricSet, "netRxPackets", netRxPackets)
	setMetric(metricSet, "netTxPackets", netTxPackets)

	if stats.OSType == "windows" {
		setMetric(metricSet, "cpuPercent", calculateCPUPercentWindows(containerStats))
		setMetric(metricSet, "blkReadSizeBytes", containerStats.StorageStats.ReadSizeBytes)
		setMetric(metricSet, "blkWriteSizeBytes", containerStats.StorageStats.WriteSizeBytes)
		setMetric(metricSet, "mem", float64(containerStats.MemoryStats.PrivateWorkingSet))
		setMetric(metricSet, "numProcs", containerStats.NumProcs)
		setMetric(metricSet, "memCommitBytes", containerStats.MemoryStats.Commit)
		setMetric(metricSet, "memCommitPeakBytes", containerStats.MemoryStats.CommitPeak)
		setMetric(metricSet, "memPrivateWorkingSet", containerStats.MemoryStats.PrivateWorkingSet)
	} else {
		setMetric(metricSet, "previousCPU", containerStats.PreCPUStats.CPUUsage.TotalUsage)
		setMetric(metricSet, "onlineCPUs", containerStats.CPUStats.OnlineCPUs)
		setMetric(metricSet, "systemUsage", containerStats.CPUStats.SystemUsage)
		setMetric(metricSet, "cpuPercent", calculateCPUPercentUnix(containerStats.PreCPUStats.CPUUsage.TotalUsage, containerStats.PreCPUStats.SystemUsage, containerStats))
		blkReadBytes, blkWriteBytes := calculateBlockIO(containerStats.BlkioStats)
		setMetric(metricSet, "blkReadBytes", blkReadBytes)
		setMetric(metricSet, "blkWriteBytes", blkWriteBytes)
		mem := calculateMemUsageUnixNoCache(containerStats.MemoryStats)
		setMetric(metricSet, "mem", mem)
		memLimit := float64(containerStats.MemoryStats.Limit)
		setMetric(metricSet, "memLimit", memLimit)
		setMetric(metricSet, "memPercent", calculateMemPercentUnixNoCache(memLimit, mem))
		setMetric(metricSet, "memUsage", float64(containerStats.MemoryStats.Usage))
		setMetric(metricSet, "memMaxUsage", float64(containerStats.MemoryStats.MaxUsage))
		setMetric(metricSet, "memFailCount", float64(containerStats.MemoryStats.Failcnt))
		setMetric(metricSet, "pidsStatsCurrent", float64(containerStats.PidsStats.Current))
		setMetric(metricSet, "pidsStatsLimit", float64(containerStats.PidsStats.Limit))
		setMetric(metricSet, "periods", float64(containerStats.CPUStats.ThrottlingData.Periods))
		setMetric(metricSet, "throttledPeriods", float64(containerStats.CPUStats.ThrottlingData.ThrottledPeriods))
		setMetric(metricSet, "throttledTime", float64(containerStats.CPUStats.ThrottlingData.ThrottledTime))
	}

	//docker inspect data
	containerInspect, _ := cli.ContainerInspect(ctx, container.ID)
	setMetric(metricSet, "restartCount", containerInspect.RestartCount)
	setMetric(metricSet, "platform", containerInspect.Platform)
	setMetric(metricSet, "driver", containerInspect.Driver)
	setMetric(metricSet, "nanoCPUs", containerInspect.HostConfig.NanoCPUs)
	setMetric(metricSet, "cpuShares", containerInspect.HostConfig.CPUShares)

	if containerInspect.Node != nil {
		setMetric(metricSet, "nodeID", containerInspect.Node.ID)
		setMetric(metricSet, "nodeName", containerInspect.Node.Name)
		setMetric(metricSet, "nodeAddr", containerInspect.Node.Addr)
		setMetric(metricSet, "nodeMemory", containerInspect.Node.Memory)
		for key, val := range containerInspect.Node.Labels {
			setMetric(metricSet, key, val)
		}
	}

	setMetric(metricSet, "pid", containerInspect.State.Pid)
	if containerInspect.State.Health != nil {
		setMetric(metricSet, "failingStreak", containerInspect.State.Health.FailingStreak)
		setMetric(metricSet, "finishedAt", containerInspect.State.FinishedAt)
	}

	if stats.OSType == "windows" {
		setMetric(metricSet, "cpuCount", containerInspect.HostConfig.CPUCount)
		setMetric(metricSet, "ioMaximumIOps", containerInspect.HostConfig.IOMaximumIOps)
		setMetric(metricSet, "ioMaximumBandwidth", containerInspect.HostConfig.IOMaximumBandwidth)
		setMetric(metricSet, "isolation", containerInspect.HostConfig.Isolation)
	} else {
		setMetric(metricSet, "cgroupParent", containerInspect.HostConfig.CgroupParent)
		setMetric(metricSet, "cpuPeriod", containerInspect.HostConfig.CPUPeriod)
		setMetric(metricSet, "cpuQuota", containerInspect.HostConfig.CPUQuota)
		setMetric(metricSet, "cpuRealtimePeriod", containerInspect.HostConfig.CPURealtimePeriod)
		setMetric(metricSet, "CPURealtimeRuntime", containerInspect.HostConfig.CPURealtimeRuntime)
		setMetric(metricSet, "blkioWeight", containerInspect.HostConfig.BlkioWeight)
		setMetric(metricSet, "diskQuota", containerInspect.HostConfig.DiskQuota)
		setMetric(metricSet, "kernelMemory", containerInspect.HostConfig.KernelMemory)
		setMetric(metricSet, "memoryReservation", containerInspect.HostConfig.MemoryReservation)
		setMetric(metricSet, "memorySwap", containerInspect.HostConfig.MemorySwap)
		setMetric(metricSet, "memorySwappiness", containerInspect.HostConfig.MemorySwappiness)
	}
}

func calculateCPUPercentUnix(previousCPU, previousSystem uint64, v types.StatsJSON) float64 {
	var (
		cpuPercent = 0.0
		// calculate the change for the cpu usage of the container in between readings
		cpuDelta = float64(v.CPUStats.CPUUsage.TotalUsage) - float64(previousCPU)
		// calculate the change for the entire system between readings
		systemDelta = float64(v.CPUStats.SystemUsage) - float64(previousSystem)
	)

	if systemDelta > 0.0 && cpuDelta > 0.0 {
		cpuPercent = (cpuDelta / systemDelta) * float64(len(v.CPUStats.CPUUsage.PercpuUsage)) * 100.0
	}
	return cpuPercent
}

func calculateCPUPercentWindows(v types.StatsJSON) float64 {
	// Max number of 100ns intervals between the previous time read and now
	possIntervals := uint64(v.Read.Sub(v.PreRead).Nanoseconds()) // Start with number of ns intervals
	possIntervals /= 100                                         // Convert to number of 100ns intervals
	possIntervals *= uint64(v.NumProcs)                          // Multiple by the number of processors

	// Intervals used
	intervalsUsed := v.CPUStats.CPUUsage.TotalUsage - v.PreCPUStats.CPUUsage.TotalUsage

	// Percentage avoiding divide-by-zero
	if possIntervals > 0 {
		return float64(intervalsUsed) / float64(possIntervals) * 100.0
	}
	return 0.00
}

func calculateBlockIO(blkio types.BlkioStats) (uint64, uint64) {
	var blkRead, blkWrite uint64
	for _, bioEntry := range blkio.IoServiceBytesRecursive {
		switch strings.ToLower(bioEntry.Op) {
		case "read":
			blkRead = blkRead + bioEntry.Value
		case "write":
			blkWrite = blkWrite + bioEntry.Value
		}
	}
	return blkRead, blkWrite
}

func calculateNetwork(network map[string]types.NetworkStats) (float64, float64, float64, float64, float64, float64, float64, float64) {
	var rx, tx, rxErrors, txErrors, rxDropped, txDropped, rxPackets, txPackets float64
	for _, v := range network {
		rx += float64(v.RxBytes)
		tx += float64(v.TxBytes)
		rxPackets += float64(v.RxPackets)
		txPackets += float64(v.TxPackets)
		rxErrors += float64(v.RxErrors)
		txErrors += float64(v.TxErrors)
		rxDropped += float64(v.RxDropped)
		txDropped += float64(v.TxDropped)
	}
	return rx, tx, rxErrors, txErrors, rxDropped, txDropped, rxPackets, txPackets
}

// calculateMemUsageUnixNoCache calculate memory usage of the container.
// Page cache is intentionally excluded to avoid misinterpretation of the output.
func calculateMemUsageUnixNoCache(mem types.MemoryStats) float64 {
	return float64(mem.Usage - mem.Stats["cache"])
}

func calculateMemPercentUnixNoCache(limit float64, usedNoCache float64) float64 {
	// MemoryStats.Limit will never be 0 unless the container is not running and we haven't
	// got any data from cgroup
	if limit != 0 {
		return usedNoCache / limit * 100.0
	}
	return 0
}
