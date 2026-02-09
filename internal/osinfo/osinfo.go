package osinfo

import (
	"bufio"
	"os"
	"strings"
	"sync"
	"time"
)

// OSInfo contains information from /etc/os-release
type OSInfo struct {
	Name      string
	VersionID string
}

// osInfoCache provides thread-safe caching of OS info with TTL
type osInfoCache struct {
	mu       sync.RWMutex
	info     OSInfo
	lastRead time.Time
	ttl      time.Duration
}

var cache = &osInfoCache{
	ttl: 5 * time.Minute, // Refresh OS info every 5 minutes
}

// GetCached returns the cached OS info, refreshing if needed
func GetCached() OSInfo {
	cache.mu.RLock()
	if time.Since(cache.lastRead) < cache.ttl && cache.info.Name != "" {
		info := cache.info
		cache.mu.RUnlock()
		return info
	}
	cache.mu.RUnlock()

	// Need to refresh
	cache.mu.Lock()
	defer cache.mu.Unlock()

	// Double-check after acquiring write lock
	if time.Since(cache.lastRead) < cache.ttl && cache.info.Name != "" {
		return cache.info
	}

	// Read fresh OS info
	cache.info = ReadOSRelease()
	cache.lastRead = time.Now()
	return cache.info
}

// ReadOSRelease reads and parses /etc/os-release file
// Returns OSInfo with Name and VersionID, or "Unknown" values if the file doesn't exist (e.g., on Windows)
func ReadOSRelease() OSInfo {
	info := OSInfo{
		Name:      "Unknown",
		VersionID: "Unknown",
	}

	file, err := os.Open("/etc/os-release")
	if err != nil {
		// File doesn't exist (e.g., on Windows)
		return info
	}
	defer file.Close()

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
