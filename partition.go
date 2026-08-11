package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os/exec"
	"strings"
)

type Partition struct {
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	SHA256   string `json:"sha256"`
	Path     string `json:"path"`
	Category string `json:"category"`
}

type XiaomiPart struct {
	Name     string
	Risk     string
	Category string
	Desc     string
}

var KnownPartitions = []XiaomiPart{
	{"boot", "critical", "boot", "Kernel & ramdisk"},
	{"dtbo", "critical", "boot", "Device tree blob"},
	{"super", "critical", "system", "System partition (system/product/vendor)"},
	{"persist", "critical", "data", "IMEI & device data"},
	{"modem", "important", "radio", "Modem firmware"},
	{"efs", "critical", "radio", "IMEI/EFS data"},
	{"misc", "important", "misc", "Bootloader misc data"},
	{"recovery", "optional", "boot", "Recovery image"},
	{"cache", "optional", "system", "Cache partition"},
	{"userdata", "important", "data", "User data"},
	{"vbmeta", "critical", "boot", "Verified boot metadata"},
	{"vbmeta_system", "important", "boot", "System vbmeta"},
	{"vbmeta_vendor", "important", "boot", "Vendor vbmeta"},
	{"cust", "optional", "system", "Customization data"},
}

var CategoryColors = map[string]string{
	"boot":   "\033[36m",
	"system": "\033[33m",
	"data":   "\033[35m",
	"radio":  "\033[31m",
	"misc":   "\033[32m",
}

var RiskColors = map[string]string{
	"critical":  "\033[31m",
	"important": "\033[33m",
	"optional":  "\033[32m",
}

func adbExec(args ...string) (string, error) {
	cmd := exec.Command("adb", args...)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func CheckADB() error {
	out, err := adbExec("devices")
	if err != nil {
		return fmt.Errorf("adb not found or not working: %v", err)
	}
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		if strings.Contains(line, "\tdevice") {
			return nil
		}
	}
	return fmt.Errorf("no device connected.\nRun 'adb devices' to check connection")
}

func ListPartitions() ([]Partition, error) {
	out, err := adbExec("shell", "ls", "-l", "/dev/block/bootdevice/by-name/")
	if err != nil {
		return nil, fmt.Errorf("cannot list partitions: %v", err)
	}

	lines := strings.Split(out, "\n")
	var parts []Partition

	for _, line := range lines {
		if !strings.Contains(line, "->") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 9 {
			continue
		}
		name := fields[len(fields)-1]
		realPath := fields[len(fields)-3]
		sizeStr := fields[4]

		if name == "" || realPath == "" {
			continue
		}

		p := Partition{
			Name: name,
			Path: realPath,
			Size: parseSize(sizeStr),
		}
		p.Category = categorizePartition(name)
		parts = append(parts, p)
	}

	return parts, nil
}

func HashPartition(name string) (string, error) {
	devPath := fmt.Sprintf("/dev/block/bootdevice/by-name/%s", name)
	out, err := adbExec("shell", "sha256sum", devPath)
	if err != nil {
		return "", err
	}
	fields := strings.Fields(out)
	if len(fields) == 0 {
		return "", fmt.Errorf("no hash output for %s", name)
	}
	return fields[0], nil
}

func HashBytes(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func BackupPartition(name, destPath string) error {
	src := fmt.Sprintf("/dev/block/bootdevice/by-name/%s", name)
	adbSrc := fmt.Sprintf("/sdcard/baki_%s.img", name)

	_, err := adbExec("shell", "dd", "if="+src, "of="+adbSrc, "bs=1M")
	if err != nil {
		return fmt.Errorf("dd failed on device: %v", err)
	}

	_, err = adbExec("pull", adbSrc, destPath)
	if err != nil {
		return fmt.Errorf("pull failed: %v", err)
	}

	adbExec("shell", "rm", "-f", adbSrc)
	return nil
}

func RestorePartition(name, srcPath string) error {
	adbSrc := fmt.Sprintf("/sdcard/baki_restore_%s.img", name)

	_, err := adbExec("push", srcPath, adbSrc)
	if err != nil {
		return fmt.Errorf("push failed: %v", err)
	}

	dst := fmt.Sprintf("/dev/block/bootdevice/by-name/%s", name)
	_, err = adbExec("shell", "dd", "if="+adbSrc, "of="+dst, "bs=1M")
	if err != nil {
		return fmt.Errorf("dd restore failed: %v", err)
	}

	adbExec("shell", "rm", "-f", adbSrc)
	return nil
}

func categorizePartition(name string) string {
	for _, kp := range KnownPartitions {
		if kp.Name == name {
			return kp.Category
		}
	}
	return "other"
}

func parseSize(s string) int64 {
	var size int64
	fmt.Sscanf(s, "%d", &size)
	return size
}
