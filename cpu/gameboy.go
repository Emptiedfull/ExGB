package cpu

import (
	"fmt"
	"log"
	"os"
	"time"
)

const (
	TCYCLES_PER_FRAME = 70224
	MCYCLES_PER_FRAME = 17556
)

type Gameboy struct {
	cpu     *Cpu
	ppu     *Ppu
	memory  Memory
	clock   Clock
	logger  *log.Logger
	logFile *os.File

	halted  bool
	haltbug bool
}

func (g *Gameboy) ReadAddr(Addr uint16) uint8 {
	return g.memory.readAddr(Addr)
}
func (g *Gameboy) WriteAddr(Addr uint16, val uint8) {
	g.memory.writeAddr(Addr, val)
}

func (g *Gameboy) GetCpuPc() uint16 {
	return g.cpu.pc
}

type Clock struct {
	totalMcycles int
	totalTcycles int
}

func GBInit() *Gameboy {
	gb := &Gameboy{
		cpu:     &Cpu{},
		memory:  Memory{mem: make([]byte, 65536)},
		clock:   Clock{},
		halted:  false,
		haltbug: false,
	}

	gb.memory.Init(65536)

	opcodes = gb.initOpcodes()

	gb.cpu.registers.sp = INITIAL_SP
	gb.cpu.pc = INTIAL_PC

	return gb
}

func GBInitDebug() *Gameboy {
	gb := &Gameboy{
		cpu:     &Cpu{},
		memory:  Memory{mem: make([]byte, 65536)},
		clock:   Clock{},
		halted:  false,
		haltbug: false,
	}

	logFile, err := os.OpenFile("gbabot.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Println("Error opening log file:", err)
	} else {
		gb.logger = log.New(logFile, "", 0)
		gb.logFile = logFile
	}

	gb.memory.Init(65536)

	opcodes = gb.initOpcodes()

	gb.cpu.registers.a = 0x01
	gb.cpu.registers.b = 0x00
	gb.cpu.registers.c = 0x13
	gb.cpu.registers.d = 0x00
	gb.cpu.registers.e = 0xD8
	gb.cpu.registers.f.setFromUint8(0xB0)
	gb.cpu.registers.h = 0x01
	gb.cpu.registers.l = 0x4D
	gb.cpu.registers.sp = 0xFFFE
	gb.cpu.pc = 0x0100

	return gb

}

func (g *Gameboy) LOADROM(rom []byte) {
	g.memory.LoadROM(rom)
	fmt.Println("ROM loaded successfully")
}

func (g *Gameboy) fetch() byte {
	value := g.memory.readAddr(g.cpu.pc)
	g.cpu.pc++
	return value
}

func (g *Gameboy) fetch16() uint16 {
	low := g.fetch()
	high := g.fetch()
	return uint16(high)<<8 | uint16(low)
}

func (g *Gameboy) Start(done chan bool) {
	g.clock.totalMcycles = 0
	g.clock.totalTcycles = 0

	ticker := time.NewTicker(time.Second / 60)

	for {
		select {
		case <-done:
			ticker.Stop()
			return
		case <-ticker.C:
			g.clock.totalTcycles += TCYCLES_PER_FRAME
			g.clock.totalMcycles += MCYCLES_PER_FRAME

			i := MCYCLES_PER_FRAME

			for i > 0 {
				i = i - g.Step()
			}

		}
	}

}
