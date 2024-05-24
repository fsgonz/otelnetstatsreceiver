# Environment Log Receiver

| Status        |           |
| ------------- |-----------|
| Stability     | [beta]: logs   |


An extension of filelogreceiver which generate logs for environment. It can tail fails and also produce metrics or output them to files

## Configuration

Apart from the filelogreceiver, it can be generated to sample metrics from the environment

Extra attributes:

## Configuration

| Field          | Default | Description                                                  |
|----------------|---------|--------------------------------------------------------------|
| `log_samplers` | []      | A list of log samplers to be added to the file log receiver. |

## Log Sampler

| Field           | Default  | Description                                                                                                                                           |
|-----------------|----------|-------------------------------------------------------------------------------------------------------------------------------------------------------|
| `metric`        | Required | The metric to sample. Possible values [netstats]                                                                                                      |
| `output`        | Required | Possible Values: [file_logger, pipeline_emitter]. file_logger will output the metric to a file. pipeline_emitter will output directly to the pipeline |
| `uri`           | Optional | The uri for the output in case of a file_logger output                                                                                                |
| `poll_interval` | Optional | The uri for the output in case of a file_logger output                                                                                                |


## Examples

This will output netstats delta metrics to a file
```yaml
envlogreceiver/metering:
include:
- /tmp/files
log_samplers:
  - metric: netstats
    output: file_logger
    uri: /tmp/file.log
    poll_interval: 20s
  include_file_name: false
  poll_interval: 10s
  fingerprint_size: 1kb
  start_at: beginning
  storage: file_storage/checkpoints
  resource:
  log_type: metering_metric
  retry_on_failure:
  enabled: true
  initial_interval: 1s
  max_interval: 10m
  max_elapsed_time: 1h
```

This will output netstats directly to the pipeline
```yaml
envlogreceiver/metering:
include:
- /tmp/files
log_samplers:
  - metric: netstats
    output: pipeline_emitter
  include_file_name: false
  poll_interval: 10s
  fingerprint_size: 1kb
  start_at: beginning
  storage: file_storage/checkpoints
  resource:
  log_type: metering_metric
  retry_on_failure:
  enabled: true
  initial_interval: 1s
  max_interval: 10m
  max_elapsed_time: 1h
```
