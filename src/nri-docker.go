package main

import (
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api"
	"github.com/docker/docker/client"

	sdkArgs "github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/data/metric"
	"github.com/newrelic/infra-integrations-sdk/integration"
	"github.com/newrelic/infra-integrations-sdk/log"
)

type argumentList struct {
	sdkArgs.DefaultArgumentList
	Local      bool   `default:"true" help:"Collect local entity info (merges host metadata into event sample)"`
	Exclude    string `default:"true" help:"Comma separated list to filter out unneeded metrics"`
	APIVersion string `default:"" help:"Force integrations client API version"`
}

const (
	integrationName    = "com.newrelic.nri-docker"
	integrationVersion = "2.0.0"
)

var args argumentList

var file = []byte{}
var err error

func main() {
	i, err := integration.New(integrationName, integrationVersion, integration.Args(&args))
	panicOnErr(err)
	integrationWithLocalEntity(i)
	panicOnErr(i.Publish())
}

var hostname string
var entity *integration.Entity
var cli *client.Client

func integrationWithLocalEntity(i *integration.Integration) {
	hostname, _ = os.Hostname()

	if args.Local == true {
		entity = i.LocalEntity()
	} else {
		entity, _ = i.Entity(hostname, "nri-docker")
	}

	panicOnErr(err)

	cli, err = setDockerClient()
	if err != nil {
		log.Fatal(err)
	}

	getHostInfo()
	getServices()
	getTasks()
	getContainerInfo()
}

// setDockerClient - Required as there can be edge cases when the integration API version, may need a matching or lower API version then the hosts docker API version
func setDockerClient() (*client.Client, error) {
	var out []byte
	var cli *client.Client

	if args.APIVersion != "" {
		cli, err = client.NewClientWithOpts(client.WithVersion(args.APIVersion))
	} else {
		log.Debug("GOOS:", runtime.GOOS)

		if err != nil {
			if runtime.GOOS == "windows" {
				out, err = exec.Command("cmd", "/C", `docker`, `version`, `--format`, `"{{json .Client.APIVersion}}"`).Output()
			} else {
				out, err = exec.Command(`docker`, `version`, `--format`, `"{{json .Client.APIVersion}}"`).Output()
				if err != nil {
					out, err = exec.Command(`/host/usr/local/bin/docker`, `version`, `--format`, `"{{json .Client.APIVersion}}"`).Output()
				}
			}
		}

		if err != nil {
			log.Debug("Unable to fetch Docker API version", err)
			log.Debug("Setting client with NewEnvClient()")
			err = nil
			cli, err = client.NewEnvClient()
		} else {
			cmdOut := string(out)
			clientAPIVersion := strings.TrimSpace(strings.Replace(cmdOut, `"`, "", -1))
			clientVer, _ := strconv.ParseFloat(clientAPIVersion, 64)
			apiVer, _ := strconv.ParseFloat(api.DefaultVersion, 64)

			if clientVer <= apiVer {
				log.Debug("Setting client with version:", clientAPIVersion)
				cli, err = client.NewClientWithOpts(client.WithVersion(clientAPIVersion))
			} else {
				log.Debug("Client API Version", clientAPIVersion, "is higher then integration version", api.DefaultVersion)
				log.Debug("Setting client with NewEnvClient()")
				cli, err = client.NewEnvClient()
			}
		}
	}

	return cli, err
}

func newMetricSet(event string) *metric.Set {
	metricSet, _ := entity.NewMetricSet(event)
	setMetric(metricSet, "integration_version", integrationVersion)
	return metricSet
}

func setMetric(metricSet *metric.Set, key string, val interface{}) {
	if checkExclusions(key) == true {
		switch val.(type) {
		case float64:
			metricSet.SetMetric(key, val, metric.GAUGE)
		case uint16:
			metricSet.SetMetric(key, val, metric.GAUGE)
		case uint32:
			metricSet.SetMetric(key, val, metric.GAUGE)
		case uint64:
			metricSet.SetMetric(key, val, metric.GAUGE)
		case int:
			metricSet.SetMetric(key, val, metric.GAUGE)
		case int32:
			metricSet.SetMetric(key, val, metric.GAUGE)
		case int64:
			metricSet.SetMetric(key, val, metric.GAUGE)
		case bool:
			metricSet.SetMetric(key, strconv.FormatBool(val.(bool)), metric.ATTRIBUTE)
		case string:
			if val.(string) != "" {
				metricSet.SetMetric(key, val, metric.ATTRIBUTE)
			}
		}
	}
}

func checkExclusions(key string) bool {
	excludes := strings.Split(args.Exclude, ",")
	passed := true
	for _, exclude := range excludes {
		if strings.Contains(key, exclude) == true {
			passed = false
			break
		}
	}
	return passed
}

func createKeyValuePairs(ds map[string]interface{}, m map[string]string) {
	for key, value := range m {
		ds[key] = value
	}
}

func errorLogToInsights(entity *integration.Entity, err error) {
	errorMetricSet, _ := entity.NewMetricSet("dockerIntegrationError")
	errorMetricSet.SetMetric("errorMsg", err.Error(), metric.ATTRIBUTE)
	errorMetricSet.SetMetric("hostname", hostname, metric.ATTRIBUTE)
}

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}

func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
