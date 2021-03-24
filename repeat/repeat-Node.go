package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

// Node Node information object
type Node struct {
	ID         uint32
	Properties map[string]interface{}
}

// NewNode Creates and returns a new Node struct, optionally pre-setting properties
func NewNode(properties map[string](interface{})) Node {
	if properties == nil {
		properties = make(map[string](interface{}), 0)
	}

	return Node{
		ID:         rand.Uint32(),
		Properties: properties,
	}
}

// GetProperty Return the given property for the given node
func (n Node) GetProperty(p *string) (interface{}, error) {
	var v interface{}
	v = n.Properties

	path := strings.Split(*p, ".")
	for _, prop := range path {
		m, ok := v.(map[string]interface{})
		if !ok {
			return nil, errors.New("Cannot traverse property")
		}

		v, ok = m[prop]
		if !ok {
			return nil, errors.New("Property not found")
		}
	}

	return v, nil
}

// Filter Indicate if the given node is in-scope based on passed filteres
func (n Node) Filter(filters *[]Filter) bool {
	for _, f := range *filters {
		v, err := n.GetProperty(&f.Key)

		if err != nil { // Handle missing properties
			switch f.Comparator {
			case "==": // key == value
				return false
			case "!=": // key != value
				continue
			}

			return false
		}

		/* Check property against filter. To support multiple filters we only
		test for conditions *NOT* met, returning false in such cases. Fall
		through and return true if no conditions *NOT* met.  */

		switch f.Type { // Route based on type
		case "s": // Test as string
			switch f.Comparator { // Route based on comparator
			case "==":
				if !strings.EqualFold(f.Value.(string), v.(string)) {
					return false
				}
			case "!=":
				if strings.EqualFold(f.Value.(string), v.(string)) {
					return false
				}
			case ">=":
				if !(len(v.(string)) >= len(f.Value.(string))) {
					return false
				}
			case "<=":
				if !(len(v.(string)) <= len(f.Value.(string))) {
					return false
				}
			case "~=":
				if -1 == strings.Index(strings.ToLower(v.(string)), strings.ToLower(f.Value.(string))) {
					return false
				}
			}
		case "f": // Test as float
			vparsed, err := strconv.ParseFloat(fmt.Sprintf("%v", v), 64)
			if err != nil {
				return false
			}

			switch f.Comparator {
			case "==":
				if !(vparsed == f.Value.(float64)) {
					return false
				}
			case "!=":
				if vparsed == f.Value.(float64) {
					return false
				}
			case ">=":
				if !(vparsed >= f.Value.(float64)) {
					return false
				}
			case "<=":
				if !(vparsed <= f.Value.(float64)) {
					return false
				}
			}
		case "i": // Test as integer
			vparsed, err := strconv.ParseInt(fmt.Sprintf("%v", v), 10, 64)
			if err != nil {
				return false
			}

			switch f.Comparator {
			case "==":
				if !(vparsed == f.Value.(int64)) {
					return false
				}
			case "!=":
				if vparsed == f.Value.(int64) {
					return false
				}
			case ">=":
				if !(vparsed >= f.Value.(int64)) {
					return false
				}
			case "<=":
				if !(vparsed <= f.Value.(int64)) {
					return false
				}
			}
		}
	}

	return true
}

// Process Process node variables and execute command
func (n Node) Process(c *[]string, async bool, l *log.Logger, wg *sync.WaitGroup) {
	if async {
		defer wg.Done()
	}

	l.Printf("%X > %v\n", n.ID, n.Properties)

	var myc []string
	for _, subc := range *c {
		for {
			subcStart := strings.Index(subc, "${")
			subcEnd := strings.Index(subc, "}")
			if subcStart < 0 || subcEnd < 0 { // Break on no remaining subs
				break
			}

			subcName := subc[subcStart+2 : subcEnd]
			subcValue, ok := n.Properties[subcName]
			if !ok {
				subcValue = ""
			}

			subc = strings.Replace(subc, subc[subcStart:subcEnd+1], fmt.Sprintf("%v", subcValue), -1)
		}
		myc = append(myc, subc)
	}

	mycCommand := myc[0]
	mycArgs := myc[1:]

	var stdout, stderr bytes.Buffer
	cmd := exec.Command(mycCommand, mycArgs...)
	cmd.Env = BuildEnv(n.Properties)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		l.Printf("%X ! %v\n", n.ID, err)
	}

	if stdout.Len() > 0 {
		l.Printf("%X 1 %v\n", n.ID, strings.TrimSpace(stdout.String()))
	}

	if stderr.Len() > 0 {
		l.Printf("%X 2 %v\n", n.ID, strings.TrimSpace(stderr.String()))
	}

	l.Printf("%X < %v\n", n.ID, cmd.ProcessState.ExitCode())
}
