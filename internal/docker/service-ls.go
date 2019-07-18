package nrdocker

import (
	"context"
	"fmt"
	"sort"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/newrelic/infra-integrations-sdk/integration"
	"source.datanerd.us/FIT/nri-docker/internal/lib"
	"vbom.ml/util/sortorder"
)

// Stack contains deployed stack information.
type Stack struct {
	// Name is the name of the stack
	Name string
	// Services is the number of the services
	Services int
	// Orchestrator is the platform where the stack is deployed
	Orchestrator string
	// Namespace is the Kubernetes namespace assigned to the stack
	Namespace string
}

const (
	// LabelNamespace is the label used to track stack resources
	LabelNamespace = "com.docker.stack.namespace"
)

// Namespace mangles names by prepending the name
type Namespace struct {
	name string
}

// GetServices x
func GetServices(cli *client.Client, entity *integration.Entity) {
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
				lib.ErrorLogToInsights(err, entity)
			}
			nodes, err := cli.NodeList(ctx, types.NodeListOptions{})
			if err != nil {
				lib.ErrorLogToInsights(err, entity)
			}
			if err == nil {
				GetServicesStatus(services, nodes, tasks, entity)
			}
		}
	}
}

// GetServicesStatus x
func GetServicesStatus(services []swarm.Service, nodes []swarm.Node, tasks []swarm.Task, entity *integration.Entity) {
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

	m := make(map[string]*Stack)
	for _, service := range services {
		metricSet := lib.NewMetricSet("dockerServiceSample", entity)
		lib.SetMetric(metricSet, "serviceID", service.ID)
		lib.SetMetric(metricSet, "name", service.Spec.Name)
		lib.SetMetric(metricSet, "createdAt", service.CreatedAt.Unix())
		lib.SetMetric(metricSet, "duration", lib.MakeTimestamp()-service.CreatedAt.Unix())
		lib.SetMetric(metricSet, "updatedAt", service.UpdatedAt.Unix())

		lib.SetMetric(metricSet, "endpointMode", fmt.Sprintf("%v", service.Endpoint.Spec.Mode))

		if service.UpdateStatus != nil {
			if service.UpdateStatus.CompletedAt != nil {
				lib.SetMetric(metricSet, "updateStatusCompletedAt", service.UpdateStatus.CompletedAt.Unix())
			}
			if service.UpdateStatus.StartedAt != nil {
				lib.SetMetric(metricSet, "updateStatusStartedAt", service.UpdateStatus.StartedAt.Unix())
			}
			lib.SetMetric(metricSet, "updateStatusMessage", fmt.Sprintf("%v", service.UpdateStatus.Message))
			lib.SetMetric(metricSet, "updateStatusState", fmt.Sprintf("%v", service.UpdateStatus.State))
		}

		lib.SetMetric(metricSet, "versionIndex", service.Version.Index)

		if service.Spec.Mode.Replicated != nil && service.Spec.Mode.Replicated.Replicas != nil {
			lib.SetMetric(metricSet, "mode", "replicated")
			lib.SetMetric(metricSet, "replicasCurrent", running[service.ID])
			lib.SetMetric(metricSet, "replicasExpected", *service.Spec.Mode.Replicated.Replicas)
		} else if service.Spec.Mode.Global != nil {
			lib.SetMetric(metricSet, "mode", "global")
			lib.SetMetric(metricSet, "replicasCurrent", running[service.ID])
			lib.SetMetric(metricSet, "replicasExpected", tasksNoShutdown[service.ID])
		}

		lib.SetMetric(metricSet, "annotationsName", service.Spec.Annotations.Name)
		for key, val := range service.Spec.Annotations.Labels {
			lib.SetMetric(metricSet, key, val)
		}

		labels := service.Spec.Labels
		for key, val := range service.Spec.Labels {
			lib.SetMetric(metricSet, key, val)
		}

		name, ok := labels[LabelNamespace]
		if ok {
			ztack, ok := m[name]
			if !ok {
				m[name] = &Stack{
					Name:     name,
					Services: 1,
				}
			} else {
				ztack.Services++
			}
		}
	}

	for stack, val := range m {
		metricSet := lib.NewMetricSet("dockerStackSample", entity)
		lib.SetMetric(metricSet, "name", stack)
		lib.SetMetric(metricSet, "services", val.Services)
	}
}
