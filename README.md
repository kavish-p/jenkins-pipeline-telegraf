# Jenkins Pipeline Metrics Collector for Telegraf

A simple program that collects Jenkins pipeline executions data and outputs it in InfluxDB Line Protocol format. The data includes the status and duration of the pipelines and their stages. The program is designed to be used with Telegraf.

- The program requires a config file named config.yaml located under $HOME/.telegraf of the user running the Telegraf service.
- A sample config.yaml file and Telegraf configuration has been provided in this repository.

### Usage
- Build the binary using go build
- Ensure that the config.yaml details are correct.
- Set the location of the binary in the Telegraf configuration.

### Program Flow:
1. The program first checks the existing pipelines and execution IDs in InfluxDB. The connection details and InfluxDB bucket being checked can be modified in the config.yaml file.
2. The program then uses the Jenkins API to obtain the pipeline data.
3. For each execution ID obtained from the Jenkins API, the program checks whether the data for that execution ID is already in InfluxDB. If it is not, the data is outputted in InfluxDB Line Protocol format.
4. The InfluxDB Line Protocol format will be picked up by Telegraf and pushed to InfluxDB.
