package osinfo

import (
	"bufio"
	"context"
	"os"
	"strings"

	"github.com/h3rmt/docker-exporter/internal/log"
)

// OSInfo contains information from /etc/os-release
type OSInfo struct {
	Name      string
	VersionID string
}

func GetOSInfo(ctx context.Context) OSInfo {
	info := OSInfo{
		Name:      "Unknown",
		VersionID: "Unknown",
	}

	file, err := os.Open("/etc/os-release")
	if err != nil {
		// File doesn't exist (e.g., on Windows)
		return info
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.GetLogger().WarnContext(ctx, "Failed to close /etc/os-release file", "error", err)
		}
	}(file)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE or KEY="VALUE"
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		value = strings.Trim(value, `"`)

		switch key {
		case "NAME":
			info.Name = cleanOSName(value)
		case "VERSION_ID":
			info.VersionID = value
		}
	}

	return info
}

// cleanOSName removes " Linux" or "/Linux" suffix from OS name
func cleanOSName(name string) string {
	// Remove " Linux" suffix
	name = strings.TrimSuffix(name, " Linux")
	// Remove "/Linux" suffix
	name = strings.TrimSuffix(name, "/Linux")
	return name
}
