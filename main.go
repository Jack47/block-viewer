package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/Jack47/block-parser/block"
)

var Bin = "./binary/48.bin"
var Bin2 = "./binary/b9.bin"

func print(blob interface{}) {
	marshaled, err := json.Marshal(blob)
	if err != nil {
		fmt.Printf("marshal parsed block failed: %v", err)
		os.Exit(1)
	}
	fmt.Printf("%s", marshaled)
}

func main() {
	bytes, err := ioutil.ReadFile(Bin)
	if err != nil {
		fmt.Printf("open file failed: %v", err)
		os.Exit(1)
	}
	b, err := block.NewBlock(bytes)
	if err != nil {
		fmt.Printf("parse block failed: %s", err)
		os.Exit(1)
	}
	// fmt.Printf("Parsed Block:\n")
	print(b)
}
