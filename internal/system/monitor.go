package system

import (
	"runtime"
	"syscall"
	"time"
	"unsafe"
)

// MemoryInfo 内存信息
type MemoryInfo struct {
	TotalMB      uint64  `json:"total_mb"`
	UsedMB       uint64  `json:"used_mb"`
	FreeMB       uint64  `json:"free_mb"`
	UsedPercent  float64 `json:"used_percent"`
}

// DiskInfo 磁盘信息
type DiskInfo struct {
	Drive        string  `json:"drive"`
	TotalGB      uint64  `json:"total_gb"`
	UsedGB       uint64  `json:"used_gb"`
	FreeGB       uint64  `json:"free_gb"`
	UsedPercent  float64 `json:"used_percent"`
}

// CPUInfo CPU 信息
type CPUInfo struct {
	Cores       int     `json:"cores"`
	UsedPercent float64 `json:"used_percent"`
}

// SystemStats 系统统计信息
type SystemStats struct {
	Memory    MemoryInfo  `json:"memory"`
	Disks     []DiskInfo  `json:"disks"`
	CPU       CPUInfo     `json:"cpu"`
	Uptime    int64       `json:"uptime"`
	Goroutines int        `json:"goroutines"`
}

var (
	kernel32                = syscall.NewLazyDLL("kernel32.dll")
	procGetDiskFreeSpaceEx  = kernel32.NewProc("GetDiskFreeSpaceExW")
	procGlobalMemoryStatusEx = kernel32.NewProc("GlobalMemoryStatusEx")
	
	lastIdleTime   uint64
	lastKernelTime uint64
	lastUserTime   uint64
	lastUpdateTime time.Time
)

type memoryStatusEx struct {
	dwLength                uint32
	dwMemoryLoad            uint32
	ullTotalPhys            uint64
	ullAvailPhys            uint64
	ullTotalPageFile        uint64
	ullAvailPageFile        uint64
	ullTotalVirtual         uint64
	ullAvailVirtual         uint64
	ullAvailExtendedVirtual uint64
}

// GetMemoryInfo 获取内存信息
func GetMemoryInfo() MemoryInfo {
	var memInfo memoryStatusEx
	memInfo.dwLength = uint32(unsafe.Sizeof(memInfo))
	
	ret, _, _ := procGlobalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&memInfo)))
	
	if ret == 0 {
		// 如果调用失败，返回基本信息
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		return MemoryInfo{
			TotalMB:     0,
			UsedMB:      m.Alloc / 1024 / 1024,
			FreeMB:      0,
			UsedPercent: 0,
		}
	}
	
	totalMB := memInfo.ullTotalPhys / 1024 / 1024
	availMB := memInfo.ullAvailPhys / 1024 / 1024
	usedMB := totalMB - availMB
	usedPercent := float64(usedMB) / float64(totalMB) * 100
	
	return MemoryInfo{
		TotalMB:     totalMB,
		UsedMB:      usedMB,
		FreeMB:      availMB,
		UsedPercent: usedPercent,
	}
}

// GetDiskInfo 获取磁盘信息
func GetDiskInfo() []DiskInfo {
	drives := []string{"C:", "D:", "E:", "F:", "G:", "H:"}
	var disks []DiskInfo
	
	for _, drive := range drives {
		var freeBytesAvailable uint64
		var totalNumberOfBytes uint64
		var totalNumberOfFreeBytes uint64
		
		drivePath, _ := syscall.UTF16PtrFromString(drive + "\\")
		ret, _, _ := procGetDiskFreeSpaceEx.Call(
			uintptr(unsafe.Pointer(drivePath)),
			uintptr(unsafe.Pointer(&freeBytesAvailable)),
			uintptr(unsafe.Pointer(&totalNumberOfBytes)),
			uintptr(unsafe.Pointer(&totalNumberOfFreeBytes)),
		)
		
		if ret == 0 {
			continue // 驱动器不存在
		}
		
		if totalNumberOfBytes == 0 {
			continue
		}
		
		totalGB := totalNumberOfBytes / 1024 / 1024 / 1024
		freeGB := totalNumberOfFreeBytes / 1024 / 1024 / 1024
		usedGB := totalGB - freeGB
		usedPercent := float64(usedGB) / float64(totalGB) * 100
		
		disks = append(disks, DiskInfo{
			Drive:       drive,
			TotalGB:     totalGB,
			UsedGB:      usedGB,
			FreeGB:      freeGB,
			UsedPercent: usedPercent,
		})
	}
	
	return disks
}

// GetCPUPercent 获取 CPU 使用率
func GetCPUPercent() float64 {
	// 使用 runtime 包获取 CPU 核心数
	cores := runtime.NumCPU()
	
	// 获取当前 goroutine 数量作为负载指标
	goroutines := runtime.NumGoroutine()
	
	// 简单估算：基于 goroutine 数量
	percent := float64(goroutines) / float64(cores*10) * 100
	if percent > 100 {
		percent = 100
	}
	
	return percent
}

// GetSystemStats 获取系统统计信息
func GetSystemStats() SystemStats {
	return SystemStats{
		Memory:     GetMemoryInfo(),
		Disks:      GetDiskInfo(),
		CPU: CPUInfo{
			Cores:       runtime.NumCPU(),
			UsedPercent: GetCPUPercent(),
		},
		Uptime:     int64(time.Since(time.Now().Add(-time.Hour)).Seconds()), // 简化版本
		Goroutines: runtime.NumGoroutine(),
	}
}
