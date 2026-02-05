# Docker Prometheus Exporter

This project is similar to [prometheus-podman-exporter](https://github.com/containers/prometheus-podman-exporter), but
for Docker containers instead of Podman containers.
It exports Docker container metrics in Prometheus format and also provides a simple homepage with live charts
for cpu and memory usage.

Grafana dashboard is available at [dashboard.json](./dashboard.json)

![dashboard_preview](.github/imgs/img.png)

## Exported Metrics

The exporter provides the following metrics:

| Metric Name                                     | Type    | Description                                                                                  | Labels                                                                    |
|-------------------------------------------------|---------|----------------------------------------------------------------------------------------------|---------------------------------------------------------------------------|
| `docker_exporter_info`                          | Gauge   | Information about the docker exporter                                                        | `hostname`, `version`                                                     |
| `docker_container_info`                         | Gauge   | Container information                                                                        | `hostname`, `container_id`, `name`, `image_id`, `command`, `network_mode` |
| `docker_container_name`                         | Gauge   | Name for the container (can be more than one)                                                | `hostname`, `container_id`, `name`                                        |
| `docker_container_state`                        | Gauge   | Container State (0=created, 1=running, 2=paused, 3=restarting, 4=removing, 5=exited, 6=dead) | `hostname`, `container_id`                                                |
| `docker_container_created_seconds`              | Gauge   | Timestamp in seconds when the container was created                                          | `hostname`, `container_id`                                                |
| `docker_container_started_seconds`              | Gauge   | Timestamp in seconds when the container was started                                          | `hostname`, `container_id`                                                |
| `docker_container_finished_at_seconds`          | Gauge   | Timestamp in seconds when the container finished                                             | `hostname`, `container_id`                                                |
| `docker_container_ports`                        | Gauge   | Forwarded Ports                                                                              | `hostname`, `container_id`, `public_port`, `private_port`, `ip`, `type`   |
| `docker_container_rootfs_size_bytes`            | Gauge   | Size of rootfs in this container in bytes                                                    | `hostname`, `container_id`                                                |
| `docker_container_rw_size_bytes`                | Gauge   | Size of files that have been created or changed by this container in bytes                   | `hostname`, `container_id`                                                |
| `docker_container_pids`                         | Gauge   | Number of processes running in the container                                                 | `hostname`, `container_id`                                                |
| `docker_container_cpu_user_nanoseconds_total`   | Counter | Time (in nanoseoconds) spent by tasks                                                        | `hostname`, `container_id`                                                |
| `docker_container_cpu_kernel_nanoseconds_total` | Counter | Time (in nanoseoconds) spent by tasks in user mode                                           | `hostname`, `container_id`                                                |
| `docker_container_cpu_nanoseconds_total`        | Counter | Time (in nanoseoconds) spent by tasks in kernel mode                                         | `hostname`, `container_id`                                                |
| `docker_container_cpu_percent`                  | Gauge   | Percentage of CPU used by the container (relative to max available CPU cores)                | `hostname`, `container_id`                                                |
| `docker_container_cpu_percent_host`             | Gauge   | Percentage of CPU used by the container (relative to host CPU cores)                         | `hostname`, `container_id`                                                |
| `docker_container_mem_limit_kib`                | Gauge   | Container memory limit in KiB                                                                | `hostname`, `container_id`                                                |
| `docker_container_mem_usage_kib`                | Gauge   | Container memory usage in KiB                                                                | `hostname`, `container_id`                                                |
| `docker_container_net_send_bytes_total`         | Counter | Total number of bytes sent                                                                   | `hostname`, `container_id`                                                |
| `docker_container_net_send_dropped_total`       | Counter | Total number of send packet drop                                                             | `hostname`, `container_id`                                                |
| `docker_container_net_send_errors_total`        | Counter | Total number of send errors                                                                  | `hostname`, `container_id`                                                |
| `docker_container_net_receive_bytes_total`      | Counter | Total number of bytes received                                                               | `hostname`, `container_id`                                                |
| `docker_container_net_receive_dropped_total`    | Counter | Total number of receive packet drop                                                          | `hostname`, `container_id`                                                |
| `docker_container_net_receive_errors_total`     | Counter | Total number of receive errors                                                               | `hostname`, `container_id`                                                |
| `docker_container_block_input_total`            | Counter | Total number of bytes read from disk                                                         | `hostname`, `container_id`                                                |
| `docker_container_block_output_total`           | Counter | Total number of bytes written to disk                                                        | `hostname`, `container_id`                                                |
| `docker_container_exit_code`                    | Gauge   | Exit code of the container                                                                   | `hostname`, `container_id`                                                |
| `docker_container_restart_count`                | Counter | Number of times the container has been restarted                                             | `hostname`, `container_id`                                                |

`docker_container_rootfs_size_bytes` and `docker_container_rw_size_bytes` are cached and only updated every 5 minutes.
This can be customized with the --size-cache-seconds flag.

`docker_container_cpu_percent` should probably be prefered over `docker_container_cpu_percent_host` as it takes the
container's cgroup settings into account. (you can use `--cpus=3` to limit the container to only three cpu cores which
this metric will report correctly)

![dashboard_preview](.github/imgs/img_1.png)

## Usage

### Command-line options

