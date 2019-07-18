package nrdocker

import (
	"context"
	"fmt"
	"sort"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/newrelic/infra-integrations-sdk/integration"
	"source.datanerd.us/FIT/nri-docker/internal/lib"
	"vbom.ml/util/sortorder"
	// "vbom.ml/util/sortorder"
)

// GetNodes x
func GetNodes(cli *client.Client, entity *integration.Entity) {
	ctx := context.Background()

	nodes, err := cli.NodeList(ctx, types.NodeListOptions{})
	if err != nil {
		lib.ErrorLogToInsights(err, entity)
	} else {

		sort.Slice(nodes, func(i, j int) bool {
			return sortorder.NaturalLess(nodes[i].Description.Hostname, nodes[j].Description.Hostname)
		})

		for _, node := range nodes {
			metricSet := lib.NewMetricSet("dockerNodeSample", entity)
			lib.SetMetric(metricSet, "nodeID", node.ID)
			lib.SetMetric(metricSet, "message", node.Status.Message)
			lib.SetMetric(metricSet, "state", fmt.Sprintf("%v", node.Status.State))
			lib.SetMetric(metricSet, "descHostname", node.Description.Hostname)
			lib.SetMetric(metricSet, "engineVersion", node.Description.Engine.EngineVersion)
			lib.SetMetric(metricSet, "platformArch", node.Description.Platform.Architecture)
			lib.SetMetric(metricSet, "platformOS", node.Description.Platform.OS)
			lib.SetMetric(metricSet, "createdAt", node.CreatedAt.Unix())
			lib.SetMetric(metricSet, "duration", lib.MakeTimestamp()-node.CreatedAt.Unix())
			lib.SetMetric(metricSet, "updatedAt", node.UpdatedAt.Unix())
			if node.ManagerStatus != nil {
				lib.SetMetric(metricSet, "managerStatusLeader", node.ManagerStatus.Leader)
				lib.SetMetric(metricSet, "managerStatusReachability", fmt.Sprintf("%v", node.ManagerStatus.Reachability))
				lib.SetMetric(metricSet, "managerStatusAddr", node.ManagerStatus.Addr)
			}
			lib.SetMetric(metricSet, "availability", fmt.Sprintf("%v", node.Spec.Availability))
			lib.SetMetric(metricSet, "name", fmt.Sprintf("%v", node.Spec.Name))
			lib.SetMetric(metricSet, "role", fmt.Sprintf("%v", node.Spec.Role))
			lib.SetMetric(metricSet, "annotationsName", node.Spec.Annotations.Name)

			for key, val := range node.Spec.Annotations.Labels {
				lib.SetMetric(metricSet, key, val)
			}
			for key, val := range node.Spec.Labels {
				lib.SetMetric(metricSet, key, val)
			}
		}
	}
}
