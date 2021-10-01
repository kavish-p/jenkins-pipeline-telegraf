package main

type PipelineData []struct {
	Links               Links    `json:"_links,omitempty"`
	ID                  string   `json:"id"`
	Name                string   `json:"name"`
	Status              string   `json:"status"`
	StartTimeMillis     int64    `json:"startTimeMillis"`
	EndTimeMillis       int64    `json:"endTimeMillis"`
	DurationMillis      int      `json:"durationMillis"`
	QueueDurationMillis int      `json:"queueDurationMillis"`
	PauseDurationMillis int      `json:"pauseDurationMillis"`
	Stages              []Stages `json:"stages"`
}

type Self struct {
	Href string `json:"href"`
}

type Links struct {
	Self Self `json:"self"`
}

type Stages struct {
	Links               Links         `json:"_links"`
	ID                  string        `json:"id"`
	Name                string        `json:"name"`
	ExecNode            string        `json:"execNode"`
	Status              string        `json:"status"`
	StartTimeMillis     int64         `json:"startTimeMillis"`
	DurationMillis      int           `json:"durationMillis"`
	PauseDurationMillis int           `json:"pauseDurationMillis"`
	StageFlowNodes      []interface{} `json:"stageFlowNodes"`
}

type Changesets struct {
	Href string `json:"href"`
}

type ExistingPipeline struct {
	PipelineName string
	BuildIDs     []int
}
