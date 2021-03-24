package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// BuildEnv Build an environment definition array for use with exec.Command
func BuildEnv(properties map[string](interface{})) []string {
	e := os.Environ()
	for k, v := range properties {
		_, ok := v.(map[string](interface{}))
		if !ok {
			e = append(e, fmt.Sprintf("%s=%v", k, v))
			continue
		}

		child := BuildEnv(v.(map[string](interface{})))
		for _, v := range child {
			e = append(e, v)
		}
	}
	return e
}

// ParseArguments Parse program command and filter arguments
func ParseArguments(args []string) ([]string, []Filter) {
	var command []string
	var filters []Filter

	for index, filterDef := range args {
		if filterDef == "-" { // Start of command flag
			command = args[index+1:]
			break
		}

		if len(filterDef) < 4 { // Not a filter C==V
			continue
		}

		filter, err := NewFilter(filterDef)
		if err != nil { // Not a filter
			continue
		}

		filters = append(filters, filter)
	}

	return command, filters
}

// ScheduleNodes Walk path to file and schedule repeats for nodes
func ScheduleNodes(path string, filters *[]Filter, command *[]string, async bool, l *log.Logger, wg *sync.WaitGroup) {
	/* Endure path is valid */

	stat, err := os.Stat(path)
	if err != nil {
		l.Printf("ERROR %v: %v\n", path, err)
		return
	}

	// If path is a directory call function for each entry
	if stat.IsDir() {
		files, err := ioutil.ReadDir(path)
		if err != nil {
			l.Printf("ERROR %v: %v\n", path, err)
			return
		}

		for _, file := range files {
			ScheduleNodes(filepath.Join(path, file.Name()), filters, command, async, l, wg)
		}

		return
	}

	// Create channel to collect nodes
	ch := make(chan Node)

	// Parse inventory files
	if strings.EqualFold(path[len(path)-4:], ".csv") {
		go ParseCSV(path, ch, l)
	} else if strings.EqualFold(path[len(path)-5:], ".json") {
		go ParseJSON(path, ch, l)
	} else {
		l.Printf("ERROR %v: Unknown or unsupported file type\n", path)
		return
	}

	// Read Nodes from channel and process
	for n := range ch {
		if n.Filter(filters) {
			if async {
				wg.Add(1)
				go n.Process(command, true, l, wg)
			} else {
				n.Process(command, false, l, wg)
			}
		}
	}
}

// main Entrypoint
func main() {
	/* Parse arguments - flags */

	async := flag.Bool("async", false, "Process selected nodes asynchronously")
	inventory := flag.String("inventory", "inventory/", "Inventory location")

	bash := flag.Bool("bash", false, "Enable bash helper")
	cmd := flag.Bool("cmd", false, "Enable cmd.exe helper")
	ps := flag.Bool("ps", false, "Enable powershell.exe helper")
	pwsh := flag.Bool("pwsh", false, "Enable pwsh.exe helper")

	flag.Parse()

	/* Parse arguments - filters and command */

	var filters []Filter
	var command []string
	command, filters = ParseArguments(flag.Args())

	/* Add command / process invocation helpers */

	if *bash {
		command = append([]string{"bash", "-c"}, command...)
	}
	if *cmd {
		command = append([]string{"cmd.exe", "/C"}, command...)
	}
	if *ps {
		command = append([]string{"powershell.exe", "-Command"}, command...)
	}
	if *pwsh {
		command = append([]string{"pwsh.exe", "-Command"}, command...)
	}

	/* Schedule nodes for repeat executions of command */

	l := log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lmicroseconds)
	var wg sync.WaitGroup
	ScheduleNodes(*inventory, &filters, &command, *async, l, &wg)

	wg.Wait() // Wait for all goroutines to cleanup and exit
}
