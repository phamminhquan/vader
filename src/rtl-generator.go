package main

import (
	"fmt"
	"log"
	"os"
)

// Function to write SV file using info in regmap
func GenRTL(filePath string, regmap Regmap) {
	// Create SV file (or overwrite)
	svFile, err := os.Create(filePath)
	if err != nil {
		log.Fatalf("Failed to create SV file: %v", err)
		panic(err)
		return
	}
	defer svFile.Close() // Automatically closes SV file when main finishes
	
	// Writing to SV file
	fmt.Printf("Writing to SV file: regmap.sv")
	fmt.Fprintf(svFile, "// Auto-generated\n")
	fmt.Fprintf(svFile, "module %s (\n", regmap.ModuleName)
	fmt.Fprintf(svFile, "  // Clock and reset,\n")
	fmt.Fprintf(svFile, "  input I_clk,\n")
	fmt.Fprintf(svFile, "  input I_rst_n,\n\n")
	fmt.Fprintf(svFile, "  // Bitfields\n")
	for _, bf := range regmap.Bitfields {
		if bf.Access == "RW" {
			fmt.Fprintf(svFile, "  output [%d:0] O_%s,\n", *bf.Width-1, bf.Name)
		} else if bf.Access == "R" {
			fmt.Fprintf(svFile, "  input  [%d:0] I_%s,\n", *bf.Width-1, bf.Name)
		} else {
			log.Fatalf("Do not recognize access type %s for bitfield %s " + 
				"(supported access types are RW and R)", bf.Access, bf.Name)
			return
		}
	}
	fmt.Fprintf(svFile, "\n  // APB bus,\n")
	fmt.Fprintf(svFile, "  input  [31:0]  I_paddr,\n")
	fmt.Fprintf(svFile, "  input  [2:0]   I_pprot,\n")
	fmt.Fprintf(svFile, "  input          I_psel,\n")
	fmt.Fprintf(svFile, "  input          I_penable,\n")
	fmt.Fprintf(svFile, "  input          I_pwrite,\n")
	fmt.Fprintf(svFile, "  input  [31:0]  I_pwdata,\n")
	fmt.Fprintf(svFile, "  input  [3:0]   I_pstrb,\n")
	fmt.Fprintf(svFile, "  output         O_pready,\n")
	fmt.Fprintf(svFile, "  output         O_pslverr,\n")
	fmt.Fprintf(svFile, "  output [31:0]  O_prdata\n")
	fmt.Fprintf(svFile, ");")
}
