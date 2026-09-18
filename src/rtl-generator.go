package main

import (
	"fmt"
)

// Function to write SV file using info in regmap
func GenRTL(regmap *Regmap) string {
	// Create file string
	var rtlString string

	// Writing to SV file
	rtlString += fmt.Sprintf("// Auto-generated RTL\n")
	rtlString += fmt.Sprintf("module %s (\n", regmap.ModuleName)
	rtlString += fmt.Sprintf("  // Clock and reset,\n")
	rtlString += fmt.Sprintf("  input I_clk,\n")
	rtlString += fmt.Sprintf("  input I_rst_n,\n\n")
	rtlString += fmt.Sprintf("  // Bitfields\n")
	for _, bf := range regmap.Bitfields {
		if bf.Access == "RW" {
			if *bf.Width == 1 {
				rtlString += fmt.Sprintf("  output        O_%s,\n", bf.Name)
			} else {
				rtlString += fmt.Sprintf("  output [%d:0] O_%s,\n", *bf.Width-1, bf.Name)
			}
		} else if bf.Access == "RO" {
			if *bf.Width == 1 {
				rtlString += fmt.Sprintf("  input         I_%s,\n", bf.Name)
			} else {
				rtlString += fmt.Sprintf("  input  [%d:0] I_%s,\n", *bf.Width-1, bf.Name)
			}
		} else {
			return fmt.Sprintf("Do not recognize access type %s for bitfield %s " +
				"(supported access types are RW and R)", bf.Access, bf.Name)
		}
	}
	rtlString += fmt.Sprintf("\n  // APB bus,\n")
	rtlString += fmt.Sprintf("  input  [31:0]  I_paddr,\n")
	rtlString += fmt.Sprintf("  input  [2:0]   I_pprot,\n")
	rtlString += fmt.Sprintf("  input          I_psel,\n")
	rtlString += fmt.Sprintf("  input          I_penable,\n")
	rtlString += fmt.Sprintf("  input          I_pwrite,\n")
	rtlString += fmt.Sprintf("  input  [31:0]  I_pwdata,\n")
	rtlString += fmt.Sprintf("  input  [3:0]   I_pstrb,\n")
	rtlString += fmt.Sprintf("  output         O_pready,\n")
	rtlString += fmt.Sprintf("  output         O_pslverr,\n")
	rtlString += fmt.Sprintf("  output [31:0]  O_prdata\n")
	rtlString += fmt.Sprintf(");\n")

	// TODO: Add APB logic
	
	// TODO: Add address decoder logic (registers)
	
	// TODO: add output assignment

	rtlString += fmt.Sprintf("endmodule\n")
	
	return rtlString
}
