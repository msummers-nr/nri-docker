package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	yaml "gopkg.in/yaml.v2"

	sdkArgs "github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/data/metric"
	"github.com/newrelic/infra-integrations-sdk/integration"
	"github.com/newrelic/infra-integrations-sdk/log"
)

type argumentList struct {
	sdkArgs.DefaultArgumentList
}

const (
	integrationName    = "com.nr.docker-ohi"
	integrationVersion = "1.0.0"
)

var args argumentList

//DockerData struct
type DockerData struct {
	DockerHost                        string  `json:",omitempty"`
	ContainerID, ContainerNetworkMode string  `json:",omitempty"`
	ContainerImage, ContainerImageID  string  `json:",omitempty"`
	ContainerState, ContainerStatus   string  `json:",omitempty"`
	ContainerCommand                  string  `json:",omitempty"`
	ContainerPorts, ContainerMounts   string  `json:",omitempty"`
	ContainerCreated                  int64   `json:",omitempty"`
	ContainerLabels                   string  `json:",omitempty"`
	ContainerSizeRootFs               int64   `json:",omitempty"`
	ContainerSizeRw                   int64   `json:",omitempty"`
	MemPercent, CPUPercent            float64 `json:",omitempty"`
	BlkRead, BlkWrite                 uint64  `json:",omitempty"`
	Mem, MemLimit                     float64 `json:",omitempty"`
	PidsStatsCurrent, PidsStatsLimit  float64 `json:",omitempty"`
	PreviousCPU                       uint64  `json:",omitempty"`
	PreviousSystem                    uint64  `json:",omitempty"`
	NetRx, NetTx                      float64 `json:",omitempty"`
	NetRxErrors, NetTxErrors          float64 `json:",omitempty"`
	NetRxDropped, NetTxDropped        float64 `json:",omitempty"`
	NetRxPackets, NetTxPackets        float64 `json:",omitempty"`
	Periods                           float64 `json:",omitempty"`
	ThrottledPeriods, ThrottledTime   float64 `json:",omitempty"`
	NumProcs                          uint32  `json:",omitempty"`
}

// Config YAML Struct
type Config struct {
	API     string
	Exclude string
}

var file = []byte{}
var err error
var yml string
var c Config

func integrationWithLocalEntity(i *integration.Integration) {
	hostName, _ := os.Hostname()
	entity, _ := i.Entity(hostName, ("docker-ohi"))
	panicOnErr(err)

	getDockerStats(entity)
}

func integrationWithRemoteEntities(i *integration.Integration) {}

func getDockerStats(entity *integration.Entity) {

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.WithVersion(c.API))
	panicOnErr(err)

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	panicOnErr(err)

	log.Debug("GOOS:", runtime.GOOS)

	for _, container := range containers {
		fetchStats(ctx, cli, entity, container)
	}
}

