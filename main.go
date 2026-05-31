package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) < 2 || os.Args[1] == "help" || os.Args[1] == "--help" || os.Args[1] == "-h" {
		PrintHelp()
		return
	}

	cmd := os.Args[1]

	switch cmd {
	case "init":
		CmdInit()

	case "commit":
		msg := strings.Join(os.Args[2:], " ")
		withBackup := false
		if strings.Contains(msg, "--backup") {
			withBackup = true
			msg = strings.ReplaceAll(msg, "--backup", "")
			msg = strings.TrimSpace(msg)
		}
		CmdCommit(msg, withBackup)

	case "log":
		CmdLog()

	case "status":
		CmdStatus()

	case "diff":
		CmdDiff()

	case "checkout":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Error: snapshot ID required\n")
			os.Exit(1)
		}
		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid snapshot ID: %s\n", os.Args[2])
			os.Exit(1)
		}
		CmdCheckout(id)

	case "recommend":
		CmdRecommend()

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		fmt.Fprintf(os.Stderr, "Run 'baki help' for usage.\n")
		os.Exit(1)
	}
}
