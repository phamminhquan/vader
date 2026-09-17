package main

import (
	"fmt"
	"log"
	"os"
	"sync"
	
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
	// Read raw toml file
	fileBytes, err := os.ReadFile("example-regmap.toml")
	if err != nil {
		log.Fatalf("[FATAL] Failed to read file: %v", err)
		panic(err)
	}
	fmt.Printf("[INFO] Read raw TOML file: regmap.toml.\n")
	
	// Unmarshal: parse into struct
	var regmap Regmap
	err = toml.Unmarshal(fileBytes, &regmap)
	if err != nil {
		log.Fatalf("[FATAL] Failed to unmarshal toml file: %v", err)
		panic(err)
	}
	fmt.Printf("[INFO] Unmarshal TOML file to struct.\n")

	// Assign ID to bitfields as order in the array
	for id, _ := range regmap.Bitfields {
		regmap.Bitfields[id].ID = id
	}

	// Assign ID to registers as order in the array
	for id, _ := range regmap.Registers {
		regmap.Registers[id].ID = id
	}

	// Make error channel and wait group for validation workers
	errChan := make(chan error, len(regmap.Bitfields) + len(regmap.Registers) + 3)
	var wg sync.WaitGroup

	// Launch module definition validation workers
	wg.Add(1)
	go ValidateModuleDefinition(&regmap, errChan, &wg)

	// Launch bitfield validation workers as goroutines
	for i, _ := range regmap.Bitfields {
		wg.Add(1)
		go ValidateBitfield(&regmap.Bitfields[i], errChan, &wg)
	}

	// Launch bitfield name overlap validation workers
	wg.Add(1)
	go ValidateBitfieldUnique(&regmap.Bitfields, errChan, &wg)

	// Launch register validation workers as goroutines
	for i, _ := range regmap.Registers {
		wg.Add(1)
		go ValidateRegister(&regmap.Registers[i], &regmap, errChan, &wg)
	}

	// Launch register address overlap validation workers
	wg.Add(1)
	go ValidateRegisterAddressOverlap(&regmap, errChan, &wg)

	// Launch a background goroutine ONLY to close the channels
	go func() {
		wg.Wait()
		close(errChan) // Safely breaks the loop in error collection
	}()

	// Collecting errors
	var collectedErrors []error
	for err = range errChan { // This loop runs (blocking) till errChan is closed
		if err != nil {
			collectedErrors = append(collectedErrors, err)
		}
	}

	// Print error and exit if there is error
	if len(collectedErrors) > 0 {
		fmt.Printf("[ERROR] VALIDATION FAILED: %d bitfields have errors.\n",
			len(collectedErrors))
		// Loop through each error and print
		for _, err := range collectedErrors {
			fmt.Printf("%v\n", err)
		}
		// Stop the program and exit with error status code (non-zero)
		//os.Exit(1)
	} else {
		fmt.Printf("[INFO] VALIDATION PASSED.\n")
	}
	
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

	for _, reg := range regmap.Registers {
		fmt.Printf("Register:\n")
		fmt.Printf("\tID: %d\n", reg.ID)
		fmt.Printf("\tName: %s\n", reg.Name)
		fmt.Printf("\tAddress: 0x%08x\n", *reg.Address)
		fmt.Printf("\tRepetition:")
		fmt.Printf("\tTimes: %d", *reg.Repetition.Times)
		fmt.Printf("\tAddress Increment: 0x%08x\n", *reg.Repetition.AddressIncrement)
		fmt.Printf("\tBitfield Reference:\n")
		for _, ref := range reg.BitfieldReference {
			fmt.Printf("\t\tRegister Offset: %d", *ref.RegOffset)
			fmt.Printf("\tSlice Start Index: %d", *ref.SliceStartIdx)
			fmt.Printf("\tSlice Width: %d", *ref.SliceWidth)
			fmt.Printf("\tBitfield Name: %s\n", ref.BfName)
		}
	}
	
	// Write the SV code
	//WriteSv("regmap.sv", regmap)
	
}

