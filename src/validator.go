package main

import (
	"fmt"
	"errors"
	"sync"
	"slices"
	"strings"
)

// Function that calls all the validation workers
// Should be sequentially called in main
func Validate(regmap *Regmap) string {
	// Make error channel and wait group for validation workers
	errChan := make(chan error, len(regmap.Bitfields) + len(regmap.Registers) + 3)
	var wg sync.WaitGroup

	// Launch module definition validation workers
	wg.Add(1)
	go ValidateModuleDefinition(regmap, errChan, &wg)

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
		go ValidateRegister(&regmap.Registers[i], regmap, errChan, &wg)
	}

	// Launch register address overlap validation workers
	wg.Add(1)
	go ValidateRegisterAddressOverlap(regmap, errChan, &wg)

	// Launch a background goroutine ONLY to close the channels
	go func() {
		wg.Wait()
		close(errChan) // Safely breaks the loop in error collection
	}()

	// Collecting errors
	var collectedErrors []error
	for err := range errChan { // This loop runs (blocking) till errChan is closed
		if err != nil {
			collectedErrors = append(collectedErrors, err)
		}
	}

	// Print error and exit if there is error
	if len(collectedErrors) > 0 {
		valResult := make([]string, len(collectedErrors) + 1)
		valResult[0] = fmt.Printf("[ERROR] VALIDATION FAILED: %d" +
			"bitfields have errors.\n", len(collectedErrors))
			copy(valResult[1:0], colelctedErrors)
		// Stop the program and exit with error status code (non-zero)
		//os.Exit(1)
		return strings.Join(valResult, "")
	} else {
		return fmt.Sprintf("[INFO] VALIDATION PASSED.\n")
	}
}

// Function to validate module definition
func ValidateModuleDefinition(regmap *Regmap, errChan chan <- error, wg *sync.WaitGroup) {
	// Execute when function returns
	defer wg.Done()

	// Create a slice to collect all errors for this bitfield
	var err []error

	// Rule: Module Name exits check
	if regmap.ModuleName == "" {
		err = append(err, fmt.Errorf("[ERROR] Module definition:" +
			" module name is empty or not explicitly declared."))
	}

	// Rule: Register Width exist check
	if regmap.RegisterWidth == nil || *regmap.RegisterWidth == 0 {
		err = append(err, fmt.Errorf("[ERROR] Module definition:" +
			" register width is 0 or not explicitly declared."))
	}

	// Send errors back through channel
	if len(err) > 0 {
		errChan <- errors.Join(err...)
	} else {
		errChan <- nil
	}
}

// Function to validate each bitfield
func ValidateBitfield(bf *Bitfield, errChan chan <- error, wg *sync.WaitGroup) {
	// Execute when function returns
	defer wg.Done()

	// Create a slice to collect all errors for this bitfield
	var err []error

	// Rule: Name exist check
	if bf.Name == "" {
		err = append(err, fmt.Errorf("[ERROR] Bitfield ID %d:" +
			" name is empty or not explicitly declared.", bf.ID))
	}

	// Rule: Width exist check
	if bf.Width == nil {
		if bf.Name == "" {
			err = append(err, fmt.Errorf("[ERROR] Bitfield ID %d:" +
				" width is not explicitly declared.", bf.ID))
		} else {
			err = append(err, fmt.Errorf("[ERROR] Bitfield name %s:" +
				" width is not explicitly declared.", bf.Name))
		}
	}

	// Rule: DefaultValue exist check
	// DefaultValue is intentionally declared as a pointer so that if it is not
	// explicitly declared, it will have nil value, which can be checked
	if bf.DefaultValue == nil {
		if bf.Name == "" {
			err = append(err, fmt.Errorf("[ERROR] Bitfield ID %d: " +
				"default value is not explicitly declared.", bf.ID))
		} else {
			err = append(err, fmt.Errorf("[ERROR] Bitfield name %s: " +
				"default value is not explicitly declared.", bf.Name))
		}
	}

	// Rule: Access only has 3 possible values
	// RO (read-only), WO (write-only), RW (read/write)
	if bf.Access != "RO" && bf.Access != "WO" && bf.Access != "RW" {
		if bf.Name == "" {
			err = append(err, fmt.Errorf("[ERROR] Bitfield ID %d: " +
				"access is not RO (read-only), WO (write-only), or RW (read/write).",
				bf.ID))
		} else {
			err = append(err, fmt.Errorf("[ERROR] Bitfield name %s: " +
				"access is not RO (read-only), WO (write-only), or RW (read/write).",
				bf.Name))
		}
	}

	// Rule: Width is between 1 and 64
	// 64 is upper limit for now since Go only has upto 64-bit integer type
	if bf.Width != nil && (*bf.Width < 1 || *bf.Width > 64) {
		if bf.Name == "" {
			err = append(err, fmt.Errorf("[ERROR] Bitfield ID %d: " +
				"width is not within range of 1 and 64.", bf.ID))
		} else {
			err = append(err, fmt.Errorf("[ERROR] Bitfield name %s: " +
				"width is not within range of 1 and 64.", bf.Name))
		}
	}

	// Rule: DefaultValue must fit within the bitwidth specified by Width
	if bf.DefaultValue != nil && bf.Width != nil &&
	(*bf.DefaultValue > (1 << *bf.Width) - 1) {
		if bf.Name == "" {
			err = append(err, fmt.Errorf("[ERROR] Bitfield ID %d: " +
				"default value exceed maximum allowed by width.", bf.ID))
		} else {
			err = append(err, fmt.Errorf("[ERROR] Bitfield name %s: " +
				"default value exceed maximum allowed by width.", bf.Name))
		}
	}

	// Send errors back through channel
	if len(err) > 0 {
		errChan <- errors.Join(err...)
	} else {
		errChan <- nil
	}
}

