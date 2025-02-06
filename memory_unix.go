//go:build !windows

package main

import (
    "fmt"
    "os"
    "strconv"
    "strings"
)

func checkAvailableMemory() (uint64, error) {
    // For Unix/Linux/macOS systems, try to get memory info from /proc/meminfo
    data, err := os.ReadFile("/proc/meminfo")
    if err != nil {
        return 0, fmt.Errorf("failed to read memory info: %v", err)
    }

    // Parse the MemAvailable line
    lines := strings.Split(string(data), "\n")
    for _, line := range lines {
        if strings.HasPrefix(line, "MemAvailable:") {
            fields := strings.Fields(line)
            if len(fields) < 2 {
                continue
            }
            available, err := strconv.ParseUint(fields[1], 10, 64)
            if err != nil {
                return 0, fmt.Errorf("failed to parse memory value: %v", err)
            }
            // Convert from KB to bytes
            return available * 1024, nil
        }
    }

    return 0, fmt.Errorf("could not determine available memory")
} 