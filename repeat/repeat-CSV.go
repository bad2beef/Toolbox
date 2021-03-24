package main

import (
	"encoding/csv"
	"io"
	"log"
	"os"
)

// ParseCSV Parse CSV file writing found nodes to chan
func ParseCSV(path string, ch chan<- Node, l *log.Logger) {
	defer close(ch)

	f, err := os.Open(path)
	if err != nil {
		l.Printf("ERROR %v: %v\n", path, err)
		return
	}
	defer f.Close()

	r := csv.NewReader(f)
	var schema []string = nil

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalln(err)
		}

		if schema == nil {
			for _, item := range record {
				schema = append(schema, item)
			}
			continue
		}

		properties := make(map[string](interface{}), len(record))
		n := NewNode(properties)

		for index, item := range record {
			n.Properties[schema[index]] = item
		}

		ch <- n
	}
}
