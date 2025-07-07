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

	gb.cpu.pc = 0x0000

	// Initialize the PPU
	gb.ppu = &Ppu{
		Scanline:       0,
		ScalineCounter: 0,
	}

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

	g.memory.writeAddr(LCDC, 0x91) // Enable LCD, set mode to 0, enable sprites, enable background

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
