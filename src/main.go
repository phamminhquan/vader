package main

import (
	"fmt"
	"syscall/js"

	"github.com/pelletier/go-toml/v2"
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
	Repetition struct {
		Times *uint64 `toml.first-index`
		AddressIncrement *uint64 `toml:"address-increment"`
	} `toml:"repetition"`
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

// Function to process input textbox, we expose this function to JS
func processText(this js.Value, args []js.Value) any {
	// Get input from textbox
	if len(args) < 1 {
		return ""
	}
	txtInputStr := args[0].String()

	// Convert input text from string to bytes
	txtInputBytes := []byte(txtInputStr)

	// Unmarshal: parse into struct
	var regmap Regmap
	err := toml.Unmarshal(txtInputBytes, &regmap)
	if err != nil {
		// Return error to log textbox
		return js.ValueOf(map[string]any {
			"log": fmt.Sprintf("[FATAL] Failed to unmarshal toml file: %v", err),
			"result": "Fatal TOML.",
		})
	}

	// Validate the TOML input (validator.go)
	valStatus, valResult := Validate(&regmap)
	if valStatus == true {
		// Return error to log textbox
		return js.ValueOf(map[string]any {
			"log": fmt.Sprintf("%s", valResult),
			"result": "Error TOML",
		})
	}
	
	// Generate RTL
	rtlGenResult := GenRTL(&regmap)
	
	// Return error to log textbox
	return js.ValueOf(map[string]any {
		"log": fmt.Sprintf("%s", valResult),
		"result": rtlGenResult,
	})
}

// Main function bind the processText to JS and keep the program alive
func main() {
	// Expose the Go function to the browser window object
	js.Global().Set("goProcessText", js.FuncOf(processText))

	// Block forever to keep instance alive
	select {}
}

