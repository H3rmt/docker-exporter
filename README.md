# Docker Prometheus Exporter

This project is similar to [prometheus-podman-exporter](https://github.com/containers/prometheus-podman-exporter), but for Docker containers instead of Podman containers. It exports Docker container metrics in Prometheus format.

## Exported Metrics

The exporter provides the following metrics:

| Metric Name                | Type    | Description                         | Labels                                                      |
|----------------------------|---------|-------------------------------------|-------------------------------------------------------------|
| `docker_container_info`    | Gauge   | Docker container information        | `container_id`, `name`, `image_id`, `command`               |
| `docker_container_name`    | Gauge   | Docker container name               | `container_id`, `name`                                      |
| `docker_container_state`   | Gauge   | Docker container state              | `container_id`                                              |
| `docker_container_created` | Counter | Docker container creation timestamp | `container_id`                                              |
| `docker_container_ports`   | Gauge   | Docker container exposed ports      | `container_id`, `public_port`, `private_port`, `ip`, `type` |

### Container States

The `docker_container_state` metric uses the following values:

| Value | State      |
|-------|------------|
| 0     | Created    |
| 1     | Running    |
| 2     | Paused     |
| 3     | Restarting |
| 4     | Removing   |
| 5     | Exited     |
| 6     | Dead       |

## Usage

### Running the exporter

```bash
# Run with default settings
./docker-exporter

# Run with custom port and address
./docker-exporter --port 9101 --address 127.0.0.1

# Enable verbose logging
./docker-exporter --verbose

# Enable internal metrics
./docker-exporter --internal-metrics
```

### Command-line options

| Option                | Description                           | Default                       |
|-----------------------|---------------------------------------|-------------------------------|
| `--verbose`, `-v`     | Enable verbose mode (debug logs)      | `false`                       |
| `--quiet`, `-q`       | Enable quiet mode (disable info logs) | `false`                       |
| `--internal-metrics`  | Enable internal metrics               | `false`                       |
| `--address`, `-a`     | Address to listen on                  | `0.0.0.0`                     |
| `--port`, `-p`        | Port to listen on                     | `9100`                        |
| `--docker-host`, `-d` | Host to connect to                    | `unix:///var/run/docker.sock` |

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