package main

import (
	"os"

	"github.com/docker/docker/client"
	"source.datanerd.us/FIT/nri-docker/internal/docker"
	"source.datanerd.us/FIT/nri-docker/internal/lib"

	"github.com/newrelic/infra-integrations-sdk/integration"
	"github.com/newrelic/infra-integrations-sdk/log"
)

var file = []byte{}
var err error

func main() {
	i, err := integration.New(lib.IntegrationName, lib.IntegrationVersion, integration.Args(&lib.Args))
	lib.PanicOnErr(err)
	integrationWithLocalEntity(i)
	lib.PanicOnErr(i.Publish())
}

var entity *integration.Entity
var cli *client.Client

func integrationWithLocalEntity(i *integration.Integration) {
	lib.Hostname, _ = os.Hostname()

	if lib.Args.Local == true {
		entity = i.LocalEntity()
	} else {
		entity, _ = i.Entity(lib.Hostname, "nri-docker")
	}

	lib.PanicOnErr(err)

	cli, err = setDockerClient()
	if err != nil {
		log.Fatal(err)
	}

	nrdocker.GetHostInfo(cli, entity)
	nrdocker.GetContainerInfo(cli, entity)
	nrdocker.GetServices(cli, entity)

	if lib.SwarmState == "active" {
		nrdocker.GetNodes(cli, entity)
		nrdocker.GetTasks(cli, entity)
	}
}

func setDockerClient() (*client.Client, error) {
	var err error
	if lib.Args.APIVersion != "" {
		cli, err = client.NewClientWithOpts(client.WithVersion(lib.Args.APIVersion))
	} else {
		cli, err = client.NewClientWithOpts(client.WithAPIVersionNegotiation())
	}
	return cli, err
}
