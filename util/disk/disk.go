// Package disk 基于 gopsutil 提供跨平台的系统磁盘挂载信息获取
package disk

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/shirou/gopsutil/v4/disk"
)

// Info 磁盘挂载信息
type Info struct {
	MountPoint  string  // 挂载点路径
	FSType      string  // 文件系统类型
	Device      string  // 设备名
	TotalBytes  int64   // 总容量（字节）
	UsedBytes   int64   // 已用容量（字节）
	FreeBytes   int64   // 可用容量（字节）
	UsedPercent float64 // 使用百分比
}

// GetMountPoints 返回系统的物理磁盘挂载点列表
// 使用 gopsutil 跨平台获取分区与使用量，自动过滤虚拟文件系统
func GetMountPoints() []Info {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return nil
	}

	var result []Info
	for _, p := range partitions {
		if isVirtual(p.Device, p.Fstype, p.Mountpoint) {
			continue
		}

		usage, err := disk.Usage(p.Mountpoint)
		if err != nil {
			continue
		}

		result = append(result, Info{
			MountPoint:  p.Mountpoint,
			FSType:      p.Fstype,
			Device:      p.Device,
			TotalBytes:  int64(usage.Total),
			UsedBytes:   int64(usage.Used),
			FreeBytes:   int64(usage.Free),
			UsedPercent: usage.UsedPercent,
		})
	}

	return result
}

// isVirtual 判断是否为虚拟文件系统
func isVirtual(device, fstype, mountPoint string) bool {
	ft := strings.ToLower(fstype)
	switch ft {
	case "devfs", "devtmpfs", "proc", "sysfs", "tmpfs", "cgroup", "cgroup2",
		"debugfs", "tracefs", "pstore", "hugetlbfs", "configfs", "securityfs",
		"fusectl", "mqueue", "bpf", "ramfs", "rootfs", "autofs",
		"rpc_pipefs", "binfmt_misc":
		return true
	}

	if runtime.GOOS == "darwin" {
		if strings.HasPrefix(device, "map ") {
			return true
		}
	}

	if runtime.GOOS == "linux" {
		if !strings.HasPrefix(device, "/dev/") {
			return true
		}
	}

	return false
}

// DirSize 递归计算目录总大小（字节）
func DirSize(path string) int64 {
	if path == "" {
		return 0
	}
	fi, err := os.Stat(path)
	if err != nil {
		return 0
	}
	if !fi.IsDir() {
		return fi.Size()
	}

	var size int64
	filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size
}