func fetchStats(ctx context.Context, cli *client.Client, entity *integration.Entity, container types.Container) {
	var DD DockerData
	// Place into map string interface to loop through easier, and avoid using reflect to loop through keys in a struct
	dockerStats := map[string]interface{}{}

	dockerStats["Hostname"], _ = os.Hostname()
	c, _ := cpu.Percent(0, false)
	dockerStats["HostCPUPerc"] = c[0]
	m, _ := mem.VirtualMemory()
	dockerStats["HostMemTotal"] = m.Total
	dockerStats["HostMemFree"] = m.Free
	dockerStats["HostMemUsedPerc"] = m.UsedPercent

	stats, _ := cli.ContainerStats(ctx, container.ID, false)
	var containerStats types.StatsJSON
	json.NewDecoder(stats.Body).Decode(&containerStats)

	dockerStats["ID"] = container.ID
	dockerStats["IDShort"] = container.ID[0:12]
	dockerStats["Image"] = container.Image
	dockerStats["ImageID"] = container.ImageID
	dockerStats["State"] = container.State
	dockerStats["Status"] = container.Status
	dockerStats["Command"] = container.Command
	dockerStats["Created"] = container.Created
	// hsingh:
	// Multiple labels returned will be treated as seperate attributes
	// dockerStats["Labels"] =  <== Remove attribute named label
	createKeyValuePairs(dockerStats, container.Labels)
	dockerStats["NetworkMode"] = container.HostConfig.NetworkMode
	dockerStats["SizeRw"] = container.SizeRw
	dockerStats["SizeRootFs"] = container.SizeRootFs

	b := new(bytes.Buffer)
	fmt.Fprintf(b, "%v,", container.Ports)
	dockerStats["Ports"] = b.String()

	b = new(bytes.Buffer)
	fmt.Fprintf(b, "%v,", container.Mounts)
	dockerStats["Mounts"] = b.String()

	dockerStats["NetRx"], dockerStats["NetTx"],
		dockerStats["NetRxErrors"], dockerStats["NetTxErrors"],
		dockerStats["NetRxDropped"], dockerStats["NetTxDropped"],
		dockerStats["NetRxPackets"], dockerStats["NetTxPackets"] = calculateNetwork(containerStats.Networks)

	if stats.OSType == "windows" {
		dockerStats["CPUPercent"] = calculateCPUPercentWindows(containerStats)
		dockerStats["BlkRead"] = containerStats.StorageStats.ReadSizeBytes
		dockerStats["BlkWrite"] = containerStats.StorageStats.WriteSizeBytes
		dockerStats["Mem"] = float64(containerStats.MemoryStats.PrivateWorkingSet)
	} else {
		dockerStats["PreviousCPU"] = containerStats.PreCPUStats.CPUUsage.TotalUsage
		dockerStats["PreviousSystem"] = containerStats.PreCPUStats.SystemUsage
		dockerStats["CPUPercent"] = calculateCPUPercentUnix(containerStats.PreCPUStats.CPUUsage.TotalUsage, containerStats.PreCPUStats.SystemUsage, containerStats)
		dockerStats["BlkRead"], dockerStats["BlkWrite"] = calculateBlockIO(containerStats.BlkioStats)
		DD.Mem = calculateMemUsageUnixNoCache(containerStats.MemoryStats)
		dockerStats["Mem"] = DD.Mem
		DD.MemLimit = float64(containerStats.MemoryStats.Limit)
		dockerStats["MemLimit"] = DD.MemLimit
		dockerStats["MemPercent"] = calculateMemPercentUnixNoCache(DD.MemLimit, DD.Mem)
	}

	dockerStats["NumProcs"] = containerStats.NumProcs
	dockerStats["PidsStatsCurrent"] = float64(containerStats.PidsStats.Current)
	dockerStats["PidsStatsLimit"] = float64(containerStats.PidsStats.Limit)
	dockerStats["Periods"] = float64(containerStats.CPUStats.ThrottlingData.Periods)
	dockerStats["ThrottledPeriods"] = float64(containerStats.CPUStats.ThrottlingData.ThrottledPeriods)
	dockerStats["ThrottledTime"] = float64(containerStats.CPUStats.ThrottlingData.ThrottledTime)

	metricSet, _ := entity.NewMetricSet("dockerContainerEvent")
	for key, val := range dockerStats {
		if checkExclusions(key) == true {
			switch val.(type) {
			case float64:
				metricSet.SetMetric(key, val, metric.GAUGE)
			case uint64:
				metricSet.SetMetric(key, val, metric.GAUGE)
			case int64:
				metricSet.SetMetric(key, val, metric.GAUGE)
			case uint32:
				metricSet.SetMetric(key, val, metric.GAUGE)
			case string:
				metricSet.SetMetric(key, val, metric.ATTRIBUTE)
			}
		}
	}
}

func main() {
	remoteEntities := false
	i, err := integration.New(integrationName, integrationVersion, integration.Args(&args))
	panicOnErr(err)

	//Load yml config file
	file, err = ioutil.ReadFile("docker-config.yml")
	if err != nil {
		log.Fatal(err)
	}
	yml = string(file)

	//Unmarshal yml config
	c = Config{}
	err = yaml.Unmarshal([]byte(yml), &c)
	if err != nil {
		log.Fatal(err)
	}

	if remoteEntities {
		integrationWithRemoteEntities(i)
	} else {
		integrationWithLocalEntity(i)
	}

	panicOnErr(i.Publish())
}

