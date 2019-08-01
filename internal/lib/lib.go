package lib

import (
	"strconv"
	"strings"
	"time"

	sdkArgs "github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/data/metric"
	"github.com/newrelic/infra-integrations-sdk/integration"
)

var IntegrationName = "com.newrelic.nri-docker"
var IntegrationVersion = "Unknown-SNAPSHOT"
var Hostname = ""
var SwarmState = "inactive"

type ArgumentList struct {
	sdkArgs.DefaultArgumentList
	Local      bool   `default:"true" help:"Collect local entity info (merges host metadata into event sample)"`
	Exclude    string `default:"true" help:"Comma separated list to filter out unneeded metrics"`
	APIVersion string `default:"" help:"Force integrations client API version"`
}

var Args ArgumentList

// SetMetric x
func SetMetric(metricSet *metric.Set, key string, val interface{}) {
	if checkExclusions(key, Args) == true {
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

func checkExclusions(key string, args ArgumentList) bool {
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

// NewMetricSet x
func NewMetricSet(event string, entity *integration.Entity) *metric.Set {
	metricSet := entity.NewMetricSet(event)
	SetMetric(metricSet, "integration_version", IntegrationVersion)
	return metricSet
}

func createKeyValuePairs(ds map[string]interface{}, m map[string]string) {
	for key, value := range m {
		ds[key] = value
	}
}

// ErrorLogToInsights x
func ErrorLogToInsights(err error, entity *integration.Entity) {
	errorMetricSet := entity.NewMetricSet("dockerIntegrationError")
	errorMetricSet.SetMetric("errorMsg", err.Error(), metric.ATTRIBUTE)
	errorMetricSet.SetMetric("hostname", Hostname, metric.ATTRIBUTE)
}

func PanicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}

func MakeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func ApplyLabel(label string, metricSet *metric.Set, customKey string) {
	labelSplit := strings.SplitN(label, "=", 2)
	if len(labelSplit) == 2 {
		if labelSplit[0] != "" && labelSplit[1] != "" {
			key := labelSplit[0]
			if customKey != "" {
				key = customKey
			}
			SetMetric(metricSet, key, labelSplit[1])
		}
	}
}
