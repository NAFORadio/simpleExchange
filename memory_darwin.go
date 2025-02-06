// +build darwin

package main

import (
    "fmt"
    "syscall"
)

func checkAvailableMemory() (uint64, error) {
    // Use sysctl to get memory info on macOS
    var hint int64
    var size = uint64(8)
    
    // Get free memory pages
    _, err := syscall.Sysctl("vm.page_free_count")
    if err != nil {
        return 0, fmt.Errorf("failed to get free memory: %v", err)
    }
    
    // Get page size
    pageSize := uint64(syscall.Getpagesize())
    
    // Calculate available memory
    available := uint64(hint) * pageSize
    return available, nil
} 