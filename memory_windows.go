//go:build windows

package main

import (
    "syscall"
    "unsafe"
)

type memoryStatusEx struct {
    dwLength                uint32
    dwMemoryLoad           uint32
    ullTotalPhys           uint64
    ullAvailPhys           uint64
    ullTotalPageFile       uint64
    ullAvailPageFile       uint64
    ullTotalVirtual        uint64
    ullAvailVirtual        uint64
    ullAvailExtendedVirtual uint64
}

func checkAvailableMemory() (uint64, error) {
    var memInfo memoryStatusEx
    memInfo.dwLength = uint32(unsafe.Sizeof(memInfo))
    
    kernel32 := syscall.NewLazyDLL("kernel32.dll")
    globalMemoryStatusEx := kernel32.NewProc("GlobalMemoryStatusEx")
    
    r1, _, err := globalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&memInfo)))
    if r1 == 0 {
        return 0, err
    }
    return memInfo.ullAvailPhys, nil
} 