package main

import (
	"fmt"
	"syscall/js"
	"github.com/pelletier/go-toml/v2"
)

// Function bind the processText to JS and keep the program alive
func JsBind() {
	// Expose the Go function to the browser window object
	js.Global().Set("goProcessText", js.FuncOf(processText))
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
