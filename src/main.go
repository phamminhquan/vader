package main

import (
	"fmt"
	"flag"
)

type Bitfield struct {
	ID int
	Name string `toml:"name"`
	Width *uint64 `toml:"width"`
	DefaultValue *uint64 `toml:"default-value"`
	Access string `toml:"access"`
}

type Register struct {
	ID int
	Name string `toml:"name"`
	Address *uint64 `toml:"address"`
	BitfieldReference []struct {
		RegOffset *uint64 `toml:"reg-offset"`
		SliceStartIdx *uint64 `toml:"slice-start-idx"`
		SliceWidth *uint64 `toml:"slice-width"`
		BfName string `toml:"bf-name"`
	} `toml:"bitfield-reference"`
}

type Regmap struct {
	ModuleName string `toml:"module-name"`
	RegisterWidth *uint64 `toml:"register-width"`
	Bitfields []Bitfield `toml:"bitfield"`
	Registers []Register `toml:"register"`
}

// Main function bind the processText to JS and keep the program alive
func main() {
	// Parse commandline argument using flags
	localTestFlag := flag.Bool("test", false, "Run local test")

	// Crucial: must call flag.Parse() to executre the parsing
	flag.Parse()

	// Local test
	if *localTestFlag == true {
		fmt.Printf("[INFO] Running local test mode.\n")
		LocalTest()
	} else {
		fmt.Printf("[INFO] Running web-demo mode.\n")
		JsBind()
		select {}	// Block forever to keep instance alive
	}
}

