package main

import (
	"context"
	"fmt"
	"sort"

	"github.com/docker/cli/cli/command/formatter"
	"github.com/docker/cli/cli/compose/convert"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"vbom.ml/util/sortorder"
)

func getServices() {
	ctx := context.Background()
	services, err := cli.ServiceList(ctx, types.ServiceListOptions{})

	if err == nil {
		sort.Slice(services, func(i, j int) bool {
			return sortorder.NaturalLess(services[i].Spec.Name, services[j].Spec.Name)
		})
		if len(services) > 0 {
			// only non-empty services, should we call TaskList and NodeList api
			taskFilter := filters.NewArgs()
			for _, service := range services {
				taskFilter.Add("service", service.ID)
			}
			tasks, err := cli.TaskList(ctx, types.TaskListOptions{Filters: taskFilter})
			if err != nil {
				errorLogToInsights(err)
			}
			nodes, err := cli.NodeList(ctx, types.NodeListOptions{})
			if err != nil {
				errorLogToInsights(err)
			}
			if err == nil {
				GetServicesStatus(services, nodes, tasks)
			}
		}
	}
}

// GetServicesStatus x
func GetServicesStatus(services []swarm.Service, nodes []swarm.Node, tasks []swarm.Task) {
	running := map[string]int{}
	tasksNoShutdown := map[string]int{}

	activeNodes := make(map[string]struct{})
	for _, n := range nodes {
		if n.Status.State != swarm.NodeStateDown {
			activeNodes[n.ID] = struct{}{}
		}
	}

	for _, task := range tasks {
		if task.DesiredState != swarm.TaskStateShutdown {
			tasksNoShutdown[task.ServiceID]++
		}
		if _, nodeActive := activeNodes[task.NodeID]; nodeActive && task.Status.State == swarm.TaskStateRunning {
			running[task.ServiceID]++
		}
	}

	m := make(map[string]*formatter.Stack)
	for _, service := range services {
		metricSet := newMetricSet("dockerServiceSample")
		setMetric(metricSet, "serviceID", service.ID)
		setMetric(metricSet, "name", service.Spec.Name)
		setMetric(metricSet, "createdAt", service.CreatedAt.Unix())
		setMetric(metricSet, "duration", makeTimestamp()-service.CreatedAt.Unix())
		setMetric(metricSet, "updatedAt", service.UpdatedAt.Unix())

		setMetric(metricSet, "endpointMode", fmt.Sprintf("%v", service.Endpoint.Spec.Mode))

		if service.UpdateStatus != nil {
			setMetric(metricSet, "updateStatusCompletedAt", service.UpdateStatus.CompletedAt.Unix())
			setMetric(metricSet, "updateStatusStartedAt", service.UpdateStatus.StartedAt.Unix())
			setMetric(metricSet, "updateStatusMessage", fmt.Sprintf("%v", service.UpdateStatus.Message))
			setMetric(metricSet, "updateStatusState", fmt.Sprintf("%v", service.UpdateStatus.State))
		}

		setMetric(metricSet, "versionIndex", service.Version.Index)

		if service.Spec.Mode.Replicated != nil && service.Spec.Mode.Replicated.Replicas != nil {
			setMetric(metricSet, "mode", "replicated")
			setMetric(metricSet, "replicasCurrent", running[service.ID])
			setMetric(metricSet, "replicasExpected", *service.Spec.Mode.Replicated.Replicas)
		} else if service.Spec.Mode.Global != nil {
			setMetric(metricSet, "mode", "global")
			setMetric(metricSet, "replicasCurrent", running[service.ID])
			setMetric(metricSet, "replicasExpected", tasksNoShutdown[service.ID])
		}

		setMetric(metricSet, "annotationsName", service.Spec.Annotations.Name)
		for key, val := range service.Spec.Annotations.Labels {
			setMetric(metricSet, key, val)
		}

		labels := service.Spec.Labels
		for key, val := range service.Spec.Labels {
			setMetric(metricSet, key, val)
		}

		name, ok := labels[convert.LabelNamespace]
		if ok {
			ztack, ok := m[name]
			if !ok {
				m[name] = &formatter.Stack{
					Name:     name,
					Services: 1,
				}
			} else {
				ztack.Services++
			}
		}
	}

	for stack, val := range m {
		metricSet, _ := entity.NewMetricSet("dockerStackSample")
		setMetric(metricSet, "name", stack)
		setMetric(metricSet, "services", val.Services)
	}
}
