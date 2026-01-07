# Docker Prometheus Exporter

This project is similar to [prometheus-podman-exporter](https://github.com/containers/prometheus-podman-exporter), but for Docker containers instead of Podman containers. It exports Docker container metrics in Prometheus format.

Grafana dashboard is available at [dashboard.json](./dashboard.json)

## Exported Metrics

The exporter provides the following metrics:

| Metric Name                                      | Type    | Description                                                                                  | Labels                                                        |
|--------------------------------------------------|---------|----------------------------------------------------------------------------------------------|---------------------------------------------------------------|
| `docker_exporter_info`                           | Gauge   | Information about the docker exporter                                                        | `version`                                                     |
| `docker_container_info`                          | Gauge   | Container information                                                                        | `container_id`, `name`, `image_id`, `command`, `network_mode` |
| `docker_container_name`                          | Gauge   | Name for the container (can be more than one)                                                | `container_id`, `name`                                        |
| `docker_container_state`                         | Gauge   | Container State (0=created, 1=running, 2=paused, 3=restarting, 4=removing, 5=exited, 6=dead) | `container_id`                                                |
| `docker_container_created_seconds`               | Gauge   | Timestamp in seconds when the container was created                                          | `container_id`                                                |
| `docker_container_started_seconds`               | Gauge   | Timestamp in seconds when the container was started                                          | `container_id`                                                |
| `docker_container_finished_at_seconds`           | Gauge   | Timestamp in seconds when the container finished                                             | `container_id`                                                |
| `docker_container_ports`                         | Gauge   | Forwarded Ports                                                                              | `container_id`, `public_port`, `private_port`, `ip`, `type`   |
| `docker_container_rootfs_size_bytes`             | Gauge   | Size of rootfs in this container in bytes                                                    | `container_id`                                                |
| `docker_container_rw_size_bytes`                 | Gauge   | Size of files that have been created or changed by this container in bytes                   | `container_id`                                                |
| `docker_container_pids`                          | Gauge   | Number of processes running in the container                                                 | `container_id`                                                |
| `docker_container_cpu_user_microseconds_total`   | Counter | Time (in microseconds) spent by tasks in user mode                                           | `container_id`                                                |
| `docker_container_cpu_kernel_microseconds_total` | Counter | Time (in microseconds) spent by tasks in kernel mode                                         | `container_id`                                                |
| `docker_container_mem_limit_kib`                 | Gauge   | Container memory limit in KiB                                                                | `container_id`                                                |
| `docker_container_mem_usage_kib`                 | Gauge   | Container memory usage in KiB                                                                | `container_id`                                                |
| `docker_container_net_send_bytes_total`          | Counter | Total number of bytes sent                                                                   | `container_id`                                                |
| `docker_container_net_send_dropped_total`        | Counter | Total number of send packet drop                                                             | `container_id`                                                |
| `docker_container_net_send_errors_total`         | Counter | Total number of send errors                                                                  | `container_id`                                                |
| `docker_container_net_receive_bytes_total`       | Counter | Total number of bytes received                                                               | `container_id`                                                |
| `docker_container_net_receive_dropped_total`     | Counter | Total number of receive packet drop                                                          | `container_id`                                                |
| `docker_container_net_receive_errors_total`      | Counter | Total number of receive errors                                                               | `container_id`                                                |
| `docker_container_block_input_total`             | Counter | Total number of bytes read from disk                                                         | `container_id`                                                |
| `docker_container_block_output_total`            | Counter | Total number of bytes written to disk                                                        | `container_id`                                                |
| `docker_container_exit_code`                     | Gauge   | Exit code of the container                                                                   | `container_id`                                                |
| `docker_container_restart_count`                 | Counter | Number of times the container has been restarted                                             | `container_id`                                                |

## Usage

### Running the exporter

```bash
# Run with default settings
./docker-exporter

# Run with custom port and address
./docker-exporter --port 9101 --address 127.0.0.1

# Enable verbose logging
./docker-exporter --verbose

# Use JSON log format
./docker-exporter --log-format=json

# Use logfmt log format (default)
./docker-exporter --log-format=logfmt

# Enable internal metrics
./docker-exporter --internal-metrics
```

### Command-line options

| Option                | Description                                   | Default                       |
|-----------------------|-----------------------------------------------|-------------------------------|
| `--verbose`, `-v`     | Enable verbose mode (debug logs)              | `false`                       |
| `--quiet`, `-q`       | Enable quiet mode (disable info logs)         | `false`                       |
| `--log-format`        | Log format: 'logfmt', 'json', or 'text'       | `logfmt`                      |
| `--internal-metrics`  | Enable internal metrics                       | `false`                       |
| `--address`, `-a`     | Address to listen on                          | `0.0.0.0`                     |
| `--port`, `-p`        | Port to listen on                             | `9100`                        |
| `--docker-host`, `-d` | Host to connect to                            | `unix:///var/run/docker.sock` |

### Logging

The exporter uses structured logging with support for multiple output formats:

- **logfmt** (default): Human-readable key-value format, compatible with log aggregation tools like Grafana Alloy
- **json**: JSON-formatted logs for easy parsing and integration with log processing systems

Logs include contextual information such as container IDs, error details, and operation metadata. Use `--verbose` to enable debug-level logs with additional details about container operations.

Example logfmt output:
```
time=2025-12-18T17:12:19.779Z level=INFO msg="Starting Docker Prometheus exporter" version=dev uid=1001 gid=1001 docker_host=unix:///var/run/docker.sock
```

Example JSON output:
```json
{"time":"2025-12-18T17:12:27.549Z","level":"INFO","msg":"Starting Docker Prometheus exporter","version":"dev","uid":1001,"gid":1001,"docker_host":"unix:///var/run/docker.sock"}
```

### Endpoints

- `/metrics` - Prometheus metrics endpoint
- `/status` - Status endpoint

## Requirements

- Docker daemon running with access to the Docker socket (`/var/run/docker.sock`)
- Go 1.24 or higher (for building)

## Building from source

```bash
go build -o docker-exporter ./cmd/main.go
```

## Develop with air

```bash
go install github.com/air-verse/air@latest

air
```

## Run with docker

```bash
docker run -d --name docker-exporter -p 9100:9100 -v /var/run/docker.sock:/var/run/docker.sock:ro ghcr.io/h3rmt/docker-exporter:latest -p 9100
```

## Run with docker-compose

```yaml
services:
  docker-exporter:
    image: ghcr.io/h3rmt/docker-exporter:latest
    container_name: docker-exporter
    restart: unless-stopped
    ports:
      - "9100:9100"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
```
