package disk

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/shirou/gopsutil/v4/disk"
)

type Info struct {
	MountPoint  string
	FSType      string
	Device      string
	TotalBytes  int64
	UsedBytes   int64
	FreeBytes   int64
	UsedPercent float64
}

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
