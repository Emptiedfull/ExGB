package main

import (
	"fmt"
	"os"

	"gbabot/cpu"
)

func main() {
	gb := cpu.GBInitDebug()
	rompath := "./cpu/individual/timing.gb"
	romData, err := os.ReadFile(rompath)
	if err != nil {
		fmt.Println("Error loading ROM:", err)
		return
	}
	gb.LOADROM(romData)
	fmt.Printf("Game Boy initialized: %+v\n", gb != nil)

	// Track PC to detect infinite loops
	// pcHistory := make(map[uint16]int)

	for { // Reduced limit for debugging
		// pc := gb.GetCpuPc()

		// // Track how many times we've been at this PC
		// pcHistory[pc]++

		// //If we've been at the same PC too many times, we're in a loop
		// if pcHistory[pc] > 1000 {
		// 	fmt.Printf("INFINITE LOOP DETECTED at PC=0x%04X after %d steps\n", pc, i)
		// 	fmt.Printf("Last few instructions:\n")
		// 	for j := 0; j < 200; j++ {
		// 		opcode := gb.ReadAddr(pc)
		// 		fmt.Printf("PC: 0x%04X, Opcode: 0x%02X\n", pc, opcode)
		// 		gb.Step()
		// 		pc = gb.GetCpuPc()
		// 	}
		// 	break
		// }

		gb.DebugStep()

		// Check for test completion every 1000 steps

		// Print any serial output immediately
		if serial := gb.ReadAddr(0xFF01); serial != 0 {
			fmt.Printf("%c", serial)
			gb.WriteAddr(0xFF01, 0)
		}
	}
}
