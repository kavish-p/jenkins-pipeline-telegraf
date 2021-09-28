package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

func main() {
	getJenkinsPipelineData()
	// getInfluxDBLatestBuildNumbers()
	// fmt.Println("test")
}

func getInfluxDBLatestBuildNumbers() {
	client := influxdb2.NewClient("http://10.168.0.69:8086", "YjvGujgJCGT2O_JxMkzd59CYrQzdMJMM3YaTyjZG1xPzFnsvyzNIzX1A89nx-NO4xqDatl3fWw46jb2NuaY4bQ==")
	queryAPI := client.QueryAPI("M9")

	result, err := queryAPI.Query(context.Background(), `
		import "influxdata/influxdb/schema"

		schema.measurementTagValues(
		bucket: "app-servers",
		measurement: "cpu",
		tag: "cpu"
		)
	`)
	if err == nil {
		// Use Next() to iterate over query result lines
		for result.Next() {
			// Observe when there is new grouping key producing new table
			if result.TableChanged() {
				fmt.Printf("table: %s\n", result.TableMetadata().String())
			}
			// read result
			fmt.Printf("row: %s\n", result.Record().String())
		}
		if result.Err() != nil {
			fmt.Printf("Query error: %s\n", result.Err().Error())
		}
	}
	client.Close()
}

func getJenkinsPipelineData() {
	method := "GET"
	url := "http://10.168.0.60:8080/job/1-Deploy%20to%20Dev/wfapi/runs?fullStages=true"
	payload := strings.NewReader(``)

	plainCredentials := "admin" + ":" + "11ed270f88640f859c121f4480c6517781"
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
	// fmt.Println(string(body))

	var pipelineData PipelineData
	json.Unmarshal(body, &pipelineData)
	// fmt.Printf("API Response as struct %+v\n", pipelineData)

	fmt.Println(pipelineData[0].Stages[0].Name)
}
