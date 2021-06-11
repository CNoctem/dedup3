package main

import (
	"dedup3/collect"
	"dedup3/util"
	"fmt"
	"os"
	"strings"
)

type cmd int

const (
	COLLECT cmd = 0
	LIST    cmd = 1
	HELP    cmd = 2
	UNKNOWN cmd = 3
)

func main() {
	checkArgs()
	command := getCommand()

	switchStart := 3

	var collectionPath string
	if command == COLLECT {
		if len(os.Args) < 4 || !isDir(os.Args[3]) {
			fmt.Println("If command is collect, I need a valid collection path.")
			os.Exit(3)
		}
		collectionPath = os.Args[3]
		switchStart = 4
	}
	filter := parseArgs(switchStart)

	printInit(command, filter, collectionPath)

	if command == COLLECT {
		collect.Collect(os.Args[1], collectionPath, filter)
	}
}

func parseArgs(switchStart int) util.Filter {
	var checkSize util.Filtertype
	var checkDate util.Filtertype
	var extensions []string
	for i := switchStart; i < len(os.Args); i++ {
		a := os.Args[i]
		if a == "--size" || a == "-s" {
			checkSize = util.EQ
		} else if a == "--nesize" || a == "-S" {
			checkSize = util.NEQ
		} else if a == "--date" || a == "-d" {
			checkDate = util.EQ
		} else if a == "--nedate" || a == "-D" {
			checkDate = util.NEQ
		} else if a == "-f" || a == "--filter" {
			i++
			extensions = getExtensions(extensions, i)
		} else {
			fmt.Printf("Unknown argument %s\n", a)
		}
	}
	return util.NewFilter(checkSize, checkDate, extensions)
}

func getExtensions(extensions []string, i int) []string {
	ext := strings.Split(os.Args[i], ",")
	extensionList := make([]string, 0)
	for _, e := range ext {
		extensionList = append(extensionList, "." + e)
	}
	return extensionList
}

func getCommand() cmd {
	switch os.Args[2] {
	case "collect":
		return COLLECT
	case "list":
		return LIST
	case "help":
		return HELP
	default:
		fmt.Printf("Unknown command %s\n", os.Args[2])
		os.Exit(2)
	}
	return UNKNOWN
}

func checkArgs() {
	if len(os.Args) < 3 {
		printHelp()
		os.Exit(0)
	}
	if !isDir(os.Args[1]) {
		fmt.Println("First argument must be the root directory for deduplication.")
		printHelp()
		os.Exit(1)
	}
}

func isDir(arg string) bool {
	info, err := os.Stat(arg)
	if err != nil {
		fmt.Println(err)
	}
	if info == nil {
		return false
	}
	return info.Mode().IsDir()
}

func cmdToString(c cmd) string {
	return []string {"collect", "list", "help", "unknown"}[c]
}

func ftToString(filtertype util.Filtertype) string {
	return []string {"UNSET", "EQ", "NEQ"}[filtertype]
}

func printHelp() {
	fmt.Println("help message")
}

func printInit(command cmd, filter util.Filter, collectionPath string) {
	fmt.Printf("Root: %s\ncommand: %s\nsize: %s\ndate: %s\nextension filter: %s\n",
		os.Args[1], cmdToString(command), ftToString(filter.CheckSize), ftToString(filter.CheckDate), filter.ExtensionList)
	if command == COLLECT {
		fmt.Println("collection path:", collectionPath)
	}
}
