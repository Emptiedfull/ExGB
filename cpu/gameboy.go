package cpu

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

const (
	TCYCLES_PER_FRAME = 70224
	MCYCLES_PER_FRAME = 17556
)

type Gameboy struct {
	cpu    *Cpu
	ppu    *Ppu
	memory Memory

	Screen       [160][144][3]byte
	tileScanline [160]uint8

	clock   Clock
	logger  *log.Logger
	logFile *os.File

	halted  bool
	haltbug bool

	Soc *websocket.Conn
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

func GBInit(conn *websocket.Conn) *Gameboy {
	gb := &Gameboy{
		cpu:     &Cpu{},
		memory:  Memory{mem: make([]byte, 65536)},
		clock:   Clock{},
		halted:  false,
		haltbug: false,
		Soc:     conn,
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

	// logFile, err := os.OpenFile("gbabot.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	// if err != nil {
	// 	fmt.Println("Error opening log file:", err)
	// } else {
	// 	gb.logger = log.New(logFile, "", 0)
	// 	gb.logFile = logFile
	// }

	gb.memory.Init(65536)

	opcodes = gb.initOpcodes()
	gb.SetPostBootState()

	gb.cpu.pc = 0x0100

	// Initialize the PPU
	gb.ppu = &Ppu{
		Scanline:       0,
		ScalineCounter: 0,
	}

	return gb

}

func (g *Gameboy) SetPostBootState() {
	// Set CPU registers to post-boot state
	g.cpu.registers.a = 0x01
	g.cpu.registers.f.setFromUint8(0xB0)
	g.cpu.registers.b = 0x00
	g.cpu.registers.c = 0x13
	g.cpu.registers.d = 0x00
	g.cpu.registers.e = 0xD8
	g.cpu.registers.h = 0x01
	g.cpu.registers.l = 0x4D
	g.cpu.registers.sp = 0xFFFE
	g.cpu.pc = 0x0100

	g.WriteAddr(0xFF05, 0x00) // TIMA
	g.WriteAddr(0xFF06, 0x00) // TMA
	g.WriteAddr(0xFF07, 0x00) // TAC
	g.WriteAddr(0xFF10, 0x80) // NR10
	g.WriteAddr(0xFF11, 0xBF) // NR11
	g.WriteAddr(0xFF12, 0xF3) // NR12
	g.WriteAddr(0xFF14, 0xBF) // NR14
	g.WriteAddr(0xFF16, 0x3F) // NR21
	g.WriteAddr(0xFF17, 0x00) // NR22
	g.WriteAddr(0xFF19, 0xBF) // NR24
	g.WriteAddr(0xFF1A, 0x7F) // NR30
	g.WriteAddr(0xFF1B, 0xFF) // NR31
	g.WriteAddr(0xFF1C, 0x9F) // NR32
	g.WriteAddr(0xFF1E, 0xBF) // NR33
	g.WriteAddr(0xFF20, 0xFF) // NR41
	g.WriteAddr(0xFF21, 0x00) // NR42
	g.WriteAddr(0xFF22, 0x00) // NR43
	g.WriteAddr(0xFF23, 0xBF) // NR44
	g.WriteAddr(0xFF24, 0x77) // NR50
	g.WriteAddr(0xFF25, 0xF3) // NR51
	g.WriteAddr(0xFF26, 0xF1) // NR52
	g.WriteAddr(0xFF40, 0x91) // LCDC - LCD on, BG on, sprites off, BG tilemap 9800-9BFF
	g.WriteAddr(0xFF42, 0x00) // SCY
	g.WriteAddr(0xFF43, 0x00) // SCX
	g.WriteAddr(0xFF45, 0x00) // LYC
	g.WriteAddr(0xFF47, 0xFC) // BGP - Background palette
	g.WriteAddr(0xFF48, 0xFF) // OBP0 - Sprite palette 0
	g.WriteAddr(0xFF49, 0xFF) // OBP1 - Sprite palette 1
	g.WriteAddr(0xFF4A, 0x00) // WY
	g.WriteAddr(0xFF4B, 0x00) // WX
	g.WriteAddr(0xFFFF, 0x00) // IE - Interrupt enable

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

func (g *Gameboy) Start(done chan bool, manual chan bool) {
	g.clock.totalMcycles = 0
	g.clock.totalTcycles = 0

	ticker := time.NewTicker(time.Second / 6)

	if manual != nil {
		ticker.Stop()
	}

	g.memory.writeAddr(LCDC, 0x91) // Enable LCD, set mode to 0, enable sprites, enable background
	frame := 0
	for {
		select {
		case <-done:
			ticker.Stop()
			return
		case <-manual:
			i := MCYCLES_PER_FRAME

			for g.memory.mem[LY] <= 154 {
				cycles := g.DebugStep()
				g.UpdateClock(cycles)
				g.UpdateGraphics(cycles)
				i -= cycles
			}
			fmt.Println("Manual step completed, frame:", frame)
			frame++

			if frame%60 == 0 {
				frame = 0
			}
		case <-ticker.C:
			g.clock.totalTcycles += TCYCLES_PER_FRAME
			g.clock.totalMcycles += MCYCLES_PER_FRAME

			i := MCYCLES_PER_FRAME

			for i > 0 {
				cycles := g.DebugStep()
				g.UpdateClock(cycles)
				g.UpdateGraphics(cycles)
				i -= cycles
			}

			g.StreamScreen()

		}
	}

}

func (g *Gameboy) StreamScreen() {
	if g.Soc == nil {
		// fmt.Println("WebSocket connection is not initialized.")
		return
	}

	data := g.getPixelData()
	if err := g.Soc.WriteJSON(data); err != nil {
		log.Println("Error sending pixel data:", err)
	}
}

type PixelData struct {
	Height int    `json:"height"`
	Width  int    `json:"width"`
	Pixels []byte `json:"pixels"`
}

func (g *Gameboy) getPixelData() PixelData {
	pixels := make([]byte, 160*144*3) // 160x144 pixels, 3 bytes per pixel (RGB)
	// fmt.Println(g.Screen)
	for y := 0; y < 144; y++ {
		for x := 0; x < 160; x++ {
			index := (y*160 + x) * 3
			pixels[index] = g.Screen[x][y][0]
			pixels[index+1] = g.Screen[x][y][1]
			pixels[index+2] = g.Screen[x][y][2]
		}
	}

	data := PixelData{
		Height: 144,
		Width:  160,
		Pixels: pixels,
	}

	return data

}
