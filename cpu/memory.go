package cpu

import (
	"fmt"
	"time"
)

type Memory struct {
	cartridge_mem []uint8

	mem []uint8

	divCounter    uint16
	timerCounter  uint16
	timerFreq     uint16
	timeroverflow bool

	joypadState uint8

	mcb            mcbtype
	currentRomBank uint8
	romBankHigh    uint8
	romBankLow     uint8

	enableRam      bool
	ramBanks       []uint8
	currentRamBank int

	RomBankingMode bool

	prevdebugbank uint8

	RtcRegister  rtcRegister
	RtcSelected  bool
	CurrentRtc   uint8
	lastRtcWrite uint64
}

type rtcRegister struct {
	seconds uint8
	minutes uint8
	hours   uint8
	dayLow  uint8
	dayHigh uint8
}

type mcbtype int

const (
	mcbNone mcbtype = iota
	mcbMBC1
	mcbMBC2
	mcbMBC3
)

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
	m.cartridge_mem = make([]uint8, len(rom))
	copy(m.cartridge_mem, rom)
	m.ApplyBanking()
}

func (m *Memory) ApplyBanking() {
	rom := m.cartridge_mem
	mcbint := rom[0x147]
	fmt.Println(mcbint)
	switch mcbint {
	case 0x00:
		m.mcb = mcbNone
	case 0x01, 0x02, 0x03:
		m.mcb = mcbMBC1
	case 0x05, 0x06:
		m.mcb = mcbMBC2
	case 0x0F, 0x10, 0x11, 0x12, 0x13:
		m.mcb = mcbMBC3
		m.RtcRegister = rtcRegister{
			seconds: 0,
			minutes: 0,
			hours:   0,
			dayLow:  0,
			dayHigh: 0}
	default:
		fmt.Println("Unknown MBC type:", mcbint)
		return
	}
	m.currentRomBank = 1

	var ramSize int
	switch rom[0x149] {
	case 0x00:
		ramSize = 0
	case 0x01:
		ramSize = 2 * 1024
	case 0x02:
		ramSize = 8 * 1024
	case 0x03:
		ramSize = 32 * 1024
	case 0x04:
		ramSize = 128 * 1024
	case 0x05:
		ramSize = 64 * 1024
	}

	m.ramBanks = make([]uint8, ramSize)
	m.currentRamBank = 0
	m.romBankLow = 1
	m.romBankHigh = 0

	m.RomBankingMode = true
	m.enableRam = false
	m.updateCurrentRomBank()

	fmt.Println(len(m.cartridge_mem))

	m.CheckRomBankDuplicates()

	copy(m.mem[ROM_BANK_0_START:ROM_BANK_0_END+1], m.cartridge_mem[ROM_BANK_0_START:ROM_BANK_0_END+1])

}

func (m *Memory) CheckRomBankDuplicates() {
	const bankSize = 16 * 1024 // 16KB per bank
	numBanks := len(m.cartridge_mem) / bankSize
	bank0 := m.cartridge_mem[:bankSize]

	for i := 1; i < numBanks; i++ {
		banki := m.cartridge_mem[i*bankSize : (i+1)*bankSize]
		same := true
		for j := 0; j < bankSize; j++ {
			if banki[j] != bank0[j] {
				same = false
				break
			}
		}
		if same {
			fmt.Printf("Bank %d is identical to bank 0\n", i)
		}
	}
}

func (m *Memory) readAddr(Addr uint16) uint8 {
	if int(Addr) >= len(m.mem) {
		fmt.Println("Memory read out of bounds at address:", Addr, len(m.mem))
		return 0
	}

	switch {
	case Addr <= 0x7FFF && Addr >= 0x4000:
		if m.currentRomBank != m.prevdebugbank {

			m.prevdebugbank = m.currentRomBank
		}

		var romBank uint8

		if int(m.currentRomBank) >= len(m.cartridge_mem)/0x4000-1 {
			romBank = m.currentRomBank % uint8(len(m.cartridge_mem)/0x4000)

		} else {
			romBank = m.currentRomBank
		}
		if m.currentRomBank == 0 {
			romBank = 1
		}

		newaddr := uint64(Addr) - 0x4000 + uint64(romBank)*0x4000
		if int(newaddr) >= len(m.cartridge_mem) {
			fmt.Println("ROM read out of bounds at address:", newaddr)
			return 0
		}
		return m.cartridge_mem[newaddr]
	case Addr >= 0xA000 && Addr <= 0xBFFF:
		if !m.enableRam || len(m.ramBanks) == 0 {
			return 0xFF
		}
		var ramBank int
		if m.mcb == mcbMBC1 && !m.RomBankingMode {
			ramBank = m.currentRamBank
		} else {
			ramBank = 0
		}

		newAddr := Addr - 0xA000 + uint16(ramBank*0x2000)
		if int(newAddr) >= len(m.ramBanks) {
			fmt.Printf("RAM read out of bounds: Addr=0x%04X, newAddr=0x%04X\n", Addr, newAddr)
			return 0
		}
		return m.ramBanks[newAddr]
	case Addr == 0xFF00:
		return m.getJoyPadState()
	default:
		return m.mem[Addr]
	}

}

