package main

import (
	"errors"
	"strconv"
	"strings"
)

// Filter Filter to apply to a list of nodes
type Filter struct {
	Key        string
	Value      interface{}
	Type       string
	Comparator string // ==, >=, <=, !=, ~=

}

// NewFilter Create a Filter object from a string filter definition
func NewFilter(filter string) (Filter, error) {
	index := strings.LastIndex(filter, "=")
	if index < 0 {
		return Filter{}, errors.New("Invalid filter")
	}

	/* Parse for type hints */

	var vAsString string = filter[index+1:]
	var vAsI interface{} = vAsString
	vtype := "s"

	vAsFloat, err := strconv.ParseFloat(vAsString, 64) // If parsed, set values
	if err == nil {
		vtype = "f"
		vAsI = vAsFloat
	}

	vAsInt, err := strconv.ParseInt(vAsString, 10, 64) // If parsed, set values
	if err == nil {
		vtype = "i"
		vAsI = vAsInt
	}

	return Filter{
		Key:        filter[:index-1],
		Value:      vAsI,
		Type:       vtype,
		Comparator: filter[index-1 : index+1],
	}, nil
}