func checkExclusions(key string) bool {
	excludes := strings.Split(c.Exclude, ",")
	passed := true
	for _, exclude := range excludes {
		if strings.Contains(key, exclude) == true {
			passed = false
			break
		}
	}
	return passed
}

func calculateCPUPercentUnix(previousCPU, previousSystem uint64, v types.StatsJSON) float64 {
	var (
		cpuPercent = 0.0
		// calculate the change for the cpu usage of the container in between readings
		cpuDelta = float64(v.CPUStats.CPUUsage.TotalUsage) - float64(previousCPU)
		// calculate the change for the entire system between readings
		systemDelta = float64(v.CPUStats.SystemUsage) - float64(previousSystem)
	)

	if systemDelta > 0.0 && cpuDelta > 0.0 {
		cpuPercent = (cpuDelta / systemDelta) * float64(len(v.CPUStats.CPUUsage.PercpuUsage)) * 100.0
	}
	return cpuPercent
}

func calculateCPUPercentWindows(v types.StatsJSON) float64 {
	// Max number of 100ns intervals between the previous time read and now
	possIntervals := uint64(v.Read.Sub(v.PreRead).Nanoseconds()) // Start with number of ns intervals
	possIntervals /= 100                                         // Convert to number of 100ns intervals
	possIntervals *= uint64(v.NumProcs)                          // Multiple by the number of processors

	// Intervals used
	intervalsUsed := v.CPUStats.CPUUsage.TotalUsage - v.PreCPUStats.CPUUsage.TotalUsage

	// Percentage avoiding divide-by-zero
	if possIntervals > 0 {
		return float64(intervalsUsed) / float64(possIntervals) * 100.0
	}
	return 0.00
}

func calculateBlockIO(blkio types.BlkioStats) (uint64, uint64) {
	var blkRead, blkWrite uint64
	for _, bioEntry := range blkio.IoServiceBytesRecursive {
		switch strings.ToLower(bioEntry.Op) {
		case "read":
			blkRead = blkRead + bioEntry.Value
		case "write":
			blkWrite = blkWrite + bioEntry.Value
		}
	}
	return blkRead, blkWrite
}

func calculateNetwork(network map[string]types.NetworkStats) (float64, float64, float64, float64, float64, float64, float64, float64) {
	var rx, tx, rxErrors, txErrors, rxDropped, txDropped, rxPackets, txPackets float64
	for _, v := range network {
		rx += float64(v.RxBytes)
		tx += float64(v.TxBytes)
		rxPackets += float64(v.RxPackets)
		txPackets += float64(v.TxPackets)
		rxErrors += float64(v.RxErrors)
		txErrors += float64(v.TxErrors)
		rxDropped += float64(v.RxDropped)
		txDropped += float64(v.TxDropped)
	}
	return rx, tx, rxErrors, txErrors, rxDropped, txDropped, rxPackets, txPackets
}

// calculateMemUsageUnixNoCache calculate memory usage of the container.
// Page cache is intentionally excluded to avoid misinterpretation of the output.
func calculateMemUsageUnixNoCache(mem types.MemoryStats) float64 {
	return float64(mem.Usage - mem.Stats["cache"])
}

func calculateMemPercentUnixNoCache(limit float64, usedNoCache float64) float64 {
	// MemoryStats.Limit will never be 0 unless the container is not running and we haven't
	// got any data from cgroup
	if limit != 0 {
		return usedNoCache / limit * 100.0
	}
	return 0
}

//hsingh
//Updating to parse multiple labels as attributes
func createKeyValuePairs(ds map[string]interface{}, m map[string]string) {
	for key, value := range m {
		ds[key] = value
	}
}

func errorLogToInsights(entity *integration.Entity, err error, hostLabel string) {
	errorMetricSet, _ := entity.NewMetricSet("dockerEventError")
	errorMetricSet.SetMetric("errorMsg", err.Error(), metric.ATTRIBUTE)
	errorMetricSet.SetMetric("hostLabel", hostLabel, metric.ATTRIBUTE)
}

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}
