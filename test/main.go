package main

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

var data = `
a1: b1
a2: b2
`

type goStruct struct {
	A1 string `json:"a1,omitempty" yaml:"a1,omitempty"`
	A3 string `json:"a2,omitempty" yaml:"a2,omitempty"`
}

func main() {

	var x goStruct
	if err := yaml.Unmarshal([]byte(data), &x); err != nil {
		panic(err)
	}
	fmt.Println(x)

}
