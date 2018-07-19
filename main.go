package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Jack47/block-viewer/block"
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
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s block_hash\nfor example:\n%s 000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f", os.Args[0], os.Args[0])
		os.Exit(1)
	}
	if len(os.Args[1]) != 64 {
		fmt.Printf("block_hash must be exactly 32 bytes")
		os.Exit(1)
	}
	bytes, err := block.FetchBlock(os.Args[1], false /* dumpFile */)
	if err != nil {
		fmt.Printf("fetch block failed: %s", err)
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
