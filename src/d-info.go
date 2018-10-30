package main

import (
	"context"
	"fmt"
)

func getHostInfo() {
	ctx := context.Background()
	info, err := cli.Info(ctx)
	if err == nil {
		metricSet := newMetricSet("dockerInfoSample")
		setMetric(metricSet, "containers", info.Containers)
		setMetric(metricSet, "containersRunning", info.ContainersRunning)
		setMetric(metricSet, "containersPaused", info.ContainersPaused)
		setMetric(metricSet, "containersStopped", info.ContainersStopped)
		setMetric(metricSet, "images", info.Images)
		setMetric(metricSet, "clientVersion", cli.ClientVersion())
		setMetric(metricSet, "kernelVersion", info.KernelVersion)
		setMetric(metricSet, "osType", info.OSType)
		setMetric(metricSet, "arch", info.Architecture)
		setMetric(metricSet, "name", info.Name)
		setMetric(metricSet, "ID", info.ID)
		setMetric(metricSet, "goroutines", info.NGoroutines)
		setMetric(metricSet, "cpus", info.NCPU)
		setMetric(metricSet, "memTotal", info.MemTotal)
		setMetric(metricSet, "eventsListeners", info.NEventsListener)
		setMetric(metricSet, "swapLimit", info.SwapLimit)
		setMetric(metricSet, "memTotal", info.MemTotal)
		setMetric(metricSet, "memLimit", info.MemoryLimit)

		serverVersion, err := cli.ServerVersion(ctx)
		if err == nil {
			setMetric(metricSet, "serverVersion", serverVersion.Version)
			setMetric(metricSet, "serverGoVersion", serverVersion.GoVersion)
			setMetric(metricSet, "serverApiVersion", serverVersion.APIVersion)
			setMetric(metricSet, "serverGitCommit", serverVersion.GitCommit)
			setMetric(metricSet, "serverBuildTime", serverVersion.BuildTime)
		}

		setMetric(metricSet, "swarmState", fmt.Sprintf("%v", info.Swarm.LocalNodeState))
		setMetric(metricSet, "swarmControlAvailable", fmt.Sprintf("%v", info.Swarm.ControlAvailable))
		setMetric(metricSet, "swarmError", info.Swarm.Error)
		setMetric(metricSet, "swarmNodeID", info.Swarm.NodeID)
		setMetric(metricSet, "swarmNodes", info.Swarm.Nodes)

		if info.Swarm.Cluster != nil {
			setMetric(metricSet, "swarmClusterCreatedAt", info.Swarm.Cluster.CreatedAt.Unix())
			setMetric(metricSet, "swarmClusterDuration", makeTimestamp()-(info.Swarm.Cluster.CreatedAt.Unix()*1000))
			setMetric(metricSet, "swarmClusterUpdatedAt", info.Swarm.Cluster.UpdatedAt.Unix())
			setMetric(metricSet, "swarmClusterID", info.Swarm.Cluster.ID)
			setMetric(metricSet, "swarmClusterVersionIndex", info.Swarm.Cluster.Version.Index)
			setMetric(metricSet, "swarmManagers", info.Swarm.Managers)
			setMetric(metricSet, "swarmNodeAddr", info.Swarm.NodeAddr)
		}

	} else {
		errorLogToInsights(entity, err)
	}
}
