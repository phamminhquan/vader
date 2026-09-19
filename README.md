# About
This repository contains a Golang based register map RTL generator where:
* Register map management file is TOML format
* Golang based register map validation and RTL generation

# Web Demo
I deployed a [Github IO page](phamminhquan.github.io/vader) for this repository
as a demo. In the demo, you can write your own register map TOML file. There is
a "EXECUTION LOGS" textbox that will display the TOML validation result to make sure
that your TOML file adhere to the correct format, which I will mention in the
following section. If your TOML passed the validation, the RTL will be generated
and displayed in the "GENERATED RTL" textbox.

# TOML format
I've provided an example TOML file in `example/example-regmap.toml`. Here are
the required fields:

* `module-name`: name of register module
* `register-width`: bitwidth of the registers in this module
* `bitfield`: an array of tables for bitfields
  * `name`: name of bitfield
  * `width`: bitwidth of bitfield
  * `default-value`: value upon reset
  * `access`: APB access permission, RW (read/write), RO (read-only), WO (write-only)
* `register`: an array of tables for registers
  * `name`: name of register
  * `address`: address of register for APB access
  * `bitfield-reference`: array of bitfields that is in this register
    * `reg-offset`: the index of the register at which this bitfield reference starts
    * `slice-start-idx`: the index of the bitfield at which this reference starts
    * `slice-width`: the bitwidth of this bitfield reference
    * `bf-name`: name of the bitfield it's referencing

# Purpose
This repository is my proof of concept for using a simple text format like TOML
for register map management, which provides high readability and
maintainability.

At the same, this is my first time learning Golang and I found that it is faster
and more enjoyable to write than Python (which I would have used in the past
for stuff like this).

# TODO
* Create a CLI tool-set (binaries) that would do the functionality in terminal
