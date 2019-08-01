package nrdocker

import (
	"context"
	"fmt"

	"github.com/docker/docker/client"
	"github.com/newrelic/infra-integrations-sdk/integration"
	"source.datanerd.us/FIT/nri-docker/internal/lib"
)

// GetHostInfo x
func GetHostInfo(cli *client.Client, entity *integration.Entity) {
	ctx := context.Background()
	info, err := cli.Info(ctx)
	if err == nil {
		metricSet := lib.NewMetricSet("dockerInfoSample", entity)
		lib.SetMetric(metricSet, "containers", info.Containers)
		lib.SetMetric(metricSet, "containersRunning", info.ContainersRunning)
		lib.SetMetric(metricSet, "containersPaused", info.ContainersPaused)
		lib.SetMetric(metricSet, "containersStopped", info.ContainersStopped)
		lib.SetMetric(metricSet, "images", info.Images)
		lib.SetMetric(metricSet, "clientVersion", cli.ClientVersion())
		lib.SetMetric(metricSet, "kernelVersion", info.KernelVersion)
		lib.SetMetric(metricSet, "osType", info.OSType)
		lib.SetMetric(metricSet, "arch", info.Architecture)
		lib.SetMetric(metricSet, "nodeName", info.Name) // correlation purpose
		lib.SetMetric(metricSet, "host", info.Name)     // correlation purpose
		lib.SetMetric(metricSet, "name", info.Name)
		lib.SetMetric(metricSet, "ID", info.ID)
		lib.SetMetric(metricSet, "goroutines", info.NGoroutines)
		lib.SetMetric(metricSet, "cpus", info.NCPU)
		lib.SetMetric(metricSet, "memTotal", info.MemTotal)
		lib.SetMetric(metricSet, "eventsListeners", info.NEventsListener)
		lib.SetMetric(metricSet, "swapLimit", info.SwapLimit)
		lib.SetMetric(metricSet, "memTotal", info.MemTotal)
		lib.SetMetric(metricSet, "memLimit", info.MemoryLimit)

		for _, label := range info.Labels {
			lib.ApplyLabel(label, metricSet, "")
		}

		serverVersion, err := cli.ServerVersion(ctx)
		if err == nil {
			lib.SetMetric(metricSet, "serverVersion", serverVersion.Version)
			lib.SetMetric(metricSet, "serverGoVersion", serverVersion.GoVersion)
			lib.SetMetric(metricSet, "serverApiVersion", serverVersion.APIVersion)
			lib.SetMetric(metricSet, "serverGitCommit", serverVersion.GitCommit)
			lib.SetMetric(metricSet, "serverBuildTime", serverVersion.BuildTime)
		}

		lib.SetMetric(metricSet, "swarmState", fmt.Sprintf("%v", info.Swarm.LocalNodeState))
		lib.SwarmState = fmt.Sprintf("%v", info.Swarm.LocalNodeState)
		lib.SetMetric(metricSet, "swarmControlAvailable", fmt.Sprintf("%v", info.Swarm.ControlAvailable))
		lib.SetMetric(metricSet, "swarmError", info.Swarm.Error)
		lib.SetMetric(metricSet, "swarmNodeID", info.Swarm.NodeID)
		lib.SetMetric(metricSet, "swarmNodes", info.Swarm.Nodes)
		lib.SetMetric(metricSet, "swarmManagers", info.Swarm.Managers)
		lib.SetMetric(metricSet, "swarmNodeAddr", info.Swarm.NodeAddr)

		if info.Swarm.Cluster != nil {
			lib.SetMetric(metricSet, "swarmClusterCreatedAt", info.Swarm.Cluster.CreatedAt.Unix())
			lib.SetMetric(metricSet, "swarmClusterDuration", lib.MakeTimestamp()-(info.Swarm.Cluster.CreatedAt.Unix()*1000))
			lib.SetMetric(metricSet, "swarmClusterUpdatedAt", info.Swarm.Cluster.UpdatedAt.Unix())
			lib.SetMetric(metricSet, "swarmClusterID", info.Swarm.Cluster.ID)
			lib.SetMetric(metricSet, "swarmClusterVersionIndex", info.Swarm.Cluster.Version.Index)
		}

	} else {
		lib.ErrorLogToInsights(err, entity)
	}
}
