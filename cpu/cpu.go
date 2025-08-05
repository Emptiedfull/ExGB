package cpu

import (
	"encoding/binary"
	"fmt"
)

type Cpu struct {
	registers registers
	pc        uint16

	IME    bool
	IMESet bool

	CurrentStep func(g *Gameboy)
}

type registers struct {
	a  uint8
	b  uint8
	c  uint8
	d  uint8
	e  uint8
	f  flagRegister
	h  uint8
	l  uint8
	sp uint16
}

type flagRegister struct {
	z bool
	n bool
	h bool
	c bool
}

type r16 int
type r8 int

const (
	BC r16 = iota
	DE
	HL
	AF
	SP
)

const (
	A r8 = iota
	B
	C
	D
	E
	F
	H
	L

	HLP
)

func (r *registers) set16(reg r16, val uint16) {
	switch reg {
	case BC:
		r.setBc(val)
	case DE:
		r.setDe(val)
	case HL:
		r.setHl(val)
	case AF:
		r.setAf(val)
	case SP:
		r.sp = val
	}
}

func (r *registers) get16(reg r16) uint16 {
	switch reg {
	case BC:
		return r.getBc()
	case DE:
		return r.getDe()
	case HL:
		return r.getHl()
	case AF:
		return r.getAf()
	case SP:
		return r.sp
	default:
		return 0
	}
}

func (g *Gameboy) get8(reg r8) uint8 {
	r := g.cpu.registers
	switch reg {
	case A:
		return r.a
	case B:
		return r.b
	case C:
		return r.c
	case D:
		return r.d
	case E:
		return r.e
	case F:
		return r.f.convertInt8()
	case H:
		return r.h
	case L:
		return r.l
	case HLP:
		return g.memory.readAddr(r.getHl())
	default:
		return 0
	}
}

func (g *Gameboy) set8(reg r8, val uint8) {

	switch reg {
	case A:
		g.cpu.registers.a = val
	case B:
		g.cpu.registers.b = val
	case C:
		g.cpu.registers.c = val
	case D:
		g.cpu.registers.d = val
	case E:
		g.cpu.registers.e = val
	case F:
		g.cpu.registers.f.setFromUint8(val)
	case H:
		g.cpu.registers.h = val
	case L:
		g.cpu.registers.l = val
	case HLP:
		addr := g.cpu.registers.getHl()
		g.memory.writeAddr(addr, val)
	default:
		// Do nothing for invalid register
	}
}

func setBit(val uint8, bit uint8) uint8 {
	return val | (1 << bit)
}

func (f *flagRegister) convertInt8() uint8 {
	var flags uint8
	if f.z {
		flags = setBit(flags, 7)
	}
	if f.n {
		flags = setBit(flags, 6)
	}
	if f.h {
		flags = setBit(flags, 5)
	}
	if f.c {
		flags = setBit(flags, 4)
	}

	return flags
}

func (f *flagRegister) setFromUint8(flags uint8) {
	f.z = (flags & (1 << 7)) != 0
	f.n = (flags & (1 << 6)) != 0
	f.h = (flags & (1 << 5)) != 0
	f.c = (flags & (1 << 4)) != 0
}

func (r *registers) setBc(val uint16) {
	b := val >> 8
	c := val & 0xFF
	r.b = uint8(b)
	r.c = uint8(c)
}

func (r *registers) setDe(val uint16) {
	d := val >> 8
	e := val & 0xFF
	r.d = uint8(d)
	r.e = uint8(e)
}

func (r *registers) setAf(val uint16) {
	a := val >> 8
	f := val & 0xFF
	r.a = uint8(a)
	r.f.setFromUint8(uint8(f))
}

func (r *registers) setHl(val uint16) {
	h := val >> 8
	l := val & 0xFF
	r.h = uint8(h)
	r.l = uint8(l)
}

func (r *registers) getBc() uint16 {
	return uint16(r.b)<<8 | uint16(r.c)&0xFF
}

func (r *registers) getDe() uint16 {
	return uint16(r.d)<<8 | uint16(r.e)&0xFF
}

func (r *registers) getAf() uint16 {
	return uint16(r.a)<<8 | uint16(r.f.convertInt8())&0xFF
}

func (r *registers) getHl() uint16 {
	return uint16(r.h)<<8 | uint16(r.l)&0xFF
}


