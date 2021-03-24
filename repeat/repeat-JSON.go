package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

// ParseJSON Parse JSON file writing nodes to chan
func ParseJSON(path string, ch chan<- Node, l *log.Logger) {
	defer close(ch)

	f, err := os.Open(path)
	if err != nil {
		l.Printf("ERROR %v: %v\n", path, err)
		return
	}
	defer f.Close()

	fileBytes, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatalln(err)
	}

	nodes := make([]map[string](interface{}), 0)
	json.Unmarshal(fileBytes, &nodes)

	for _, properties := range nodes {
		n := NewNode(properties)
		ch <- n
	}

}
