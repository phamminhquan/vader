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
"\n" +
"// Flag when transfer is a valid write request\n" +
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
	rtlString += fmt.Sprintf("  input  [31:0]      I_paddr,\n")
	rtlString += fmt.Sprintf("  input  [2:0]       I_pprot,\n")
	rtlString += fmt.Sprintf("  input              I_psel,\n")
	rtlString += fmt.Sprintf("  input              I_penable,\n")
	rtlString += fmt.Sprintf("  input              I_pwrite,\n")
	rtlString += fmt.Sprintf("  input  [31:0]      I_pwdata,\n")
	rtlString += fmt.Sprintf("  input  [3:0]       I_pstrb,\n")
	rtlString += fmt.Sprintf("  output reg         O_pready,\n")
	rtlString += fmt.Sprintf("  output reg         O_pslverr,\n")
	rtlString += fmt.Sprintf("  output reg [31:0]  O_prdata\n")
	rtlString += fmt.Sprintf(");\n\n")

	// Add APB boiler plate logic
	rtlString += ApbLogic
	
	// Add the actual register logic
	rtlString += "//=============================================================================\n"
	rtlString += "// Register and write logic\n"
	rtlString += "//=============================================================================\n\n"
	// Set up the strings needed for both write/read logic
	var prdataStr string
	var errorResponseStr string
	for _, reg := range regmap.Registers {
		// Grab the strings needed from each bitfield reference
		var bitfieldDefaultStr string
		var bitfieldWriteStr string
		var bitfieldReadStr string
		for _, ref := range reg.BitfieldReference {
			if accessMap[ref.BfName] == "RW" || accessMap[ref.BfName] == "WO" {
				if *ref.SliceWidth == 1 {
					bitfieldDefaultStr += fmt.Sprintf("    O_%s[%d] <= '0;\n",
						ref.BfName, *ref.SliceStartIdx)
				} else {
					bitfieldDefaultStr += fmt.Sprintf("    O_%s[%d:%d] <= '0;\n",
						ref.BfName, *ref.SliceStartIdx + *ref.SliceWidth - 1,
						*ref.SliceStartIdx)
				}
				var j uint64
				for j = 0; j < *ref.SliceWidth; j++ {
					bitfieldWriteStr += fmt.Sprintf("    O_%s[%d] <= pstrb[%d] & pwdata[%d]\n",
						ref.BfName, *ref.SliceStartIdx + j, (*ref.RegOffset + j) / 8,
						*ref.RegOffset + j)
				}
			}
			// Grab the strings needed from each bitfield reference for read
			if accessMap[ref.BfName] == "RW" || accessMap[ref.BfName] == "RO" {
				if *ref.SliceWidth == 1 {
					bitfieldReadStr += fmt.Sprintf("  rdata_%s[%d] = O_%s[%d];\n",
						reg.Name, *ref.RegOffset, ref.BfName, *ref.SliceStartIdx)
				} else {
					bitfieldReadStr += fmt.Sprintf("  rdata_%s[%d:%d] = O_%s[%d:%d];\n",
						reg.Name, *ref.RegOffset + *ref.SliceWidth - 1, *ref.RegOffset,
						ref.BfName, *ref.SliceStartIdx + *ref.SliceWidth - 1, *ref.SliceWidth)
				}
			}
		}

		// Register write string (skip if there is no bitfield to write)
		if bitfieldDefaultStr != "" {
			rtlString += fmt.Sprintf("// Register at address 32'h%08x\n", *reg.Address)
			rtlString += fmt.Sprintf("always_ff @(posedge clk or negedge rst_n) begin\n")
			rtlString += fmt.Sprintf("  if (!rst_n) begin\n")
			rtlString += bitfieldDefaultStr
			rtlString += fmt.Sprintf("  end else if (wstrb & (paddr == 32'h%08x)) begin\n", *reg.Address)
			rtlString += bitfieldWriteStr
			rtlString += fmt.Sprintf("  end\n")
			rtlString += fmt.Sprintf("end\n\n")
		}
		
		// Register read string (skip if there is no bitfield to read)
		rtlString += fmt.Sprintf("// Readback wire\n")
		rtlString += fmt.Sprintf("reg[31:0] rdata_%s;\n", reg.Name)
		rtlString += fmt.Sprintf("always_comb begin\n")
		rtlString += fmt.Sprintf("  rdata_%s = '0;\n", reg.Name)
		if bitfieldReadStr != "" {
			rtlString += bitfieldReadStr
		}
		rtlString += fmt.Sprintf("end\n\n")

		// Setup read data output assignment string
		prdataStr += fmt.Sprintf("    32'h%08x: prdata = rdata_%s;\n",
			*reg.Address, reg.Name)

		// Setup error handling string
		errorResponseStr += fmt.Sprintf("      32'h%08x: pslverr = 1'b0;\n",
			*reg.Address)
	}
	
	// Add output assignment
	rtlString += "//=============================================================================\n"
	rtlString += "// Readback data muxing\n"
	rtlString += "//=============================================================================\n\n"
	rtlString += fmt.Sprintf("always_comb begin\n")
	rtlString += fmt.Sprintf("  case (paddr)\n")
	rtlString += prdataStr
	rtlString += fmt.Sprintf("  default: prdata = '0;\n")
	rtlString += fmt.Sprintf("  endcase\n")
	rtlString += fmt.Sprintf("end\n\n")

	// PSLVERR handling
	rtlString += "//=============================================================================\n"
	rtlString += "// APB error response: transfer request for unmapped address\n"
	rtlString += "//=============================================================================\n\n"
	rtlString += "always_comb begin\n"
	rtlString += "  if (trans_valid) begin\n"
	rtlString += "    case (paddr)\n"
	rtlString += errorResponseStr
	rtlString += "    default: pslverr = 1'b1;\n"
	rtlString += "    endcase\n"
	rtlString += "  end\n"
	rtlString += "end\n\n"

	rtlString += fmt.Sprintf("endmodule\n")
	
	return rtlString
}