func (m *Memory) writeAddr(Addr uint16, val uint8) {
	if int(Addr) >= len(m.mem) {
		fmt.Println("Memory write out of bounds at address:", Addr)
		return
	}

	if Addr <= 0x7FFF {
		if m.mcb != mcbNone {
			m.HandleBanking(Addr, val)
		}
		return
	}

	if Addr >= 0xA000 && Addr <= 0xBFFF {
		if !m.enableRam || len(m.ramBanks) == 0 {
			return
		}
		var ramBank int
		if m.mcb == mcbMBC1 && !m.RomBankingMode {
			ramBank = m.currentRamBank
		} else {
			ramBank = 0
		}
		newAddr := Addr - 0xA000 + uint16(ramBank*0x2000)
		if int(newAddr) >= len(m.ramBanks) {
			fmt.Printf("RAM write out of bounds: Addr=0x%04X, newAddr=0x%04X\n", Addr, newAddr)
			return
		}
		m.ramBanks[newAddr] = val
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
	case 0xFF00: // Joypad

		m.mem[0xFF00] = val

		selectedGroup := val & 0x30
		var groupState uint8
		if selectedGroup&0x10 == 0 {
			groupState = (m.joypadState >> 4) & 0x0F
		} else if selectedGroup&0x20 == 0 {
			groupState = m.joypadState & 0x0F
		}
		if groupState != 0x0F {
			m.EFSet(INTERRUPT_FLAG_JOYPAD)
		}
		return
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

func (m *Memory) HandleBanking(Addr uint16, val uint8) {
	if m.mcb == mcbMBC1 {
		switch {
		case Addr < 0x2000:
			if m.mcb != mcbNone {
				data := val & 0x0F
				if data == 0xA {
					m.enableRam = true
				} else {
					m.enableRam = false
				}
			}
		case Addr >= 0x2000 && Addr < 0x4000:
			if m.mcb != mcbNone {
				// fmt.Println("Switching ROM bank LO")
				m.LoRomBankChange(val)
			}

		case Addr >= 0x4000 && Addr < 0x6000:
			if m.mcb == mcbMBC1 {
				if m.RomBankingMode {
					m.HiRomBankChange(val)
				} else {
					fmt.Println("Switching RAM bank", val&0x03)
					m.currentRamBank = int(val & 0x03)
				}
			}

		case Addr >= 0x6000 && Addr < 0x8000:
			if m.mcb == mcbMBC1 {
				m.RomBankingMode = (val & 0x01) == 0
				m.updateCurrentRomBank()
			}
		}
	}

	if m.mcb == mcbMBC3 {
		if Addr < 0x2000 {
			if val&0x0F == 0x0A {
				m.enableRam = true
			} else {
				m.enableRam = false
			}
		}
		if Addr >= 0x2000 && Addr < 0x4000 {
			m.currentRomBank = val & 0x7F
			if m.currentRomBank == 0 {
				m.currentRomBank = 1
			}
		}

		if Addr >= 0x4000 && Addr < 0x6000 {
			if val <= 0x07 {
				m.currentRamBank = int(val)
				m.RtcSelected = false
			} else if val >= 0x08 && val <= 0x0C {
				m.RtcSelected = true

				switch val {
				case 0x08:
					m.CurrentRtc = m.RtcRegister.seconds
				case 0x09:
					m.CurrentRtc = m.RtcRegister.minutes
				case 0x0A:
					m.CurrentRtc = m.RtcRegister.hours
				case 0x0B:
					m.CurrentRtc = m.RtcRegister.dayLow
				case 0x0C:
					m.CurrentRtc = m.RtcRegister.dayHigh
				}
			}
		}

		if Addr >= 0x6000 && Addr < 0x8000 {
			if m.lastRtcWrite == 0x00 && val == 0x01 {
				m.latchRtc()
			}
			m.lastRtcWrite = uint64(val)
		}

	}

}

func (m *Memory) latchRtc() {
	now := time.Now()
	m.RtcRegister.seconds = uint8(now.Second())
	m.RtcRegister.minutes = uint8(now.Minute())
	m.RtcRegister.hours = uint8(now.Hour())
	day := now.YearDay() // Use day of year for example
	m.RtcRegister.dayLow = uint8(day & 0xFF)
	m.RtcRegister.dayHigh = uint8((day >> 8) & 0x01)
}

func (m *Memory) HiRomBankChange(val uint8) {
	if m.mcb == mcbMBC1 {
		m.romBankHigh = val & 0x03
		m.updateCurrentRomBank()
	}

}

func (m *Memory) LoRomBankChange(val uint8) {

	if m.mcb == mcbMBC1 {
		m.romBankLow = val & 0x1F
		if m.romBankLow == 0 {
			m.romBankLow = 1
		}
		m.updateCurrentRomBank()
	}
}

func (m *Memory) updateCurrentRomBank() {
	if m.romBankLow == 0 {
		m.romBankLow = 1
	}
	if m.RomBankingMode {
		m.currentRomBank = (m.romBankHigh << 5) | m.romBankLow
	} else {
		m.currentRomBank = m.romBankLow
	}
	m.currentRomBank &= 0x7F
	if m.currentRomBank == 0 {
		m.currentRomBank = 1
	}

}
