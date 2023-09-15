package nrdocker

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/newrelic-experimental/nri-docker/internal/lib"
	"github.com/newrelic/infra-integrations-sdk/integration"
	log "github.com/newrelic/infra-integrations-sdk/log"
)

// GetTasks x
func GetTasks(cli *client.Client, entity *integration.Entity) {
	ctx := context.Background()
	tasks, err := cli.TaskList(ctx, types.TaskListOptions{})
	if err != nil {
		log.Debug(err.Error())
	} else {
		for _, task := range tasks {
			metricSet := lib.NewMetricSet("dockerTaskSample", entity)
			lib.SetMetric(metricSet, "taskID", task.ID)
			lib.SetMetric(metricSet, "nodeID", task.NodeID)
			lib.SetMetric(metricSet, "serviceID", task.ServiceID)
			lib.SetMetric(metricSet, "desiredState", task.DesiredState)
			lib.SetMetric(metricSet, "createdAt", task.CreatedAt.Unix())
			lib.SetMetric(metricSet, "duration", lib.MakeTimestamp()-task.CreatedAt.Unix())
			lib.SetMetric(metricSet, "updatedAt", task.UpdatedAt.Unix())
			lib.SetMetric(metricSet, "name", task.Name)
			lib.SetMetric(metricSet, "annotationsName", task.Annotations.Name)
			lib.SetMetric(metricSet, "versionIndex", task.Version.Index)
			if task.Status.ContainerStatus != nil {
				lib.SetMetric(metricSet, "containerID", task.Status.ContainerStatus.ContainerID)
				lib.SetMetric(metricSet, "containerExitCode", task.Status.ContainerStatus.ExitCode)
				lib.SetMetric(metricSet, "containerPID", task.Status.ContainerStatus.PID)
			}
			lib.SetMetric(metricSet, "error", task.Status.Err)
			lib.SetMetric(metricSet, "message", task.Status.Message)
			lib.SetMetric(metricSet, "state", fmt.Sprintf("%v", task.Status.State))
			lib.SetMetric(metricSet, "statusTimestamp", task.Status.Timestamp.Unix())
			lib.SetMetric(metricSet, "desiredState", fmt.Sprintf("%v", task.DesiredState))

			if task.Spec.ContainerSpec != nil {
				lib.SetMetric(metricSet, "image", task.Spec.ContainerSpec.Image)
				img := strings.Split(task.Spec.ContainerSpec.Image, "@")
				lib.SetMetric(metricSet, "imageShort", img[0])
				if task.Spec.ContainerSpec.Healthcheck != nil {
					lib.SetMetric(metricSet, "healthcheckRetries", task.Spec.ContainerSpec.Healthcheck.Retries)
				}
			}

			for key, val := range task.Spec.ContainerSpec.Labels {
				lib.SetMetric(metricSet, key, val)
			}
			for key, val := range task.Annotations.Labels {
				lib.SetMetric(metricSet, key, val)
			}
			for key, val := range task.Labels {
				lib.SetMetric(metricSet, key, val)
			}
		}
	}
}
