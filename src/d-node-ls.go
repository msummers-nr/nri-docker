package main

import (
	"context"
	"fmt"
	"sort"

	"github.com/docker/docker/api/types"
	"vbom.ml/util/sortorder"
)

func getNodes() {
	ctx := context.Background()

	nodes, err := cli.NodeList(ctx, types.NodeListOptions{})
	if err != nil {
		errorLogToInsights(err)
	} else {

		sort.Slice(nodes, func(i, j int) bool {
			return sortorder.NaturalLess(nodes[i].Description.Hostname, nodes[j].Description.Hostname)
		})

		for _, node := range nodes {
			metricSet := newMetricSet("dockerNodeSample")
			setMetric(metricSet, "nodeID", node.ID)
			setMetric(metricSet, "message", node.Status.Message)
			setMetric(metricSet, "state", fmt.Sprintf("%v", node.Status.State))
			setMetric(metricSet, "descHostname", node.Description.Hostname)
			setMetric(metricSet, "engineVersion", node.Description.Engine.EngineVersion)
			setMetric(metricSet, "platformArch", node.Description.Platform.Architecture)
			setMetric(metricSet, "platformOS", node.Description.Platform.OS)
			setMetric(metricSet, "createdAt", node.CreatedAt.Unix())
			setMetric(metricSet, "duration", makeTimestamp()-node.CreatedAt.Unix())
			setMetric(metricSet, "updatedAt", node.UpdatedAt.Unix())
			if node.ManagerStatus != nil {
				setMetric(metricSet, "managerStatusLeader", node.ManagerStatus.Leader)
				setMetric(metricSet, "managerStatusReachability", fmt.Sprintf("%v", node.ManagerStatus.Reachability))
				setMetric(metricSet, "managerStatusAddr", node.ManagerStatus.Addr)
			}
			setMetric(metricSet, "availability", fmt.Sprintf("%v", node.Spec.Availability))
			setMetric(metricSet, "name", fmt.Sprintf("%v", node.Spec.Name))
			setMetric(metricSet, "role", fmt.Sprintf("%v", node.Spec.Role))
			setMetric(metricSet, "annotationsName", node.Spec.Annotations.Name)

			for key, val := range node.Spec.Annotations.Labels {
				setMetric(metricSet, key, val)
			}
			for key, val := range node.Spec.Labels {
				setMetric(metricSet, key, val)
			}
		}
	}
}
