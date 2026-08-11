package main

import (
	"fmt"
	"os"
	"strings"
)

const (
	reset   = "\033[0m"
	bold    = "\033[1m"
	dim     = "\033[2m"
	red     = "\033[31m"
	green   = "\033[32m"
	yellow  = "\033[33m"
	blue    = "\033[34m"
	magenta = "\033[35m"
	cyan    = "\033[36m"
	white   = "\033[37m"
)

func isTerm() bool {
	s, _ := os.Stdout.Stat()
	return (s.Mode() & os.ModeCharDevice) != 0
}

func c(col, s string) string {
	if !isTerm() {
		return s
	}
	return col + s + reset
}

func PrintInit(device string, count int) {
	fmt.Println()
	fmt.Printf(" %s %s\n", c(bold+green, "✓"), c(bold, "baki initialized"))
	fmt.Printf("   %s %s\n", c(cyan, "Device:"), device)
	fmt.Printf("   %s %d partitions found\n", c(cyan, "Parts:"), count)
	fmt.Printf("   %s %s\n\n", c(cyan, "Dir:"), c(dim, ".baki/"))
}

func PrintSnapshotCreated(id int, msg string) {
	fmt.Printf(" %s %s #%d\n", c(bold+green, "✓"), c(bold, "Snapshot"), id)
	fmt.Printf("   %s %s\n\n", c(dim, msg), "")
}

func PrintSnapshots(snaps []Snapshot) {
	if len(snaps) == 0 {
		fmt.Printf(" %s\n\n", c(yellow, "No snapshots yet. Run 'baki commit'"))
		return
	}

	fmt.Printf(" %s\n", c(bold+cyan, "📸 Snapshot History"))
	fmt.Println(strings.Repeat("─", 50))

	for i := len(snaps) - 1; i >= 0; i-- {
		s := snaps[i]
		arrow := c(green, "→")
		fmt.Printf(" %s #%d  %s\n", arrow, s.ID, c(bold, s.Message))
		fmt.Printf("     %s\n", c(dim, s.Timestamp))
		fmt.Printf("     %d partitions\n", len(s.Partitions))
		if i > 0 {
			fmt.Println()
		}
	}
	fmt.Println()
}

func PrintStatus(parts []Partition, changes int) {
	fmt.Printf("\n %s\n", c(bold+cyan, "📊 Partition Status"))
	fmt.Println(strings.Repeat("─", 50))

	for _, p := range parts {
		catColor := CategoryColors[p.Category]
		if catColor == "" {
			catColor = white
		}
		name := fmt.Sprintf("%-15s", p.Name)
		size := fmt.Sprintf("%6d KB", p.Size/1024)

		hashShort := p.SHA256
		if len(hashShort) > 12 {
			hashShort = hashShort[:12]
		}

		fmt.Printf("   %s%s %s %s\n",
			c(catColor, name),
			c(dim, size),
			c(dim, hashShort),
			c(catColor, "["+p.Category+"]"),
		)
	}

	status := c(green, "✓ No changes")
	if changes > 0 {
		status = c(red, fmt.Sprintf("⚠ %d partition(s) changed since last snapshot", changes))
	}
	fmt.Printf("\n   %s\n\n", status)
}

func PrintDiff(modified, added, removed []string) {
	if len(modified) == 0 && len(added) == 0 && len(removed) == 0 {
		fmt.Printf(" %s\n\n", c(green, "✓ No changes since last snapshot"))
		return
	}

	fmt.Printf(" %s\n", c(bold+cyan, "🔄 Changes"))

	for _, p := range modified {
		fmt.Printf("   %s %s\n", c(yellow, "~"), p)
	}
	for _, p := range added {
		fmt.Printf("   %s %s\n", c(green, "+"), p)
	}
	for _, p := range removed {
		fmt.Printf("   %s %s\n", c(red, "-"), p)
	}
	fmt.Println()
}

func PrintRecommend(critical, important, optional []string, backedUp int) {
	fmt.Printf("\n %s\n", c(bold+cyan, "💡 Backup Recommendations"))
	fmt.Println(strings.Repeat("─", 50))

	if len(critical) > 0 {
		fmt.Printf("   %s %s\n", c(red, "🔴 CRITICAL"), c(bold, "— backup these first"))
		for _, p := range critical {
			fmt.Printf("     %s %s\n", c(red, "■"), p)
		}
	}
	if len(important) > 0 {
		fmt.Printf("\n   %s\n", c(yellow, "🟡 IMPORTANT"))
		for _, p := range important {
			fmt.Printf("     %s %s\n", c(yellow, "■"), p)
		}
	}
	if len(optional) > 0 {
		fmt.Printf("\n   %s\n", c(green, "🟢 OPTIONAL"))
		for _, p := range optional {
			fmt.Printf("     %s %s\n", c(green, "■"), p)
		}
	}

	fmt.Printf("\n   %s %d/%d backed up\n", c(cyan, "📦"), backedUp, len(critical)+len(important)+len(optional))
	fmt.Printf("   %s baki commit --backup\n\n", c(dim, "Run:"))
}

func PrintHelp() {
	h := `
 %s

  %s        Initialize baki in current directory
  %s        Create a snapshot
  %s        Show snapshot history
  %s        Show partition status vs last snapshot
  %s        Show detailed partition changes
  %s        Rollback to a snapshot
  %s        Show backup recommendations
  %s        Show this help

 %s
  %s
  %s
  %s

`
	fmt.Printf(h,
		c(bold+cyan, "baki — Xiaomi Mod State Manager"),
		c(cyan, "baki init"),
		c(cyan, "baki commit <message>"),
		c(cyan, "baki log"),
		c(cyan, "baki status"),
		c(cyan, "baki diff"),
		c(cyan, "baki checkout <id>"),
		c(cyan, "baki recommend"),
		c(cyan, "baki help"),
		c(bold, "Examples:"),
		c(dim, "  baki init"),
		c(dim, "  baki commit \"Before flashing kernel X\""),
		c(dim, "  baki checkout 1"),
	)
}
