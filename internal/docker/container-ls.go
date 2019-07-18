package nrdocker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/newrelic/infra-integrations-sdk/integration"
	log "github.com/newrelic/infra-integrations-sdk/log"
	"source.datanerd.us/FIT/nri-docker/internal/lib"
)

// GetContainerInfo x
func GetContainerInfo(cli *client.Client, entity *integration.Entity) {
	ctx := context.Background()
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{All: true, Since: "24 hours ago"})
	if err != nil {
		lib.ErrorLogToInsights(err, entity)
	}

	var wg sync.WaitGroup
	wg.Add(len(containers))
	for _, container := range containers {
		go func(container types.Container) {
			defer wg.Done()
			FetchStats(ctx, container, cli, entity)
		}(container)
	}
	wg.Wait()
}

// FetchStats x
func FetchStats(ctx context.Context, container types.Container, cli *client.Client, entity *integration.Entity) {
	metricSet := lib.NewMetricSet("dockerContainerSample", entity)
	lib.SetMetric(metricSet, "hostname", lib.Hostname)
	lib.SetMetric(metricSet, "containerID", container.ID)
	lib.SetMetric(metricSet, "IDShort", container.ID[0:12])
	lib.SetMetric(metricSet, "image", container.Image)
	img := strings.Split(container.Image, "@")
	lib.SetMetric(metricSet, "imageShort", img[0])
	lib.SetMetric(metricSet, "imageID", container.ImageID)
	lib.SetMetric(metricSet, "state", container.State)
	lib.SetMetric(metricSet, "status", container.Status)
	lib.SetMetric(metricSet, "command", container.Command)
	lib.SetMetric(metricSet, "created", container.Created)
	lib.SetMetric(metricSet, "duration", lib.MakeTimestamp()-(container.Created*1000))

	if container.NetworkSettings.Networks != nil {
		for i, networkSetting := range container.NetworkSettings.Networks {
			lib.SetMetric(metricSet, "network."+i+".ipv4", networkSetting.IPAddress)
			lib.SetMetric(metricSet, "network."+i+".gateway", networkSetting.Gateway)
			lib.SetMetric(metricSet, "network."+i+".ipv6", networkSetting.GlobalIPv6Address)
		}
	}

	lib.SetMetric(metricSet, "networkMode", container.HostConfig.NetworkMode)
	lib.SetMetric(metricSet, "sizeRw", container.SizeRw)
	lib.SetMetric(metricSet, "sizeRootFs", container.SizeRootFs)

	for key, val := range container.Labels {
		lib.SetMetric(metricSet, key, val)
	}

	b := new(bytes.Buffer)
	fmt.Fprintf(b, "%v,", container.Ports)
	lib.SetMetric(metricSet, "ports", b.String())

	b = new(bytes.Buffer)
	fmt.Fprintf(b, "%v,", container.Mounts)
	lib.SetMetric(metricSet, "mounts", b.String())

	//docker stats data
	stats, err := cli.ContainerStats(ctx, container.ID, false)
	if err != nil {
		log.Debug(err.Error())
	} else {
		var containerStats types.StatsJSON
		json.NewDecoder(stats.Body).Decode(&containerStats)

		netRx, netTx, netRxErrors, netTxErrors, netRxDropped, netTxDropped, netRxPackets, netTxPackets := calculateNetwork(containerStats.Networks)
		lib.SetMetric(metricSet, "netRx", netRx)
		lib.SetMetric(metricSet, "netTx", netTx)
		lib.SetMetric(metricSet, "netRxErrors", netRxErrors)
		lib.SetMetric(metricSet, "netTxErrors", netTxErrors)
		lib.SetMetric(metricSet, "netRxDropped", netRxDropped)
		lib.SetMetric(metricSet, "netTxDropped", netTxDropped)
		lib.SetMetric(metricSet, "netRxPackets", netRxPackets)
		lib.SetMetric(metricSet, "netTxPackets", netTxPackets)

		if stats.OSType == "windows" {
			lib.SetMetric(metricSet, "cpuPercent", calculateCPUPercentWindows(containerStats))
			lib.SetMetric(metricSet, "blkReadSizeBytes", containerStats.StorageStats.ReadSizeBytes)
			lib.SetMetric(metricSet, "blkWriteSizeBytes", containerStats.StorageStats.WriteSizeBytes)
			lib.SetMetric(metricSet, "mem", float64(containerStats.MemoryStats.PrivateWorkingSet))
			lib.SetMetric(metricSet, "numProcs", containerStats.NumProcs)
			lib.SetMetric(metricSet, "memCommitBytes", containerStats.MemoryStats.Commit)
			lib.SetMetric(metricSet, "memCommitPeakBytes", containerStats.MemoryStats.CommitPeak)
			lib.SetMetric(metricSet, "memPrivateWorkingSet", containerStats.MemoryStats.PrivateWorkingSet)
		} else {
			lib.SetMetric(metricSet, "previousCPU", containerStats.PreCPUStats.CPUUsage.TotalUsage)
			lib.SetMetric(metricSet, "onlineCPUs", containerStats.CPUStats.OnlineCPUs)
			lib.SetMetric(metricSet, "systemUsage", containerStats.CPUStats.SystemUsage)
			lib.SetMetric(metricSet, "cpuPercent", calculateCPUPercentUnix(containerStats.PreCPUStats.CPUUsage.TotalUsage, containerStats.PreCPUStats.SystemUsage, containerStats))
			blkReadBytes, blkWriteBytes := calculateBlockIO(containerStats.BlkioStats)
			lib.SetMetric(metricSet, "blkReadBytes", blkReadBytes)
			lib.SetMetric(metricSet, "blkWriteBytes", blkWriteBytes)
			mem := calculateMemUsageUnixNoCache(containerStats.MemoryStats)
			lib.SetMetric(metricSet, "mem", mem)
			memLimit := float64(containerStats.MemoryStats.Limit)
			lib.SetMetric(metricSet, "memLimit", memLimit)
			lib.SetMetric(metricSet, "memPercent", calculateMemPercentUnixNoCache(memLimit, mem))
			lib.SetMetric(metricSet, "memUsage", float64(containerStats.MemoryStats.Usage))
			lib.SetMetric(metricSet, "memMaxUsage", float64(containerStats.MemoryStats.MaxUsage))
			lib.SetMetric(metricSet, "memFailCount", float64(containerStats.MemoryStats.Failcnt))
			lib.SetMetric(metricSet, "pidsStatsCurrent", float64(containerStats.PidsStats.Current))
			lib.SetMetric(metricSet, "pidsStatsLimit", float64(containerStats.PidsStats.Limit))
			lib.SetMetric(metricSet, "periods", float64(containerStats.CPUStats.ThrottlingData.Periods))
			lib.SetMetric(metricSet, "throttledPeriods", float64(containerStats.CPUStats.ThrottlingData.ThrottledPeriods))
			lib.SetMetric(metricSet, "throttledTime", float64(containerStats.CPUStats.ThrottlingData.ThrottledTime))
		}
	}

	//docker inspect data
	containerInspect, err := cli.ContainerInspect(ctx, container.ID)
	if err != nil {
		log.Debug(err.Error())
	} else {
		lib.SetMetric(metricSet, "restartCount", containerInspect.RestartCount)
		lib.SetMetric(metricSet, "platform", containerInspect.Platform)
		lib.SetMetric(metricSet, "driver", containerInspect.Driver)
		lib.SetMetric(metricSet, "nanoCPUs", containerInspect.HostConfig.NanoCPUs)
		lib.SetMetric(metricSet, "cpuShares", containerInspect.HostConfig.CPUShares)

		if containerInspect.Node != nil {
			lib.SetMetric(metricSet, "nodeID", containerInspect.Node.ID)
			lib.SetMetric(metricSet, "nodeName", containerInspect.Node.Name)
			lib.SetMetric(metricSet, "nodeAddr", containerInspect.Node.Addr)
			lib.SetMetric(metricSet, "nodeMemory", containerInspect.Node.Memory)
			for key, val := range containerInspect.Node.Labels {
				lib.SetMetric(metricSet, key, val)
			}
		}

		lib.SetMetric(metricSet, "pid", containerInspect.State.Pid)
		if containerInspect.State.Health != nil {
			lib.SetMetric(metricSet, "failingStreak", containerInspect.State.Health.FailingStreak)
			lib.SetMetric(metricSet, "finishedAt", containerInspect.State.FinishedAt)
		}

		if stats.OSType == "windows" {
			lib.SetMetric(metricSet, "cpuCount", containerInspect.HostConfig.CPUCount)
			lib.SetMetric(metricSet, "ioMaximumIOps", containerInspect.HostConfig.IOMaximumIOps)
			lib.SetMetric(metricSet, "ioMaximumBandwidth", containerInspect.HostConfig.IOMaximumBandwidth)
			lib.SetMetric(metricSet, "isolation", containerInspect.HostConfig.Isolation)
		} else {
			lib.SetMetric(metricSet, "cgroupParent", containerInspect.HostConfig.CgroupParent)
			lib.SetMetric(metricSet, "cpuPeriod", containerInspect.HostConfig.CPUPeriod)
			lib.SetMetric(metricSet, "cpuQuota", containerInspect.HostConfig.CPUQuota)
			lib.SetMetric(metricSet, "cpuRealtimePeriod", containerInspect.HostConfig.CPURealtimePeriod)
			lib.SetMetric(metricSet, "CPURealtimeRuntime", containerInspect.HostConfig.CPURealtimeRuntime)
			lib.SetMetric(metricSet, "blkioWeight", containerInspect.HostConfig.BlkioWeight)
			// lib.SetMetric(metricSet, "diskQuota", containerInspect.HostConfig.DiskQuota)
			lib.SetMetric(metricSet, "kernelMemory", containerInspect.HostConfig.KernelMemory)
			lib.SetMetric(metricSet, "memoryReservation", containerInspect.HostConfig.MemoryReservation)
			lib.SetMetric(metricSet, "memorySwap", containerInspect.HostConfig.MemorySwap)
			lib.SetMetric(metricSet, "memorySwappiness", containerInspect.HostConfig.MemorySwappiness)
		}
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