func (g *Gameboy) SaveState() []byte {

	buf := make([]byte, 0)
	regs := g.cpu.registers
	buf = append(buf, regs.a, regs.b, regs.c, regs.d, regs.e, regs.h, regs.l)
	buf = append(buf, regs.f.convertInt8())
	tmp := make([]byte, 2)
	binary.LittleEndian.PutUint16(tmp, regs.sp)
	buf = append(buf, tmp...)
	binary.LittleEndian.PutUint16(tmp, g.cpu.pc)
	buf = append(buf, tmp...)

	if g.cpu.IME {
		buf = append(buf, 1)
	} else {
		buf = append(buf, 0)
	}

	if g.cpu.IMESet {
		buf = append(buf, 1)
	} else {
		buf = append(buf, 0)
	}

	if g.halted {
		buf = append(buf, 1)
	} else {
		buf = append(buf, 0)
	}

	buf = append(buf, g.memory.mem...)
	buf = append(buf, g.memory.cartridge_mem...)

	return buf

}

func (g *Gameboy) LoadState(state []byte) error {
	if len(state) < 65536+12 {
		fmt.Println("State data too short:", len(state), "bytes")
		return fmt.Errorf("state data too short: %d bytes", len(state))
	}

	idx := 0
	g.cpu.registers.a = state[idx]
	idx++
	g.cpu.registers.b = state[idx]
	idx++
	g.cpu.registers.c = state[idx]
	idx++
	g.cpu.registers.d = state[idx]
	idx++
	g.cpu.registers.e = state[idx]
	idx++
	g.cpu.registers.h = state[idx]
	idx++
	g.cpu.registers.l = state[idx]
	idx++
	g.cpu.registers.f.setFromUint8(state[idx])
	idx++

	g.cpu.registers.sp = binary.LittleEndian.Uint16(state[idx : idx+2])
	idx += 2
	g.cpu.pc = binary.LittleEndian.Uint16(state[idx : idx+2])
	idx += 2

	g.cpu.IME = state[idx] != 0
	idx++
	g.cpu.IMESet = state[idx] != 0
	idx++
	g.halted = state[idx] != 0
	idx++

	copy(g.memory.mem, state[idx:idx+65536])
	idx += 65536
	g.memory.cartridge_mem = make([]uint8, len(state[idx:]))
	copy(g.memory.cartridge_mem, state[idx:])
	fmt.Println("cartridgemem", len(g.memory.cartridge_mem), len(state[idx:]))
	g.memory.ApplyBanking()

	return nil

}

func (g *Gameboy) DebugStep() int {

	if g.cpu.IMESet {
		g.cpu.IME = true
		g.cpu.IMESet = false
	}

	if g.cpu.IME {
		IE := g.memory.readAddr(INTERRUPT_ENABLE)
		IF := g.memory.readAddr(INTERRUPT_FLAG)
		if IE&IF != 0 {
			g.halted = false
			g.handleInterrupt()
			// g.UpdateClock(5)
			return 5
		}
	}

	if g.halted {
		IE := g.memory.readAddr(INTERRUPT_ENABLE)
		IF := g.memory.readAddr(INTERRUPT_FLAG)
		if IE&IF != 0 {
			g.halted = false
			if g.cpu.IME {
				g.handleInterrupt()

			} else {
				g.haltbug = true
			}
		}
		// g.UpdateClock(1)
		return 1
	}

	code := g.fetch()
	if g.haltbug {
		g.cpu.pc--
		g.haltbug = false
		return 1
	}
	opcodeMux.Lock()
	opCode := opcodes[code]
	opcodeMux.Unlock()

	if opCode.Execute == nil {
		fmt.Println(opCode, code)
		return 1
	}

	opCode.Execute(g)
	// g.UpdateClock(opCode.MCycles)
	return opCode.MCycles

}

func (g *Gameboy) handleInterrupt() {
	IE := g.memory.readAddr(INTERRUPT_ENABLE)
	IF := g.memory.readAddr(INTERRUPT_FLAG)

	interrupts := IE & IF

	if interrupts&0x01 != 0 { // V-Blank
		g.memory.writeAddr(INTERRUPT_FLAG, IF&^0x01)
		g.serviceInterrupt(0x40)
	} else if interrupts&0x02 != 0 { // LCD STAT
		g.memory.writeAddr(INTERRUPT_FLAG, IF&^0x02)
		g.serviceInterrupt(0x48)
	} else if interrupts&0x04 != 0 { // Timer
		g.memory.writeAddr(INTERRUPT_FLAG, IF&^0x04)
		g.serviceInterrupt(0x50)
	} else if interrupts&0x08 != 0 { // Serial
		g.memory.writeAddr(INTERRUPT_FLAG, IF&^0x08)
		g.serviceInterrupt(0x58)
	} else if interrupts&0x10 != 0 { // Joypad
		g.memory.writeAddr(INTERRUPT_FLAG, IF&^0x10)
		g.serviceInterrupt(0x60)
	}
}

func (g *Gameboy) serviceInterrupt(addr uint16) {
	g.cpu.IME = false
	g.pushStack16(g.cpu.pc)
	g.cpu.pc = addr
}
