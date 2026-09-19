package main

import (
	"fmt"
)

// File header string
const SvHeader string = "" +
"// Auto-generated RTL\n" +
"// This module contains a set of bitfields that can be connected and used\n" +
"// else where in the design. The bitfield values can be written to or read\n" +
"// from their associated memory-mapped registers via APB bus. Write data will\n" +
"// be valid 1 clock cycle after the APB transfer and read data is\n" +
"// combinational output based on address.\n\n"

// APB boiler plate logic.
const ApbLogic string = "" +
"//=============================================================================\n" +
"// APB boiler plate logic\n" +
"//=============================================================================\n" +
"\n" +
"// PREADY always 1 (no waited transactions)\n" +
"always_comb pready = 1'b1;\n" +
"always_comb pslverr = 1'b0;\n" +
"\n" +
"// Transfer flags when transfer is valid and if it its write request\n" +
"wire trans_valid = (psel & penable);\n" +
"wire wstrb = (trans_valid & pwrite);\n\n"

// Function to write SV file using info in regmap
func GenRTL(regmap *Regmap) string {
	// Create file string and other useful variables
	var rtlString string
	accessMap := make(map[string]string) // Bitfields struct maped to bitfield name

	// Writing to SV file
	rtlString += SvHeader
	rtlString += fmt.Sprintf("module %s (\n", regmap.ModuleName)
	rtlString += fmt.Sprintf("  // Clock and reset,\n")
	rtlString += fmt.Sprintf("  input I_clk,\n")
	rtlString += fmt.Sprintf("  input I_rst_n,\n\n")
	rtlString += fmt.Sprintf("  // Bitfields\n")
	for _, bf := range regmap.Bitfields {
		// Construct accessMap for later use
		accessMap[bf.Name] = bf.Access
		if bf.Access == "RW" {
			if *bf.Width == 1 {
				rtlString += fmt.Sprintf("  output reg O_%s,\n", bf.Name)
			} else {
				rtlString += fmt.Sprintf("  output reg [%d:0] O_%s,\n", *bf.Width-1, bf.Name)
			}
		} else if bf.Access == "RO" {
			if *bf.Width == 1 {
				rtlString += fmt.Sprintf("  input I_%s,\n", bf.Name)
			} else {
				rtlString += fmt.Sprintf("  input [%d:0] I_%s,\n", *bf.Width-1, bf.Name)
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
	rtlString += fmt.Sprintf(");\n\n")

	// Add APB boiler plate logic
	rtlString += ApbLogic
	
	// Add the actual register logic
	rtlString += fmt.Sprintf("// Registers and write logic\n")
	for _, reg := range regmap.Registers {
		// Grab the strings needed from each bitfield reference
		var bitfieldDefaultStr string
		var bitfieldWriteStr string
		for _, ref := range reg.BitfieldReference {
			if accessMap[ref.BfName] == "RW" || accessMap[ref.BfName] == "WO" {
				bitfieldDefaultStr += fmt.Sprintf("    O_%s[%d:%d] <= '0;\n",
					ref.BfName, *ref.SliceStartIdx + *ref.SliceWidth - 1, *ref.SliceStartIdx)
				var j uint64
				for j = 0; j < *ref.SliceWidth; j++ {
					bitfieldWriteStr += fmt.Sprintf("    O_%s[%d] <= pstrb[%d] & pwdata[%d]\n",
						ref.BfName, *ref.SliceStartIdx + j, (*ref.RegOffset + j) / 8,
						*ref.RegOffset + j)
				}
			}
		}

		// Register string
		rtlString += fmt.Sprintf("always_ff @(posedge clk or negedge rst_n) begin\n")
		rtlString += fmt.Sprintf("  if (!rst_n) begin\n")
		rtlString += bitfieldDefaultStr
		rtlString += fmt.Sprintf("  end else if (wstrb & (paddr == 32'h%08x)) begin\n", reg.Address)
		rtlString += bitfieldWriteStr
		rtlString += fmt.Sprintf("  end\n")
		rtlString += fmt.Sprintf("end\n")
	}
	
	// TODO: add output assignment

	rtlString += fmt.Sprintf("endmodule\n")
	
	return rtlString
}