// Function to validate bitfield name overlap
func ValidateBitfieldUnique(bitfields *[]Bitfield, errChan chan <- error, wg *sync.WaitGroup) {
	// Execute when function returns
	defer wg.Done()

	// Create a slice to collect all errors for this bitfield
	var err []error

	// Rule: bitfields names must not overlap
	// ie 2 or more bitfields with the same name
	// Collect all the bitfields names and how many times it has been seen
	counts := make(map[string]int)
	for _, bf := range *bitfields {
		counts[bf.Name]++
	}

	// Error the repeated names
	for name, count := range counts {
		if count > 1 {
			err = append(err, fmt.Errorf("[ERROR] Bitfield name %s" +
				" is repeated %d times.", name, count))
		}
	}

	// Send errors back through channel
	if len(err) > 0 {
		errChan <- errors.Join(err...)
	} else {
		errChan <- nil
	}
}

// Function to validate each register
func ValidateRegister(reg *Register, regmap *Regmap, errChan chan <- error, wg *sync.WaitGroup) {
	// Execute when function returns
	defer wg.Done()

	// Create a slice to collect all errors for this bitfield
	var err []error

	// Grab some stuff in regmap that would be used for checks
	regWidth := regmap.RegisterWidth;
	bitfields := regmap.Bitfields;

	// Rule: Name exist check
	if reg.Name == "" {
		err = append(err, fmt.Errorf("[ERROR] Register ID %d:" +
			" name is empty or not explicitly declared.", reg.ID))
	}

	// Rule: Address exist check
	if reg.Address == nil {
		if reg.Name == "" {
			err = append(err, fmt.Errorf("[ERROR] Register ID %d:" +
				" address is not explicitly declared.", reg.ID))
		} else {
			err = append(err, fmt.Errorf("[ERROR] Register name %s:" +
				" width is not explicitly declared.", reg.Name))
		}
	}

	// Rule: Repetition Times exist check
	if reg.Repetition.Times == nil {
		if reg.Name == "" {
			err = append(err, fmt.Errorf("[ERROR] Register ID %d:" +
				" repetition times is not explicitly declared.", reg.ID))
		} else {
			err = append(err, fmt.Errorf("[ERROR] Register name %s:" +
				" repetition times is not explicitly declared.", reg.Name))
		}
	}

	// Rule: Address Increment exist check
	if reg.Repetition.AddressIncrement == nil {
		if reg.Name == "" {
			err = append(err, fmt.Errorf("[ERROR] Register ID %d:" +
				" repetition address increment is not explicitly declared.", reg.ID))
		} else {
			err = append(err, fmt.Errorf("[ERROR] Register name %s:" +
				" repetition address increment is not explicitly declared.", reg.Name))
		}
	}

	// Going through each bitfield reference
	for refIdx, ref := range reg.BitfieldReference {
		// Rule: Bitfield Reference does not match any defined Bitfields
		// Grab the bitfield corresponds to this bitfield reference
		var bitfield Bitfield
		if ref.BfName != "" {
			for _, bf := range bitfields {
				if bf.Name == ref.BfName {
					bitfield = bf
				}
			}
			if bitfield.Name == "" {
				if reg.Name == "" {
					err = append(err, fmt.Errorf("[ERROR] Register ID %d:" +
						" a bitfield reference name does not match any defined bitfields.",
						reg.ID))
				} else {
					err = append(err, fmt.Errorf("[ERROR] Register name %s:" +
						" a bitfield reference name does not match any defined bitfields.",
						reg.Name))
				}
			}
		}

		// Rule: Register Offset exist check
		if ref.RegOffset == nil {
			if reg.Name == "" {
				err = append(err, fmt.Errorf("[ERROR] Register ID %d:" +
					" a bitfield reference register offset is not explicitly declared.",
					reg.ID))
			} else {
				err = append(err, fmt.Errorf("[ERROR] Register name %s:" +
					" a bitfield reference register offset is not explicitly declared.",
					reg.Name))
			}
		}

		// Rule: Slice Start Index exist check
		if ref.SliceStartIdx == nil {
			if reg.Name == "" {
				err = append(err, fmt.Errorf("[ERROR] Register ID %d:" +
					" a bitfield reference slice start index is not explicitly declared.",
					reg.ID))
			} else {
				err = append(err, fmt.Errorf("[ERROR] Register name %s:" +
					" a bitfield reference slice start index is not explicitly declared.",
					reg.Name))
			}
		}

		// Rule: Slice Width exist check
		if ref.SliceWidth == nil || *ref.SliceWidth == 0 {
			if reg.Name == "" {
				err = append(err, fmt.Errorf("[ERROR] Register ID %d:" +
					" a bitfield reference slice width is 0 or not explicitly declared.",
					reg.ID))
			} else {
				err = append(err, fmt.Errorf("[ERROR] Register name %s:" +
					" a bitfield reference slice width is 0 or not explicitly declared.",
					reg.Name))
			}
		}

		// Rule: Bitfield Name exist check
		if ref.BfName == "" {
			if reg.Name == "" {
				err = append(err, fmt.Errorf("[ERROR] Register ID %d:" +
					" a bitfield reference name is empty or not explicitly declared.",
					reg.ID))
			} else {
				err = append(err, fmt.Errorf("[ERROR] Register name %s:" +
					" a bitfield reference name is empty or not explicitly declared.",
					reg.Name))
			}
		}

		// Rule: Register Offset + Slice Width exceed Register Width
		if ref.RegOffset != nil && ref.SliceWidth != nil && regWidth != nil && 
		*ref.RegOffset + *ref.SliceWidth > *regWidth {
			if reg.Name == "" {
				err = append(err, fmt.Errorf("[ERROR] Register ID %d:" +
					" a bitfield reference register offset + slice width exceeds" +
					" register width.", reg.ID))
			} else {
				err = append(err, fmt.Errorf("[ERROR] Register name %s:" +
					" a bitfield reference register offset + slice width exceeds" +
					" register width.", reg.Name))
			}
		}

		// Rule: Slice Start Index + Slice Width exceed bitfield width
		if ref.SliceStartIdx != nil && ref.SliceWidth != nil && bitfield.Name != "" &&
		*ref.SliceStartIdx > *bitfield.Width - 1 {
			if reg.Name == "" {
				err = append(err, fmt.Errorf("[ERROR] Register ID %d:" +
					" a bitfield reference slice start index + slice width exceeds" +
					" bitfield width.", reg.ID))
			} else {
				err = append(err, fmt.Errorf("[ERROR] Register name %s:" +
					" a bitfield reference slice start index + slice width exceeds" +
					" bitfield width.", reg.Name))
			}
		}

		// Rule: Bitfield reference overlap
		var startIdx0, endIdx0 uint64;
		if ref.RegOffset != nil && ref.SliceWidth != nil {
			// Get start and end index of this bitfield reference in the register
			startIdx0 = *ref.RegOffset
			endIdx0 = *ref.RegOffset + *ref.SliceWidth - 1
			// Going through each other bitfield reference skipping itself
			// to get start and end index in register
			for refIdx2 := refIdx + 1; refIdx2 < len(reg.BitfieldReference); refIdx2++ {
				ref2 := reg.BitfieldReference[refIdx2]
				if ref.BfName != ref2.BfName {
					var startIdx1, endIdx1 uint64;
					if ref2.RegOffset != nil && ref2.SliceWidth != nil {
						// Get start and end index
						startIdx1 = *ref2.RegOffset
						endIdx1 = *ref2.RegOffset + *ref2.SliceWidth - 1
						// Check if the ranges overlap (or don't overlap)
						if !(startIdx0 > endIdx1 || startIdx1 > endIdx0) {
							if reg.Name == "" {
								err = append(err, fmt.Errorf("[ERROR] Register ID %d:" +
									" there are overlapping bitfields in this register.", reg.ID))
							} else {
								err = append(err, fmt.Errorf("[ERROR] Register name %s:" +
									" there are overlapping bitfields in this register.", reg.Name))
							}
						}
					}
				}
			}
		}
	}

	// Send errors back through channel
	if len(err) > 0 {
		errChan <- errors.Join(err...)
	} else {
		errChan <- nil
	}
}

// Function to validate each register
func ValidateRegisterAddressOverlap(regmap *Regmap, errChan chan <- error,
wg *sync.WaitGroup) {
	// Execute when function returns
	defer wg.Done()

	// Create a slice to collect all errors for this bitfield
	var err []error

	// Use a map from address to an array of bitfield names
	overlap := make(map[uint64][]string)
	for _, reg := range regmap.Registers {
		if reg.Address != nil && reg.Repetition.Times != nil &&
		reg.Repetition.AddressIncrement != nil {
			var i uint64
			for i = 0; i < *reg.Repetition.Times + 1; i++ {
				addr := *reg.Address + i * (*reg.Repetition.AddressIncrement)
				if !slices.Contains(overlap[addr], reg.Name) {
					overlap[addr] = append(overlap[addr], reg.Name)
				}
			}
		}
	}

	// Error all the overlaps
	for addr, regNames := range overlap {
		if len(regNames) > 1 {
			err = append(err, fmt.Errorf("[ERROR] Address Overlap at 0x%08x for" +
				" registers %s.", addr, strings.Join(regNames, ", ")))
		}
	}

	// Send errors back through channel
	if len(err) > 0 {
		errChan <- errors.Join(err...)
	} else {
		errChan <- nil
	}
}
