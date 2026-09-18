package main

import (
	"fmt"
	//"log"
	//"os"
	//"sync"
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

func main() {
	//// Read raw toml file
	//fileBytes, err := os.ReadFile("example/example-regmap.toml")
	//if err != nil {
	//	log.Fatalf("[FATAL] Failed to read file: %v", err)
	//	panic(err)
	//}
	//fmt.Printf("[INFO] Read raw TOML file: regmap.toml.\n")
	//
	//// Unmarshal: parse into struct
	//var regmap Regmap
	//err = toml.Unmarshal(fileBytes, &regmap)
	//if err != nil {
	//	log.Fatalf("[FATAL] Failed to unmarshal toml file: %v", err)
	//	panic(err)
	//}
	//fmt.Printf("[INFO] Unmarshal TOML file to struct.\n")

	//// Assign ID to bitfields as order in the array
	//for id, _ := range regmap.Bitfields {
	//	regmap.Bitfields[id].ID = id
	//}

	//// Assign ID to registers as order in the array
	//for id, _ := range regmap.Registers {
	//	regmap.Registers[id].ID = id
	//}

	// Validate the TOML input (validator.go)
	//Validate(&regmap)
	
	//// Print back the toml
	//fmt.Printf("Print back data in TOML.\n")
	//fmt.Printf("ModuleName: %s\n", regmap.ModuleName)
	//for _, bf := range regmap.Bitfields {
	//	fmt.Printf("Bitfield:\n")
	//	fmt.Printf("\tID: %d\n", bf.ID)
	//	fmt.Printf("\tName: %s\n", bf.Name)
	//	fmt.Printf("\tWidth: %d\n", *bf.Width)
	//	fmt.Printf("\tDefault Value: 0b%0*b\n", *bf.Width, *bf.DefaultValue)
	//	fmt.Printf("\tAccess: %s\n", bf.Access)
	//}

	//for _, reg := range regmap.Registers {
	//	fmt.Printf("Register:\n")
	//	fmt.Printf("\tID: %d\n", reg.ID)
	//	fmt.Printf("\tName: %s\n", reg.Name)
	//	fmt.Printf("\tAddress: 0x%08x\n", *reg.Address)
	//	fmt.Printf("\tRepetition:")
	//	fmt.Printf("\tTimes: %d", *reg.Repetition.Times)
	//	fmt.Printf("\tAddress Increment: 0x%08x\n", *reg.Repetition.AddressIncrement)
	//	fmt.Printf("\tBitfield Reference:\n")
	//	for _, ref := range reg.BitfieldReference {
	//		fmt.Printf("\t\tRegister Offset: %d", *ref.RegOffset)
	//		fmt.Printf("\tSlice Start Index: %d", *ref.SliceStartIdx)
	//		fmt.Printf("\tSlice Width: %d", *ref.SliceWidth)
	//		fmt.Printf("\tBitfield Name: %s\n", ref.BfName)
	//	}
	//}
	
	// Generate RTL (rtl-generator.go)
	//GenRTL("regmap.sv", regmap)
	

	// Expose the Go function to the browser window object
	js.Global().Set("goProcessText", js.FuncOf(processText))

	// Block forever to keep instance alive
	select {}
	
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
			"result": "Invalid TOML.",
		})
	}

	// TODO: Validate toml text
	// Validate the TOML input (validator.go)
	valResult := Validate(&regmap)
	
	return js.ValueOf(map[string]any {
		"log": valResult,
		"result": "[INFO] Successfully unmarshal toml file.",
	})
}
