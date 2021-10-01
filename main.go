package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/spf13/viper"
)

func main() {
	initConfig()

	pipelines := viper.GetStringSlice("pipelines")
	for _, pipeline := range pipelines {
		getJenkinsPipelineData(pipeline)
	}
}

// Gets the existing pipelines and their build IDs currently in InfluxDB
// This is used in getJenkinsPipelineData to ensure that existing records are not inserted again
func getInfluxDBExistingBuilds() []ExistingPipeline {
	influxDBBaseURL := viper.Get("influxDBBaseURL").(string)
	influxDBToken := viper.Get("influxDBToken").(string)
	influxDBOrg := viper.Get("influxDBOrg").(string)
	influxDBBucket := viper.Get("influxDBBucket").(string)

	var existingPipelines []ExistingPipeline

	client := influxdb2.NewClient(influxDBBaseURL, influxDBToken)
	queryAPI := client.QueryAPI(influxDBOrg)

	result, err := queryAPI.Query(context.Background(), `
		import "influxdata/influxdb/schema"

		schema.measurementTagValues(
		bucket: "`+influxDBBucket+`",
		measurement: "pipelineData",
		tag: "pipeline"
		)
	`)

	if err == nil {
		for result.Next() {
			record := result.Record().String()
			pipeline_name := extractValue(record)
			existingPipelines = append(existingPipelines, ExistingPipeline{PipelineName: pipeline_name})
		}
		if result.Err() != nil {
			fmt.Printf("Query error: %s\n", result.Err().Error())
		}
	}

	for i, pipeline := range existingPipelines {
		builds_result, build_err := queryAPI.Query(context.Background(), `
			from(bucket: "`+influxDBBucket+`")
				|> range(start: 2021-09-24, stop: now())
				|> filter(fn: (r) => r._measurement == "pipelineData")
				|> filter(fn: (r) => r.pipeline == "`+pipeline.PipelineName+`")
				|> keyValues(keyColumns: ["executionId"])
				|> group()
				|> keep(columns: ["executionId"])
				|> distinct(column: "executionId")
		`)
		if build_err == nil {
			for builds_result.Next() {
				record := builds_result.Record().String()
				buildID := extractValue(record)
				buidlIDInt, _ := strconv.Atoi(buildID)
				existingPipelines[i].BuildIDs = append(existingPipelines[i].BuildIDs, buidlIDInt)
			}
			if builds_result.Err() != nil {
				fmt.Printf("Query error: %s\n", builds_result.Err().Error())
			}
		}
	}
	client.Close()
	return existingPipelines
}

func getJenkinsPipelineData(pipeline string) {
	jenkinsBaseURL := viper.Get("jenkinsBaseURL").(string)
	jenkinsUser := viper.Get("jenkinsUser").(string)
	jenkinsToken := viper.Get("jenkinsToken").(string)

	method := "GET"
	url := jenkinsBaseURL + "/job/" + pipeline + "/wfapi/runs?fullStages=true"
	payload := strings.NewReader(``)

	plainCredentials := jenkinsUser + ":" + jenkinsToken
	base64Credentials := base64.StdEncoding.EncodeToString([]byte(plainCredentials))

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}

	req.Header.Add("Authorization", "Basic "+base64Credentials)
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	var pipelineData PipelineData
	json.Unmarshal(body, &pipelineData)

	for _, execution := range pipelineData {
		pipelineName := pipeline
		order := 1
		if !existingPipelineInfo(pipelineName, execution.ID) {
			for _, stage := range execution.Stages {
				escapedStageName := strings.ReplaceAll(stage.Name, " ", "\\ ")
				escapedPipelineName := strings.ReplaceAll(pipelineName, " ", "\\ ")
				fmt.Printf("pipelineData,pipeline=%s,pipeline_status=%s,executionId=%s,stage=%s,order=%d,status=%s duration=%d\n", escapedPipelineName, execution.Status, execution.ID, escapedStageName, order, stage.Status, stage.DurationMillis+1000)
				order = order + 1
			}
		}
	}
}

func extractValue(record string) string {
	pairs := strings.Split(record, ",")
	var value_pair string
	for _, pair := range pairs {
		if strings.Contains(pair, "_value") {
			value_pair = pair
			break
		}
	}

	re := regexp.MustCompile(`_value:(.+?)$`)
	test := re.FindAllStringSubmatch(value_pair, -1)
	return test[0][1]
}

func existingPipelineInfo(pipelineName string, executionID string) bool {
	executionIDInt, _ := strconv.Atoi(executionID)
	existingPipelines := getInfluxDBExistingBuilds()
	existing := false

	for _, pipelineInfo := range existingPipelines {
		if pipelineInfo.PipelineName == pipelineName {
			for _, executionIDInfo := range pipelineInfo.BuildIDs {
				if executionIDInfo == executionIDInt {
					existing = true
					break
				}
			}
		}
	}
	return existing
}

func initConfig() {
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.telegraf")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.ReadInConfig()
}
