package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/newrelic/infra-integrations-sdk/log"
)

func getTasks() {
	ctx := context.Background()
	tasks, err := cli.TaskList(ctx, types.TaskListOptions{})
	if err != nil {
		log.Debug(err.Error())
	} else {
		for _, task := range tasks {
			metricSet := newMetricSet("dockerTaskSample")
			setMetric(metricSet, "taskID", task.ID)
			setMetric(metricSet, "nodeID", task.NodeID)
			setMetric(metricSet, "serviceID", task.ServiceID)
			setMetric(metricSet, "image", task.Spec.ContainerSpec.Image)
			img := strings.Split(task.Spec.ContainerSpec.Image, "@")
			setMetric(metricSet, "imageShort", img[0])
			setMetric(metricSet, "desiredState", task.DesiredState)
			setMetric(metricSet, "createdAt", task.CreatedAt.Unix())
			setMetric(metricSet, "duration", makeTimestamp()-task.CreatedAt.Unix())
			setMetric(metricSet, "updatedAt", task.UpdatedAt.Unix())
			setMetric(metricSet, "name", task.Name)
			setMetric(metricSet, "annotationsName", task.Annotations.Name)
			setMetric(metricSet, "versionIndex", task.Version.Index)
			setMetric(metricSet, "containerID", task.Status.ContainerStatus.ContainerID)
			setMetric(metricSet, "containerExitCode", task.Status.ContainerStatus.ExitCode)
			setMetric(metricSet, "containerPID", task.Status.ContainerStatus.PID)
			setMetric(metricSet, "error", task.Status.Err)
			setMetric(metricSet, "message", task.Status.Message)
			setMetric(metricSet, "state", fmt.Sprintf("%v", task.Status.State))
			setMetric(metricSet, "statusTimestamp", task.Status.Timestamp.Unix())
			setMetric(metricSet, "desiredState", fmt.Sprintf("%v", task.DesiredState))

			if task.Spec.ContainerSpec.Healthcheck != nil {
				setMetric(metricSet, "healthcheckRetries", task.Spec.ContainerSpec.Healthcheck.Retries)
			}

			for key, val := range task.Spec.ContainerSpec.Labels {
				setMetric(metricSet, key, val)
			}
			for key, val := range task.Annotations.Labels {
				setMetric(metricSet, key, val)
			}
			for key, val := range task.Labels {
				setMetric(metricSet, key, val)
			}
		}
	}
}
