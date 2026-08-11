package main

import (
	"fmt"
	"os"
)

func CmdInit() {
	if err := CheckADB(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := InitBaki(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	parts, err := ListPartitions()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
	}

	cfg, _ := LoadBaki()
	PrintInit(cfg.Device, len(parts))
}

func CmdCommit(msg string, withBackup bool) {
	if err := CheckADB(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if _, err := LoadBaki(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if msg == "" {
		msg = "Snapshot without message"
	}

	parts, err := ListPartitions()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	for i := range parts {
		hash, err := HashPartition(parts[i].Name)
		if err == nil {
			parts[i].SHA256 = hash
		}
	}

	snap := Snapshot{
		Message:    msg,
		Partitions: parts,
	}

	if err := SaveSnapshot(snap); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if withBackup {
		cfg, _ := LoadBaki()
		backupDir := fmt.Sprintf(".baki/backups/%d", cfg.SnapshotNext-1)
		os.MkdirAll(backupDir, 0755)

		for _, p := range parts {
			risk := getRisk(p.Name)
			if risk == "critical" || risk == "important" {
				fmt.Printf("   Backing up %s... ", p.Name)
				if err := BackupPartition(p.Name, fmt.Sprintf("%s/%s.img", backupDir, p.Name)); err != nil {
					fmt.Printf("%s\n", c(red, "FAIL"))
				} else {
					fmt.Printf("%s\n", c(green, "OK"))
				}
			}
		}
	}

	PrintSnapshotCreated(snap.ID, msg)
}

func CmdLog() {
	if _, err := LoadBaki(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	snaps, err := ListSnapshots()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	PrintSnapshots(snaps)
}

func CmdStatus() {
	if err := CheckADB(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if _, err := LoadBaki(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	current, err := ListPartitions()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	for i := range current {
		hash, err := HashPartition(current[i].Name)
		if err == nil {
			current[i].SHA256 = hash
		}
	}

	latest, err := LatestSnapshot()
	changes := 0
	if err == nil {
		snapMap := make(map[string]string)
		for _, p := range latest.Partitions {
			snapMap[p.Name] = p.SHA256
		}
		for _, p := range current {
			if oldHash, ok := snapMap[p.Name]; ok && oldHash != "" && p.SHA256 != "" && oldHash != p.SHA256 {
				changes++
			}
		}
	}

	PrintStatus(current, changes)
}

func CmdDiff() {
	if err := CheckADB(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if _, err := LoadBaki(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	current, err := ListPartitions()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	for i := range current {
		hash, _ := HashPartition(current[i].Name)
		current[i].SHA256 = hash
	}

	latest, err := LatestSnapshot()
	if err != nil {
		PrintDiff(nil, nil, nil)
		return
	}

	oldMap := make(map[string]string)
	for _, p := range latest.Partitions {
		oldMap[p.Name] = p.SHA256
	}

	var modified, added, removed []string
	for _, p := range current {
		if oldHash, ok := oldMap[p.Name]; ok {
			if oldHash != p.SHA256 && oldHash != "" && p.SHA256 != "" {
				modified = append(modified, p.Name)
			}
		} else {
			added = append(added, p.Name)
		}
	}

	curMap := make(map[string]bool)
	for _, p := range current {
		curMap[p.Name] = true
	}
	for _, p := range latest.Partitions {
		if !curMap[p.Name] {
			removed = append(removed, p.Name)
		}
	}

	PrintDiff(modified, added, removed)
}

func CmdCheckout(id int) {
	if err := CheckADB(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if _, err := LoadBaki(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	snap, err := LoadSnapshot(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: snapshot #%d not found\n", id)
		os.Exit(1)
	}

	backupDir := fmt.Sprintf(".baki/backups/%d", id)
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: no backups for snapshot #%d (use 'baki commit --backup')\n", id)
		os.Exit(1)
	}

	fmt.Printf(" %s Restoring snapshot #%d: %s\n", c(bold+yellow, "⚠"), id, snap.Message)
	fmt.Printf("   %s\n\n", c(red, "This will overwrite partitions on your device!"))

	for _, p := range snap.Partitions {
		backupFile := fmt.Sprintf("%s/%s.img", backupDir, p.Name)
		if _, err := os.Stat(backupFile); os.IsNotExist(err) {
			continue
		}
		fmt.Printf("   Restoring %s... ", p.Name)
		if err := RestorePartition(p.Name, backupFile); err != nil {
			fmt.Printf("%s\n", c(red, "FAIL"))
		} else {
			fmt.Printf("%s\n", c(green, "OK"))
		}
	}
	fmt.Println()
}

func CmdRecommend() {
	if err := CheckADB(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if _, err := LoadBaki(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	parts, err := ListPartitions()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	found := make(map[string]bool)
	for _, p := range parts {
		found[p.Name] = true
	}

	var critical, important, optional []string
	for _, kp := range KnownPartitions {
		if !found[kp.Name] {
			continue
		}
		switch kp.Risk {
		case "critical":
			critical = append(critical, fmt.Sprintf("%-15s %s", kp.Name, c(dim, kp.Desc)))
		case "important":
			important = append(important, fmt.Sprintf("%-15s %s", kp.Name, c(dim, kp.Desc)))
		case "optional":
			optional = append(optional, fmt.Sprintf("%-15s %s", kp.Name, c(dim, kp.Desc)))
		}
	}

	total := len(critical) + len(important) + len(optional)
	backedUp := 0
	for _, cat := range []string{"critical", "important", "optional"} {
		snaps, _ := ListSnapshots()
		if len(snaps) > 0 {
			backedUp = total
		}
		_ = cat
	}

	PrintRecommend(critical, important, optional, backedUp)
}

func getRisk(name string) string {
	for _, kp := range KnownPartitions {
		if kp.Name == name {
			return kp.Risk
		}
	}
	return "optional"
}
