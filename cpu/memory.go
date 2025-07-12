package cpu

import "fmt"

type Memory struct {
	mem           []uint8
	divCounter    uint16
	timerCounter  uint16
	timerFreq     uint16
	timeroverflow bool
}

const (
	INTERRUPT_ENABLE = 0xFFFF

	// INTERRUPT_VBLANK = 1 << 0
	// INTERRUPT_LCD    = 1 << 1
	// INTERRUPT_TIMER  = 1 << 2
	// INTERRUPT_SERIAL = 1 << 3
	// INTERRUPT_JOYPAD = 1 << 4

	INTERRUPT_FLAG = 0xFF0F
)

type INTERRUPT_ENABLE_MASK int
type INTERRUPT_FLAG_MASK int

const (
	INTERRUPT_VBLANK INTERRUPT_ENABLE_MASK = 1 << 0
	INTERRUPT_LCD    INTERRUPT_ENABLE_MASK = 1 << 1
	INTERRUPT_TIMER  INTERRUPT_ENABLE_MASK = 1 << 2
	INTERRUPT_SERIAL INTERRUPT_ENABLE_MASK = 1 << 3
	INTERRUPT_JOYPAD INTERRUPT_ENABLE_MASK = 1 << 4
)

const (
	INTERRUPT_FLAG_VBLANK INTERRUPT_FLAG_MASK = 1 << 0
	INTERRUPT_FLAG_LCD    INTERRUPT_FLAG_MASK = 1 << 1
	INTERRUPT_FLAG_TIMER  INTERRUPT_FLAG_MASK = 1 << 2
	INTERRUPT_FLAG_SERIAL INTERRUPT_FLAG_MASK = 1 << 3
	INTERRUPT_FLAG_JOYPAD INTERRUPT_FLAG_MASK = 1 << 4
)

const (
	DIV  = 0xFF04
	TIMA = 0xFF05
	TMA  = 0xFF06
	TAC  = 0xFF07
)

const (
	HRAM_START = 0xFF80
	HRAM_END   = 0xFFFE

	INITIAL_SP = 0xFFFE
	INTIAL_PC  = 0x0000

	ROM_BANK_0_START = 0x0000
	ROM_BANK_0_END   = 0x7FFF
)

func (m *Memory) Init(size int) {
	if size <= 0 {
		fmt.Println("Memory size must be greater than 0")
		return
	}
	m.mem = make([]uint8, size)
}

func (m *Memory) LoadROM(rom []byte) {
	if len(rom) > len(m.mem) {
		fmt.Println("ROM size exceeds memory size")
		return
	}
	copy(m.mem[ROM_BANK_0_START:ROM_BANK_0_END+1], rom)
}

func (m *Memory) readAddr(Addr uint16) uint8 {
	if int(Addr) >= len(m.mem) {

		fmt.Println("Memory read out of bounds at address:", Addr, len(m.mem))
		return 0
	}

	return m.mem[Addr]

}

func (m *Memory) writeAddr(Addr uint16, val uint8) {
	if int(Addr) >= len(m.mem) {
		fmt.Println("Memory write out of bounds at address:", Addr)
		return
	}

	switch Addr {
	case DIV:
		m.mem[DIV] = 0
		m.divCounter = 0
		return
	case TAC:
		val &= 0x07
		m.mem[Addr] = val

		switch val & 0x03 {
		case 0x00:
			m.timerFreq = 256
		case 0x01:
			m.timerFreq = 4
		case 0x02:
			m.timerFreq = 16
		case 0x03:
			m.timerFreq = 64
		}

		m.timerCounter = 0
		return
	case LY:
		m.mem[Addr] = 0
		return
	case 0xFF46: // OAM DMA
		m.DMA(val)
	default:
		m.mem[Addr] = val
	}

}

func (m *Memory) DMA(val uint8) {
	addr := uint16(val) << 8
	for i := 0; i < 0xA0; i++ {
		if int(addr+uint16(i)) >= len(m.mem) {
			fmt.Println("DMA write out of bounds at address:", addr+uint16(i))
			return
		}
		m.mem[0xFE00+i] = m.mem[addr+uint16(i)]
	}
}

func (m *Memory) EISet(mask INTERRUPT_ENABLE_MASK) {
	m.writeAddr(INTERRUPT_ENABLE, m.readAddr(INTERRUPT_ENABLE)|uint8(mask))
}

func (m *Memory) EIGet(mask INTERRUPT_ENABLE_MASK) bool {
	return (m.readAddr(INTERRUPT_ENABLE) & uint8(mask)) != 0
}

func (m *Memory) EFSet(mask INTERRUPT_FLAG_MASK) {
	Ly := m.readAddr(LY)
	fmt.Println("Setting interrupt flag:", mask, "at LY:", Ly)
	m.writeAddr(INTERRUPT_FLAG, m.readAddr(INTERRUPT_FLAG)|uint8(mask))
}

func (m *Memory) EFGet(mask INTERRUPT_FLAG_MASK) bool {
	return (m.readAddr(INTERRUPT_FLAG) & uint8(mask)) != 0
}

func (g *Gameboy) UpdateClock(Mcycles int) {

	//TODO FIX TIMER (Mcycles is not always 4)
	g.memory.divCounter += uint16(Mcycles)
	if g.memory.divCounter >= 64 {
		g.memory.divCounter -= 64
		g.memory.mem[DIV] = g.memory.mem[DIV] + 1
	}

	if g.memory.timeroverflow {
		g.memory.timeroverflow = false
		g.memory.EFSet(INTERRUPT_FLAG_TIMER)
		g.memory.mem[TIMA] = g.memory.mem[TMA]
	}

	if g.ReadAddr(TAC)&0x04 != 0 {
		g.memory.timerCounter += uint16(Mcycles)
		if g.memory.timerCounter >= g.memory.timerFreq {
			g.memory.timerCounter -= g.memory.timerFreq
			tima := g.ReadAddr(TIMA)
			if tima == 0xFF {

				g.memory.timeroverflow = true
			} else {
				g.WriteAddr(TIMA, tima+1)
			}
		}
	}
}