| Option                  | Description                                             | Default                       |
|-------------------------|---------------------------------------------------------|-------------------------------|
| `--verbose`, `-v`       | Enable verbose mode (debug logs)                        | `false`                       |
| `--quiet`, `-q`         | Enable quiet mode (disable info logs)                   | `false`                       |
| `--trace`               | Enable trace mode (very vebose logs)                    | `false`                       |
| `--log-format`          | Log format: 'logfmt' or 'json'                          | `logfmt`                      |
| `--homepage`            | Show homepage with charts.                              | `true`                        |
| `--internal-metrics`    | Enable internal go metrics                              | `false`                       |
| `--size-cache-duration` | Duration to wait before refreshing container size cache | `300s`                        |
| `--address`, `-a`       | Address to listen on                                    | `0.0.0.0`                     |
| `--port`, `-p`          | Port to listen on                                       | `9100`                        |
| `--docker-host`, `-d`   | Host to connect to                                      | `unix:///var/run/docker.sock` |

### Endpoints

- `/metrics` - Prometheus metrics endpoint
- `/status` - Status endpoint
- `/` - Homepage with live charts

<table>
  <tr>
    <td><img src=".github/imgs/img_2.png" alt="dashboard_preview" /></td>
    <td><img src=".github/imgs/img_3.png" alt="dashboard_preview" /></td>
  </tr>
</table>

### Logging

The exporter uses structured logging with support for multiple output formats:

- **logfmt** (default): Human-readable key-value format, compatible with log aggregation tools like Grafana Alloy
- **json**: JSON-formatted logs for easy parsing and integration with log processing systems

Logs include contextual information such as container IDs, error details, and operation metadata. Use `--verbose` to
enable debug-level logs with additional details about container operations.

Example logfmt output:

```
time=2025-12-18T17:12:19.779Z level=INFO msg="Starting Docker Prometheus exporter" version=dev uid=1001 gid=1001 docker_host=unix:///var/run/docker.sock
```

Example JSON output:

```json
{
  "time": "2025-12-18T17:12:27.549Z",
  "level": "INFO",
  "msg": "Starting Docker Prometheus exporter",
  "version": "dev",
  "uid": 1001,
  "gid": 1001,
  "docker_host": "unix:///var/run/docker.sock"
}
```

### Running the exporter

```bash
# Run with default settings
./docker-exporter

# Run with custom port and address
./docker-exporter --port 9101 --address 127.0.0.1

# Enable verbose logging and disable homepage
./docker-exporter --verbose --homepage=false

# Use JSON log format
./docker-exporter --log-format=json

# Enable internal metrics and use docker with tcp socket
./docker-exporter --internal-metrics --docker-host tcp://127.0.0.1:2375
```

## Building from source

```bash
go build -o docker-exporter ./cmd/main.go
```

## Run with docker

```bash
docker run -d --name docker-exporter \
  -e IP="$(ip route get 1.1.1.1 | head -1 | awk '{print $7}')" \
  -e TZ="Europe/Berlin" -p 9100:9100 \
  -v /var/run/docker.sock:/var/run/docker.sock:ro \
  -v /etc/hostname:/etc/hostname:ro -v /proc/stat:/proc/stat:ro -v /proc/meminfo:/proc/meminfo:ro \
  ghcr.io/h3rmt/docker-exporter:latest -p 9100 --log-format json
```

## Run with docker-compose

```yaml
services:
  docker_exporter:
    image: ghcr.io/h3rmt/docker-exporter:latest
    container_name: docker_exporter
    restart: always
    environment:
      - TZ="Europe/Berlin"
      - IP="10.10.10.10"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - /etc/hostname:/etc/hostname:ro
      - /proc/stat:/proc/stat:ro
      - /proc/meminfo:/proc/meminfo:ro
    ports:
      - 9100:9100
    command: [ "--size-cache-seconds=600", "-v" ]
```

### Running in docker in lxc

When running in docker in lxc mem and cpu metrics are either collected from the host or from the container depending on
the container's cgroup settings.
To collect metrics from the lxc container running the docker container a helper script is needed to access /proc/meminfo
in the lxc.

```bash
#!/usr/bin/env bash
FILE=./meminfo
echo "Starting meminfo script on $FILE"
# Check if it's a directory and remove it, or exit if it's a file
if [[ -d $FILE ]]; then
    echo "Directory $FILE exists. Removing."
    rm -rf "$FILE"
elif [[ -f $FILE ]]; then
    echo "File $FILE already exists. Exiting."
    exit 1
fi

trap 'rm -f $FILE' EXIT
# Continuously copy /proc/meminfo to ./meminfo every 2 seconds
while true; do
cat /proc/meminfo > "$FILE"
sleep 2
done
```

Start this script as a daemon.

```yaml
services:
  docker_exporter:
    image: ghcr.io/h3rmt/docker-exporter:latest
    container_name: docker_exporter
    restart: always
    environment:
      - TZ="Europe/Berlin"
      - IP="10.10.10.10"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - /etc/hostname:/etc/hostname:ro
      - /proc/stat:/proc/stat:ro
      # mount the meminfo file from the script as a volume
      - ./meminfo:/proc/meminfo:ro
    ports:
      - 9100:9100
    command: [ "--size-cache-seconds=600" ]
```