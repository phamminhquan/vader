package main

import (
	"fmt"
	"os"
	"github.com/pelletier/go-toml/v2"
)

// Function to test locally
func LocalTest(tomlPath string) {
	// Read example toml
	fileBytes, err := os.ReadFile(tomlPath)
	if err != nil {
		fmt.Printf("[FATAL] Failed to read TOML file: %v", err)
		return
	}

	// Unmarshal: parse into struct
	var regmap Regmap
	err = toml.Unmarshal(fileBytes, &regmap)
	if err != nil {
		fmt.Printf("[FATAL] Failed to unmarshal TOML: %v", err)
		return
	}

	// Validate the TOML input (validator.go)
	valStatus, valResult := Validate(&regmap)
	if valStatus == true {
		fmt.Printf("[ERROR] Validation Failed.\n")
		fmt.Printf("%s", valResult)
		return
	}
	
	// Generate RTL
	fmt.Printf("%s", GenRTL(&regmap))
	return	
}
