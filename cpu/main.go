package cpu

// func main() {
// 	GB := Gameboy{
// 		cpu:    &Cpu{},
// 		memory: Memory{mem: make([]uint8, 65536)},
// 		clock:  Clock{},
// 		halted: false,
// 	}

// 	GB.memory.Init(65536)

// 	GB.cpu.registers.sp = INITIAL_SP
// 	GB.cpu.pc = INTIAL_PC

// 	ROMFILE := "./individual/3.gb"
// 	romData, err := os.ReadFile(filepath.Clean(ROMFILE))
// 	if err != nil {
// 		panic(err)
// 	}

// 	GB.memory.LoadROM(romData)

// 	done := make(chan bool)

// 	GB.start(done)

// }
