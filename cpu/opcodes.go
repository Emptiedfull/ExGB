package cpu

import "fmt"

type OpCode struct {
	MCycles int
	Execute func(g *Gameboy)
}

type PrefixOpCode func(g *Gameboy) int

var opcodes map[uint8]OpCode

func (g *Gameboy) initOpcodes() map[uint8]OpCode {

	return map[uint8]OpCode{
		0x00: {MCycles: 1, Execute: func(g *Gameboy) {
			// fmt.Println("NOP executed")
		}}, // NOP
		0x01: {MCycles: 3, Execute: func(g *Gameboy) {
			g.Load16r16n(BC, g.fetch16())
		}}, 0x02: {MCycles: 2, Execute: func(g *Gameboy) {
			g.LoadMEMA(BC)
		}},
		0x03: {MCycles: 2, Execute: func(g *Gameboy) {
			g.inc16(BC)
		}}, 0x04: {MCycles: 1, Execute: func(g *Gameboy) {
			g.inc8(B)
		}}, 0x05: {MCycles: 1, Execute: func(g *Gameboy) {
			g.dec8(B)
		}}, 0x06: {MCycles: 2, Execute: func(g *Gameboy) {
			g.Loadr8u8(B, g.fetch())
		}}, 0x07: {MCycles: 2, Execute: func(g *Gameboy) {
			g.RLCA()
		}}, 0x08: {MCycles: 5, Execute: func(g *Gameboy) {
			Addr := g.fetch16()
			n1 := uint8(g.cpu.registers.sp & 0xFF)
			n2 := uint8(g.cpu.registers.sp >> 8)
			g.memory.writeAddr(Addr, n1)
			g.memory.writeAddr(Addr+1, n2)
		}}, 0x09: {MCycles: 2, Execute: func(g *Gameboy) {
			g.AddHLr16(BC)
		}}, 0x0A: {MCycles: 2, Execute: func(g *Gameboy) {
			g.LoadAn16(BC)
		}}, 0x0B: {MCycles: 2, Execute: func(g *Gameboy) {
			g.dec16(BC)
		}}, 0x0C: {MCycles: 1, Execute: func(g *Gameboy) {
			g.inc8(C)
		}}, 0x0D: {MCycles: 1, Execute: func(g *Gameboy) {
			g.dec8(C)
		}}, 0x0E: {MCycles: 2, Execute: func(g *Gameboy) {
			g.Loadr8u8(C, g.fetch())
		}}, 0x0F: {MCycles: 1, Execute: func(g *Gameboy) {
			g.RRCA()
		}}, 0x10: {MCycles: 1, Execute: func(g *Gameboy) {
			g.fetch()
			//STOP IDK WHAT TO DO HEREs
		}}, 0x11: {MCycles: 3, Execute: func(g *Gameboy) {
			g.Load16r16n(DE, g.fetch16())
		}}, 0x12: {MCycles: 2, Execute: func(g *Gameboy) {
			// fmt.Println("Loading DE to A")
			g.LoadMEMA(DE)
		}}, 0x13: {MCycles: 2, Execute: func(g *Gameboy) {
			g.inc16(DE)
		}}, 0x14: {MCycles: 1, Execute: func(g *Gameboy) {

			g.inc8(D)
		}}, 0x15: {MCycles: 1, Execute: func(g *Gameboy) {
			g.dec8(D)
		}}, 0x16: {MCycles: 2, Execute: func(g *Gameboy) {
			g.Loadr8u8(D, g.fetch())
		}}, 0x17: {MCycles: 1, Execute: func(g *Gameboy) {
			g.RLA()
		}}, 0x18: {MCycles: 3, Execute: func(g *Gameboy) {
			offset := int8(g.fetch())
			g.cpu.pc = uint16(int32(g.cpu.pc) + int32(offset))
		}}, 0x19: {MCycles: 2, Execute: func(g *Gameboy) {
			g.AddHLr16(DE)
		}}, 0x1A: {MCycles: 2, Execute: func(g *Gameboy) {
			g.Loadr16MEMA(DE)
		}}, 0x1B: {MCycles: 2, Execute: func(g *Gameboy) {
			g.dec16(DE)
		}}, 0x1C: {MCycles: 1, Execute: func(g *Gameboy) {

			g.inc8(E)

		}}, 0x1D: {MCycles: 1, Execute: func(g *Gameboy) {
			g.dec8(E)
		}}, 0x1E: {MCycles: 2, Execute: func(g *Gameboy) {
			g.Loadr8u8(E, g.fetch())
		}}, 0x1F: {MCycles: 1, Execute: func(g *Gameboy) {
			g.RRA()
		}}, 0x20: {MCycles: 3, Execute: func(g *Gameboy) {
			//JR NZ, i8 - Jump relative if zero flag is not set
			offset := int8(g.fetch())
			// opcode := opcodes[0x20]
			if !g.cpu.registers.f.z {
				g.cpu.pc = uint16(int32(g.cpu.pc) + int32(offset))
				// fmt.Println("jumping 0x20 to", g.cpu.pc, "offset", offset, "E", g.cpu.registers.e)
				// opcode.MCycles = 3
			} else {

				// opcode.MCycles = 2
			}
			// opcodes[0x20] = opcode
		}}, 0x21: {MCycles: 3, Execute: func(g *Gameboy) {
			g.Load16r16n(HL, g.fetch16())
		}}, 0x22: {MCycles: 2, Execute: func(g *Gameboy) {
			g.LoadMEMA(HL)
			g.inc16(HL)
		}}, 0x23: {MCycles: 2, Execute: func(g *Gameboy) {
			g.inc16(HL)
		}}, 0x24: {MCycles: 1, Execute: func(g *Gameboy) {
			g.inc8(H)
		}}, 0x25: {MCycles: 1, Execute: func(g *Gameboy) {
			g.dec8(H)
		}}, 0x26: {MCycles: 2, Execute: func(g *Gameboy) {
			g.Loadr8u8(H, g.fetch())
		}}, 0x27: {MCycles: 1, Execute: func(g *Gameboy) {

			// DAA

			a := g.cpu.registers.a

			if !g.cpu.registers.f.n {
				if g.cpu.registers.f.c || a > 0x99 {
					a += 0x60
					g.cpu.registers.f.c = true
				}
				if g.cpu.registers.f.h || (a&0x0F) > 0x09 {
					a += 0x06
				}
			} else {
				if g.cpu.registers.f.c {
					a -= 0x60
				}
				if g.cpu.registers.f.h {
					a -= 0x06
				}
			}

			g.cpu.registers.f.z = a == 0
			g.cpu.registers.f.h = false
			g.cpu.registers.a = a
		}}, 0x28: {MCycles: 3, Execute: func(g *Gameboy) {
			// JR Z, i8 - Jump relative if zero flag is set
			// fmt.Println("trying jump")
			offset := int8(g.fetch())
			// opcode := opcodes[0x28]
			if g.cpu.registers.f.z {
				g.cpu.pc = uint16(int32(g.cpu.pc) + int32(offset))
				// opcode.MCycles = 3
			} else {
				// opcode.MCycles = 2
			}
			// opcodes[0x28] = opcode
		}}, 0x29: {MCycles: 2, Execute: func(g *Gameboy) {
			g.AddHLr16(HL)
		}}, 0x2A: {MCycles: 2, Execute: func(g *Gameboy) {
			g.Loadr16MEMA(HL)
			g.inc16(HL)
		}}, 0x2B: {MCycles: 2, Execute: func(g *Gameboy) {
			g.dec16(HL)
		}}, 0x2C: {MCycles: 1, Execute: func(g *Gameboy) {
			g.inc8(L)
		}}, 0x2D: {MCycles: 1, Execute: func(g *Gameboy) {
			g.dec8(L)
		}}, 0x2E: {MCycles: 2, Execute: func(g *Gameboy) {
			g.Loadr8u8(L, g.fetch())
		}}, 0x2F: {MCycles: 1, Execute: func(g *Gameboy) {
			g.cpu.registers.a = ^g.cpu.registers.a
			g.cpu.registers.f.n = true
			g.cpu.registers.f.h = true
		}}, 0x30: {MCycles: 3, Execute: func(g *Gameboy) {
			offset := int8(g.fetch())
			// opcode := opcodes[0x30]
			if !g.cpu.registers.f.c {
				g.cpu.pc = uint16(int32(g.cpu.pc) + int32(offset))
				// opcode.MCycles = 3
			} else {
				// opcode.MCycles = 2
			}
			// opcodes[0x20] = opcode
		}}, 0x31: {MCycles: 3, Execute: func(g *Gameboy) {
			val := g.fetch16()
			g.cpu.registers.sp = val
		}}, 0x32: {MCycles: 2, Execute: func(g *Gameboy) {
			g.LoadMEMA(HL)
			g.dec16(HL)
		}}, 0x33: {MCycles: 2, Execute: func(g *Gameboy) {
			g.cpu.registers.sp += 1
		}}, 0x34: {MCycles: 3, Execute: func(g *Gameboy) {
			addr := g.cpu.registers.get16(HL)
			val := g.memory.readAddr(addr)
			res := val + 1
			g.cpu.registers.f.z = res == 0
			g.cpu.registers.f.n = false
			g.cpu.registers.f.h = (val & 0x0F) == 0x0F
			g.memory.writeAddr(addr, res)
		}}, 0x35: {MCycles: 3, Execute: func(g *Gameboy) {
			addr := g.cpu.registers.get16(HL)
			val := g.memory.readAddr(addr)
			res := val - 1
			g.cpu.registers.f.z = res == 0
			g.cpu.registers.f.n = true
			g.cpu.registers.f.h = (val & 0x0F) == 0x00
			g.memory.writeAddr(addr, res)
		}}, 0x36: {MCycles: 3, Execute: func(g *Gameboy) {
			addr := g.cpu.registers.get16(HL)
			g.memory.writeAddr(addr, g.fetch())
		}}, 0x37: {MCycles: 1, Execute: func(g *Gameboy) {
			g.cpu.registers.f.n = false
			g.cpu.registers.f.h = false
			g.cpu.registers.f.c = true
		}}, 0x38: {MCycles: 3, Execute: func(g *Gameboy) {
			//JR C, i8 - Jump relative if carry flag is  set
			offset := int8(g.fetch())
			opcode := opcodes[0x20]
			if g.cpu.registers.f.c {
				g.cpu.pc = uint16(int32(g.cpu.pc) + int32(offset))
				opcode.MCycles = 3
			} else {
				opcode.MCycles = 2
			}
			opcodes[0x20] = opcode
		}}, 0x39: {MCycles: 2, Execute: func(g *Gameboy) {
			g.AddHLr16(SP)
		}}, 0x3A: {MCycles: 2, Execute: func(g *Gameboy) {
			g.Loadr16MEMA(HL)
			g.dec16(HL)
		}}, 0x3B: {MCycles: 2, Execute: func(g *Gameboy) {
			g.dec16(SP)
		}}, 0x3C: {MCycles: 1, Execute: func(g *Gameboy) {
			g.inc8(A)
		}}, 0x3D: {MCycles: 1, Execute: func(g *Gameboy) {
			g.dec8(A)
		}}, 0x3E: {MCycles: 2, Execute: func(g *Gameboy) {
			g.Loadr8u8(A, g.fetch())
		}}, 0x3F: {MCycles: 1, Execute: func(g *Gameboy) {
			g.cpu.registers.f.c = !g.cpu.registers.f.c
			g.cpu.registers.f.n = false
			g.cpu.registers.f.h = false
		}}, 0x40: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Loadr8r8(B, B)
		}}, 0x41: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Loadr8r8(B, C)
		}}, 0x42: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Loadr8r8(B, D)
		}}, 0x43: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Loadr8r8(B, E)
		}}, 0x44: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Loadr8r8(B, H)
		}}, 0x45: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Loadr8r8(B, L)
		}}, 0x46: {MCycles: 2, Execute: func(g *Gameboy) {
			g.Loadr8HL(B)
		}}, 0x47: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Loadr8r8(B, A)
		}}, 0x48: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Loadr8r8(C, B)
		}}, 0x49: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Loadr8r8(C, C)
		}}, 0x4A: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Loadr8r8(C, D)
		}}, 0x4B: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Loadr8r8(C, E)
		}}, 0x4C: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Loadr8r8(C, H)
		}}, 0x4D: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Loadr8r8(C, L)
		}}, 0x4E: {MCycles: 2, Execute: func(g *Gameboy) {
			g.Loadr8HL(C)
		}}, 0x4F: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Loadr8r8(C, A)
		}}, 0x50: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Loadr8r8(D, B)
		}}, 0x51: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Loadr8r8(D, C)
		}}, 0x52: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Loadr8r8(D, D)
		}}, 0x53: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Loadr8r8(D, E)
		}}, 0x54: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Loadr8r8(D, H)
		}}, 0x55: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Loadr8r8(D, L)
		}}, 0x56: {MCycles: 2, Execute: func(g *Gameboy) {
			g.Loadr8HL(D)
		}}, 0x57: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Loadr8r8(D, A)
		}}, 0x58: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Loadr8r8(E, B)
		}}, 0x59: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Loadr8r8(E, C)
		}}, 0x5A: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Loadr8r8(E, D)
		}}, 0x5B: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Loadr8r8(E, E)
		}}, 0x5C: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Loadr8r8(E, H)
		}}, 0x5D: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Loadr8r8(E, L)
		}}, 0x5E: {MCycles: 2, Execute: func(g *Gameboy) {
			g.Loadr8HL(E)
		}}, 0x5F: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Loadr8r8(E, A)
		}}, 0x60: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Loadr8r8(H, B)
		}}, 0x61: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Loadr8r8(H, C)
		}}, 0x62: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Loadr8r8(H, D)
		}}, 0x63: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Loadr8r8(H, E)
		}}, 0x64: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Loadr8r8(H, H)
		}}, 0x65: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Loadr8r8(H, L)
		}}, 0x66: {MCycles: 2, Execute: func(g *Gameboy) {
			g.Loadr8HL(H)
		}}, 0x67: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Loadr8r8(H, A)
		}}, 0x68: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Loadr8r8(L, B)
		}}, 0x69: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Loadr8r8(L, C)
		}}, 0x6A: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Loadr8r8(L, D)
		}}, 0x6B: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Loadr8r8(L, E)
		}}, 0x6C: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Loadr8r8(L, H)
		}}, 0x6D: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Loadr8r8(L, L)
		}}, 0x6E: {MCycles: 2, Execute: func(g *Gameboy) {
			g.Loadr8HL(L)
		}}, 0x6F: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Loadr8r8(L, A)
		}}, 0x70: {MCycles: 2, Execute: func(g *Gameboy) {
			g.LoadHLr8(B)
		}}, 0x71: {MCycles: 2, Execute: func(g *Gameboy) {
			g.LoadHLr8(C)
		}}, 0x72: {MCycles: 2, Execute: func(g *Gameboy) {
			g.LoadHLr8(D)
		}}, 0x73: {MCycles: 2, Execute: func(g *Gameboy) {
			g.LoadHLr8(E)
		}}, 0x74: {MCycles: 2, Execute: func(g *Gameboy) {
			g.LoadHLr8(H)
		}}, 0x75: {MCycles: 2, Execute: func(g *Gameboy) {
			g.LoadHLr8(L)
		}}, 0x76: {MCycles: 0, Execute: func(g *Gameboy) {
			g.halted = true
		}}, 0x77: {MCycles: 2, Execute: func(g *Gameboy) {
			g.LoadHLr8(A)
		}}, 0x78: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Loadr8r8(A, B)
		}}, 0x79: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Loadr8r8(A, C)
		}}, 0x7A: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Loadr8r8(A, D)
		}}, 0x7B: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Loadr8r8(A, E)
		}}, 0x7C: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Loadr8r8(A, H)
		}}, 0x7D: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Loadr8r8(A, L)
		}}, 0x7E: {MCycles: 2, Execute: func(g *Gameboy) {
			g.Loadr8HL(A)
		}}, 0x7F: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Loadr8r8(A, A)
		}}, 0x80: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Addr8r8(A, B)
		}}, 0x81: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Addr8r8(A, C)
		}}, 0x82: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Addr8r8(A, D)
		}}, 0x83: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Addr8r8(A, E)
		}}, 0x84: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Addr8r8(A, H)
		}}, 0x85: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Addr8r8(A, L)
		}}, 0x86: {MCycles: 2, Execute: func(g *Gameboy) {
			g.Addr8HL(A)
		}}, 0x87: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Addr8r8(A, A)
		}}, 0x88: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Adcr8r8(A, B)
		}}, 0x89: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Adcr8r8(A, C)
		}}, 0x8A: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Adcr8r8(A, D)
		}}, 0x8B: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Adcr8r8(A, E)
		}}, 0x8C: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Adcr8r8(A, H)
		}}, 0x8D: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Adcr8r8(A, L)
		}}, 0x8E: {MCycles: 2, Execute: func(g *Gameboy) {
			g.Adcr8HL(A)
		}}, 0x8F: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Adcr8r8(A, A)
		}}, 0x90: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Subr8r8(A, B)
		}}, 0x91: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Subr8r8(A, C)
		}}, 0x92: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Subr8r8(A, D)
		}}, 0x93: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Subr8r8(A, E)
		}}, 0x94: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Subr8r8(A, H)
		}}, 0x95: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Subr8r8(A, L)
		}}, 0x96: {MCycles: 2, Execute: func(g *Gameboy) {
			g.Subr8HL(A)
		}}, 0x97: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Subr8r8(A, A)
		}}, 0x98: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Sbcr8r8(A, B)
		}}, 0x99: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Sbcr8r8(A, C)
		}}, 0x9A: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Sbcr8r8(A, D)
		}}, 0x9B: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Sbcr8r8(A, E)
		}}, 0x9C: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Sbcr8r8(A, H)
		}}, 0x9D: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Sbcr8r8(A, L)
		}}, 0x9E: {MCycles: 2, Execute: func(g *Gameboy) {
			g.Sbcr8HL(A)
		}}, 0x9F: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Sbcr8r8(A, A)
		}}, 0xA0: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Andr8r8(A, B)
		}}, 0xA1: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Andr8r8(A, C)
		}}, 0xA2: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Andr8r8(A, D)
		}}, 0xA3: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Andr8r8(A, E)
		}}, 0xA4: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Andr8r8(A, H)
		}}, 0xA5: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Andr8r8(A, L)
		}}, 0xA6: {MCycles: 2, Execute: func(g *Gameboy) {
			g.Andr8HL(A)
		}}, 0xA7: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Andr8r8(A, A)
		}}, 0xA8: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Xorr8r8(A, B)
		}}, 0xA9: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Xorr8r8(A, C)
		}}, 0xAA: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Xorr8r8(A, D)
		}}, 0xAB: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Xorr8r8(A, E)
		}}, 0xAC: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Xorr8r8(A, H)
		}}, 0xAD: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Xorr8r8(A, L)
		}}, 0xAE: {MCycles: 2, Execute: func(g *Gameboy) {
			g.Xorr8HL(A)
		}}, 0xAF: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Xorr8r8(A, A)
		}}, 0xB0: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Or8r8(A, B)
		}}, 0xB1: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Or8r8(A, C)
		}}, 0xB2: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Or8r8(A, D)
		}}, 0xB3: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Or8r8(A, E)
		}}, 0xB4: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Or8r8(A, H)
		}}, 0xB5: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Or8r8(A, L)
		}}, 0xB6: {MCycles: 2, Execute: func(g *Gameboy) {
			g.Or8HL(A)
		}}, 0xB7: {MCycles: 1, Execute: func(g *Gameboy) {
			g.Or8r8(A, A)
		}}, 0xB8: {MCycles: 1, Execute: func(g *Gameboy) {
			g.CPr8r8(A, B)
		}}, 0xB9: {MCycles: 1, Execute: func(g *Gameboy) {
			g.CPr8r8(A, C)
		}}, 0xBA: {MCycles: 1, Execute: func(g *Gameboy) {
			g.CPr8r8(A, D)
		}}, 0xBB: {MCycles: 1, Execute: func(g *Gameboy) {
			g.CPr8r8(A, E)
		}}, 0xBC: {MCycles: 1, Execute: func(g *Gameboy) {
			g.CPr8r8(A, H)
		}}, 0xBD: {MCycles: 1, Execute: func(g *Gameboy) {
			g.CPr8r8(A, L)
		}}, 0xBE: {MCycles: 2, Execute: func(g *Gameboy) {
			g.CPr8HL(A)
		}}, 0xBF: {MCycles: 1, Execute: func(g *Gameboy) {
			g.CPr8r8(A, A)
		}}, 0xC0: {MCycles: 2, Execute: func(g *Gameboy) {
			opcode := opcodes[0xC0]
			if !g.cpu.registers.f.z {
				g.ret()
				opcode.MCycles = 5
			} else {
				opcode.MCycles = 2
			}
			opcodes[0xC0] = opcode
		}}, 0xC1: {MCycles: 3, Execute: func(g *Gameboy) {
			g.pop16(BC)
		}}, 0xC2: {MCycles: 3, Execute: func(g *Gameboy) {
			opcode := opcodes[0xC2]
			Addr := g.fetch16()
			if !g.cpu.registers.f.z {
				g.jumpu16(Addr)
				opcode.MCycles = 4
			} else {
				opcode.MCycles = 3
			}
			opcodes[0xC2] = opcode
		}}, 0xC3: {MCycles: 4, Execute: func(g *Gameboy) {
			g.jumpu16(g.fetch16())
		}}, 0xC4: {MCycles: 3, Execute: func(g *Gameboy) {
			OpCode := opcodes[0xC4]
			Addr := g.fetch16()
			if !g.cpu.registers.f.z {
				g.callu16(Addr)
				OpCode.MCycles = 6
			} else {
				OpCode.MCycles = 3
			}
		}}, 0xC5: {MCycles: 4, Execute: func(g *Gameboy) {
			g.push16(BC)
		}}, 0xC6: {MCycles: 2, Execute: func(g *Gameboy) {
			g.Addr8u8(g.fetch())
		}}, 0xC7: {MCycles: 4, Execute: func(g *Gameboy) {
			g.callu16(0x0000)
		}}, 0xC8: {MCycles: 2, Execute: func(g *Gameboy) {
			opcode := opcodes[0xC0]
			if g.cpu.registers.f.z {
				g.ret()
				opcode.MCycles = 5
			} else {
				opcode.MCycles = 2
			}
			opcodes[0xC0] = opcode
		}}, 0xC9: {MCycles: 4, Execute: func(g *Gameboy) {
			g.ret()
		}}, 0xCA: {MCycles: 3, Execute: func(g *Gameboy) {
			opcode := opcodes[0xC2]
			Addr := g.fetch16()
			if g.cpu.registers.f.z {
				g.jumpu16(Addr)
				opcode.MCycles = 4
			} else {
				opcode.MCycles = 3
			}
		}}, 0xCB: {MCycles: 2, Execute: func(g *Gameboy) {
			code := prefixOpCodes[g.fetch()]
			code(g)
		}}, 0xCC: {MCycles: 3, Execute: func(g *Gameboy) {
			OpCode := opcodes[0xC4]
			Addr := g.fetch16()
			if g.cpu.registers.f.z {
				g.callu16(Addr)
				OpCode.MCycles = 6
			} else {
				OpCode.MCycles = 3
			}
		}}, 0xCD: {MCycles: 2, Execute: func(g *Gameboy) {
			g.callu16(g.fetch16())
		}}, 0xCE: {MCycles: 4, Execute: func(g *Gameboy) {
			g.AdcAu8(g.fetch())
		}}, 0xCF: {MCycles: 4, Execute: func(g *Gameboy) {
			g.callu16(0x08)
		}}, 0xD0: {MCycles: 2, Execute: func(g *Gameboy) {
			opcode := opcodes[0xC0]
			if !g.cpu.registers.f.c {
				g.ret()
				opcode.MCycles = 5
			} else {
				opcode.MCycles = 2
			}
			opcodes[0xC0] = opcode
		}}, 0xD1: {MCycles: 3, Execute: func(g *Gameboy) {
			g.pop16(DE)
		}}, 0xD2: {MCycles: 3, Execute: func(g *Gameboy) {
			opcode := opcodes[0xC2]
			Addr := g.fetch16()
			if !g.cpu.registers.f.c {
				g.jumpu16(Addr)
				opcode.MCycles = 4
			} else {
				opcode.MCycles = 3
			}
			opcodes[0xC2] = opcode
		}}, 0xD4: {MCycles: 3, Execute: func(g *Gameboy) {
			OpCode := opcodes[0xC4]
			Addr := g.fetch16()
			if !g.cpu.registers.f.c {
				g.callu16(Addr)
				OpCode.MCycles = 6
			} else {
				OpCode.MCycles = 3
			}
		}}, 0xD5: {MCycles: 4, Execute: func(g *Gameboy) {
			g.push16(DE)
		}}, 0xD6: {MCycles: 2, Execute: func(g *Gameboy) {
			g.SubAu8(g.fetch())
		}}, 0xD7: {MCycles: 4, Execute: func(g *Gameboy) {
			g.callu16(0x10)
		}}, 0xD8: {MCycles: 2, Execute: func(g *Gameboy) {
			opcode := opcodes[0xC0]
			if g.cpu.registers.f.c {
				g.ret()
				opcode.MCycles = 5
			} else {
				opcode.MCycles = 2
			}
			opcodes[0xC0] = opcode
		}}, 0xD9: {MCycles: 4, Execute: func(g *Gameboy) {
			g.ret()
			g.enableInterrupts()
		}}, 0xDA: {MCycles: 3, Execute: func(g *Gameboy) {
			opcode := opcodes[0xC2]
			Addr := g.fetch16()
			if g.cpu.registers.f.c {
				g.jumpu16(Addr)
				opcode.MCycles = 4
			} else {
				opcode.MCycles = 3
			}
		}}, 0xDC: {MCycles: 3, Execute: func(g *Gameboy) {
			OpCode := opcodes[0xC4]
			Addr := g.fetch16()
			if g.cpu.registers.f.c {
				g.callu16(Addr)
				OpCode.MCycles = 6
			}
		}}, 0xDE: {MCycles: 4, Execute: func(g *Gameboy) {
			g.SbcAu8(g.fetch())
		}}, 0xDF: {MCycles: 4, Execute: func(g *Gameboy) {
			g.callu16(0x18)
		}}, 0xE0: {MCycles: 3, Execute: func(g *Gameboy) {
			addr := uint16(0xFF00) + uint16(g.fetch())
			g.memory.writeAddr(addr, g.cpu.registers.a)
		}}, 0xE1: {MCycles: 3, Execute: func(g *Gameboy) {
			g.pop16(HL)
		}}, 0xE2: {MCycles: 3, Execute: func(g *Gameboy) {
			addr := uint16(0xFF00)
			addr += uint16(g.cpu.registers.c)

			g.memory.writeAddr(addr, g.cpu.registers.a)
		}}, 0xE5: {MCycles: 4, Execute: func(g *Gameboy) {
			g.push16(HL)
		}}, 0xE6: {MCycles: 2, Execute: func(g *Gameboy) {
			g.AndAu8(g.fetch())
		}}, 0xE7: {MCycles: 4, Execute: func(g *Gameboy) {
			g.callu16(0x20)
		}}, 0xE8: {MCycles: 3, Execute: func(g *Gameboy) {
			g.AddSPi8(int8(g.fetch()))
		}}, 0xE9: {MCycles: 1, Execute: func(g *Gameboy) {
			g.jumpu16(g.cpu.registers.get16(HL))
		}}, 0xEA: {MCycles: 4, Execute: func(g *Gameboy) {
			g.Loadn16A(g.fetch16())
		}}, 0xEE: {MCycles: 2, Execute: func(g *Gameboy) {
			g.XorAu8(g.fetch())
		}}, 0xEF: {MCycles: 4, Execute: func(g *Gameboy) {
			fmt.Println("calling 0x28")
			g.callu16(0x28)
		}}, 0xF0: {MCycles: 3, Execute: func(g *Gameboy) {
			Addr := uint16(0xFF00) + uint16(g.fetch())
			g.cpu.registers.a = g.memory.readAddr(Addr)
		}}, 0xF1: {MCycles: 3, Execute: func(g *Gameboy) {
			g.pop16(AF)
		}}, 0xF2: {MCycles: 2, Execute: func(g *Gameboy) {
			addr := 0xFF00 + uint16(g.cpu.registers.c)
			g.cpu.registers.a = g.memory.readAddr(addr)

		}}, 0xF3: {MCycles: 1, Execute: func(g *Gameboy) {
			g.cpu.IME = false
		}}, 0xF5: {MCycles: 4, Execute: func(g *Gameboy) {
			g.push16(AF)
		}}, 0xF6: {MCycles: 2, Execute: func(g *Gameboy) {
			g.OrAu8(g.fetch())
		}}, 0xF7: {MCycles: 4, Execute: func(g *Gameboy) {
			g.callu16(0x30)
		}}, 0xF8: {MCycles: 3, Execute: func(g *Gameboy) {
			i8 := int8(g.fetch())
			sp := g.cpu.registers.sp
			result := int16(sp) + int16(i8)

			g.cpu.registers.set16(HL, uint16(result))

			g.cpu.registers.f.z = false
			g.cpu.registers.f.n = false

			spLow := sp & 0xFF
			i8Unsigned := uint16(uint8(i8))

			g.cpu.registers.f.h = ((spLow & 0x0F) + (i8Unsigned & 0x0F)) > 0x0F
			g.cpu.registers.f.c = (spLow + i8Unsigned) > 0xFF

		}}, 0xF9: {MCycles: 4, Execute: func(g *Gameboy) {
			g.Load16r16n(SP, g.cpu.registers.get16(HL))
		}}, 0xFA: {MCycles: 4, Execute: func(g *Gameboy) {
			Addr := g.fetch16()
			g.Loadr8u8(A, g.memory.readAddr(Addr))
		}}, 0xFB: {MCycles: 1, Execute: func(g *Gameboy) {
			g.EI()
		}}, 0xFE: {MCycles: 2, Execute: func(g *Gameboy) {
			g.CPAu8(g.fetch())
		}}, 0xFF: {MCycles: 4, Execute: func(g *Gameboy) {
			g.callu16(0x38)
		}},
	}
}

var prefixOpCodes = map[byte]PrefixOpCode{
	0x00: func(g *Gameboy) int {
		g.RLCr8(B)
		return 2
	}, 0x01: func(g *Gameboy) int {
		g.RLCr8(C)
		return 2
	}, 0x02: func(g *Gameboy) int {
		g.RLCr8(D)
		return 2
	}, 0x03: func(g *Gameboy) int {
		g.RLCr8(E)
		return 2
	}, 0x04: func(g *Gameboy) int {
		g.RLCr8(H)
		return 2
	}, 0x05: func(g *Gameboy) int {
		g.RLCr8(L)
		return 2
	}, 0x06: func(g *Gameboy) int {
		g.RLCr8(HLP)
		return 2
	}, 0x07: func(g *Gameboy) int {
		g.RLCr8(A)
		return 2
	}, 0x08: func(g *Gameboy) int {
		g.RRCr8(B)
		return 2
	}, 0x09: func(g *Gameboy) int {
		g.RRCr8(C)
		return 2
	}, 0x0A: func(g *Gameboy) int {
		g.RRCr8(D)
		return 2
	}, 0x0B: func(g *Gameboy) int {
		g.RRCr8(E)
		return 2
	}, 0x0C: func(g *Gameboy) int {
		g.RRCr8(H)
		return 2
	}, 0x0D: func(g *Gameboy) int {
		g.RRCr8(L)
		return 2
	}, 0x0E: func(g *Gameboy) int {
		g.RRCr8(HLP)
		return 2
	}, 0x0F: func(g *Gameboy) int {
		g.RRCr8(A)
		return 2
	}, 0x10: func(g *Gameboy) int {
		g.RLr8(B)
		return 2
	}, 0x11: func(g *Gameboy) int {
		g.RLr8(C)
		return 2
	}, 0x12: func(g *Gameboy) int {
		g.RLr8(D)
		return 2
	}, 0x13: func(g *Gameboy) int {
		g.RLr8(E)
		return 2
	}, 0x14: func(g *Gameboy) int {
		g.RLr8(H)
		return 2
	}, 0x15: func(g *Gameboy) int {
		g.RLr8(L)
		return 2
	}, 0x16: func(g *Gameboy) int {
		g.RLr8(HLP)
		return 2
	}, 0x17: func(g *Gameboy) int {
		g.RLr8(A)
		return 2
	}, 0x18: func(g *Gameboy) int {
		g.RRr8(B)
		return 2
	}, 0x19: func(g *Gameboy) int {
		g.RRr8(C)
		return 2
	}, 0x1A: func(g *Gameboy) int {
		g.RRr8(D)
		return 2
	}, 0x1B: func(g *Gameboy) int {
		g.RRr8(E)
		return 2
	}, 0x1C: func(g *Gameboy) int {
		g.RRr8(H)
		return 2
	}, 0x1D: func(g *Gameboy) int {
		g.RRr8(L)
		return 2
	}, 0x1E: func(g *Gameboy) int {
		g.RRr8(HLP)
		return 2
	}, 0x1F: func(g *Gameboy) int {
		g.RRr8(A)
		return 2
	}, 0x20: func(g *Gameboy) int {
		g.Slar8(B)
		return 2
	}, 0x21: func(g *Gameboy) int {
		g.Slar8(C)
		return 2
	}, 0x22: func(g *Gameboy) int {
		g.Slar8(D)
		return 2
	}, 0x23: func(g *Gameboy) int {
		g.Slar8(E)
		return 2
	}, 0x24: func(g *Gameboy) int {
		g.Slar8(H)
		return 2
	}, 0x25: func(g *Gameboy) int {
		g.Slar8(L)
		return 2
	}, 0x26: func(g *Gameboy) int {
		g.Slar8(HLP)
		return 2
	}, 0x27: func(g *Gameboy) int {
		g.Slar8(A)
		return 2
	}, 0x28: func(g *Gameboy) int {
		g.Srar8(B)
		return 2
	}, 0x29: func(g *Gameboy) int {
		g.Srar8(C)
		return 2
	}, 0x2A: func(g *Gameboy) int {
		g.Srar8(D)
		return 2
	}, 0x2B: func(g *Gameboy) int {
		g.Srar8(E)
		return 2
	}, 0x2C: func(g *Gameboy) int {
		g.Srar8(H)
		return 2
	}, 0x2D: func(g *Gameboy) int {
		g.Srar8(L)
		return 2
	}, 0x2E: func(g *Gameboy) int {
		g.Srar8(HLP)
		return 2
	}, 0x2F: func(g *Gameboy) int {
		g.Srar8(A)
		return 2
	}, 0x30: func(g *Gameboy) int {
		g.swap(B)
		return 2
	}, 0x31: func(g *Gameboy) int {
		g.swap(C)
		return 2
	}, 0x32: func(g *Gameboy) int {
		g.swap(D)
		return 2
	}, 0x33: func(g *Gameboy) int {
		g.swap(E)
		return 2
	}, 0x34: func(g *Gameboy) int {
		g.swap(H)
		return 2
	}, 0x35: func(g *Gameboy) int {
		g.swap(L)
		return 2
	}, 0x36: func(g *Gameboy) int {
		g.swap(HLP)
		return 2
	}, 0x37: func(g *Gameboy) int {
		g.swap(A)
		return 2
	}, 0x38: func(g *Gameboy) int {
		g.Srlr8(B)
		return 2
	}, 0x39: func(g *Gameboy) int {
		g.Srlr8(C)
		return 2
	}, 0x3A: func(g *Gameboy) int {
		g.Srlr8(D)
		return 2
	}, 0x3B: func(g *Gameboy) int {
		g.Srlr8(E)
		return 2
	}, 0x3C: func(g *Gameboy) int {
		g.Srlr8(H)
		return 2
	}, 0x3D: func(g *Gameboy) int {
		g.Srlr8(L)
		return 2
	}, 0x3E: func(g *Gameboy) int {
		g.Srlr8(HLP)
		return 2
	}, 0x3F: func(g *Gameboy) int {
		g.Srlr8(A)
		return 2
	}, 0x40: func(g *Gameboy) int {
		g.BIT(0, B)
		return 2
	}, 0x41: func(g *Gameboy) int {
		g.BIT(0, C)
		return 2
	}, 0x42: func(g *Gameboy) int {
		g.BIT(0, D)
		return 2
	}, 0x43: func(g *Gameboy) int {
		g.BIT(0, E)
		return 2
	}, 0x44: func(g *Gameboy) int {
		g.BIT(0, H)
		return 2
	}, 0x45: func(g *Gameboy) int {
		g.BIT(0, L)
		return 2
	}, 0x46: func(g *Gameboy) int {
		g.BIT(0, HLP)
		return 2
	}, 0x47: func(g *Gameboy) int {
		g.BIT(0, A)
		return 2
	}, 0x48: func(g *Gameboy) int {
		g.BIT(1, B)
		return 2
	}, 0x49: func(g *Gameboy) int {
		g.BIT(1, C)
		return 2
	}, 0x4A: func(g *Gameboy) int {
		g.BIT(1, D)
		return 2
	}, 0x4B: func(g *Gameboy) int {
		g.BIT(1, E)
		return 2
	}, 0x4C: func(g *Gameboy) int {
		g.BIT(1, H)
		return 2
	}, 0x4D: func(g *Gameboy) int {
		g.BIT(1, L)
		return 2
	}, 0x4E: func(g *Gameboy) int {
		g.BIT(1, HLP)
		return 2
	}, 0x4F: func(g *Gameboy) int {
		g.BIT(1, A)
		return 2
	}, 0x50: func(g *Gameboy) int {
		g.BIT(2, B)
		return 2
	}, 0x51: func(g *Gameboy) int {
		g.BIT(2, C)
		return 2
	}, 0x52: func(g *Gameboy) int {
		g.BIT(2, D)
		return 2
	}, 0x53: func(g *Gameboy) int {
		g.BIT(2, E)
		return 2
	}, 0x54: func(g *Gameboy) int {
		g.BIT(2, H)
		return 2
	}, 0x55: func(g *Gameboy) int {
		g.BIT(2, L)
		return 2
	}, 0x56: func(g *Gameboy) int {
		g.BIT(2, HLP)
		return 2
	}, 0x57: func(g *Gameboy) int {
		g.BIT(2, A)
		return 2
	}, 0x58: func(g *Gameboy) int {
		g.BIT(3, B)
		return 2
	}, 0x59: func(g *Gameboy) int {
		g.BIT(3, C)
		return 2
	}, 0x5A: func(g *Gameboy) int {
		g.BIT(3, D)
		return 2
	}, 0x5B: func(g *Gameboy) int {
		g.BIT(3, E)
		return 2
	}, 0x5C: func(g *Gameboy) int {
		g.BIT(3, H)
		return 2
	}, 0x5D: func(g *Gameboy) int {
		g.BIT(3, L)
		return 2
	}, 0x5E: func(g *Gameboy) int {
		g.BIT(3, HLP)
		return 2
	}, 0x5F: func(g *Gameboy) int {
		g.BIT(3, A)
		return 2
	}, 0x60: func(g *Gameboy) int {
		g.BIT(4, B)
		return 2
	}, 0x61: func(g *Gameboy) int {
		g.BIT(4, C)
		return 2
	}, 0x62: func(g *Gameboy) int {
		g.BIT(4, D)
		return 2
	}, 0x63: func(g *Gameboy) int {
		g.BIT(4, E)
		return 2
	}, 0x64: func(g *Gameboy) int {
		g.BIT(4, H)
		return 2
	}, 0x65: func(g *Gameboy) int {
		g.BIT(4, L)
		return 2
	}, 0x66: func(g *Gameboy) int {
		g.BIT(4, HLP)
		return 2
	}, 0x67: func(g *Gameboy) int {
		g.BIT(4, A)
		return 2
	}, 0x68: func(g *Gameboy) int {
		g.BIT(5, B)
		return 2
	}, 0x69: func(g *Gameboy) int {
		g.BIT(5, C)
		return 2
	}, 0x6A: func(g *Gameboy) int {
		g.BIT(5, D)
		return 2
	}, 0x6B: func(g *Gameboy) int {
		g.BIT(5, E)
		return 2
	}, 0x6C: func(g *Gameboy) int {
		g.BIT(5, H)
		return 2
	}, 0x6D: func(g *Gameboy) int {
		g.BIT(5, L)
		return 2
	}, 0x6E: func(g *Gameboy) int {
		g.BIT(5, HLP)
		return 2
	}, 0x6F: func(g *Gameboy) int {
		g.BIT(5, A)
		return 2
	}, 0x70: func(g *Gameboy) int {
		g.BIT(6, B)
		return 2
	}, 0x71: func(g *Gameboy) int {
		g.BIT(6, C)
		return 2
	}, 0x72: func(g *Gameboy) int {
		g.BIT(6, D)
		return 2
	}, 0x73: func(g *Gameboy) int {
		g.BIT(6, E)
		return 2
	}, 0x74: func(g *Gameboy) int {
		g.BIT(6, H)
		return 2
	}, 0x75: func(g *Gameboy) int {
		g.BIT(6, L)
		return 2
	}, 0x76: func(g *Gameboy) int {
		g.BIT(6, HLP)
		return 2
	}, 0x77: func(g *Gameboy) int {
		g.BIT(6, A)
		return 2
	}, 0x78: func(g *Gameboy) int {
		g.BIT(7, B)
		return 2
	}, 0x79: func(g *Gameboy) int {
		g.BIT(7, C)
		return 2
	}, 0x7A: func(g *Gameboy) int {
		g.BIT(7, D)
		return 2
	}, 0x7B: func(g *Gameboy) int {
		g.BIT(7, E)
		return 2
	}, 0x7C: func(g *Gameboy) int {
		g.BIT(7, H)
		return 2
	}, 0x7D: func(g *Gameboy) int {
		g.BIT(7, L)
		return 2
	}, 0x7E: func(g *Gameboy) int {
		g.BIT(7, HLP)
		return 2
	}, 0x7F: func(g *Gameboy) int {
		g.BIT(7, A)
		return 2
	}, 0x80: func(g *Gameboy) int {
		g.RES(0, B)
		return 2
	}, 0x81: func(g *Gameboy) int {
		g.RES(0, C)
		return 2
	}, 0x82: func(g *Gameboy) int {
		g.RES(0, D)
		return 2
	}, 0x83: func(g *Gameboy) int {
		g.RES(0, E)
		return 2
	}, 0x84: func(g *Gameboy) int {
		g.RES(0, H)
		return 2
	}, 0x85: func(g *Gameboy) int {
		g.RES(0, L)
		return 2
	}, 0x86: func(g *Gameboy) int {
		g.RES(0, HLP)
		return 2
	}, 0x87: func(g *Gameboy) int {
		g.RES(0, A)
		return 2
	}, 0x88: func(g *Gameboy) int {
		g.RES(1, B)
		return 2
	}, 0x89: func(g *Gameboy) int {
		g.RES(1, C)
		return 2
	}, 0x8A: func(g *Gameboy) int {
		g.RES(1, D)
		return 2
	}, 0x8B: func(g *Gameboy) int {
		g.RES(1, E)
		return 2
	}, 0x8C: func(g *Gameboy) int {
		g.RES(1, H)
		return 2
	}, 0x8D: func(g *Gameboy) int {
		g.RES(1, L)
		return 2
	}, 0x8E: func(g *Gameboy) int {
		g.RES(1, HLP)
		return 2
	}, 0x8F: func(g *Gameboy) int {
		g.RES(1, A)
		return 2
	}, 0x90: func(g *Gameboy) int {
		g.RES(2, B)
		return 2
	}, 0x91: func(g *Gameboy) int {
		g.RES(2, C)
		return 2
	}, 0x92: func(g *Gameboy) int {
		g.RES(2, D)
		return 2
	}, 0x93: func(g *Gameboy) int {
		g.RES(2, E)
		return 2
	}, 0x94: func(g *Gameboy) int {
		g.RES(2, H)
		return 2
	}, 0x95: func(g *Gameboy) int {
		g.RES(2, L)
		return 2
	}, 0x96: func(g *Gameboy) int {
		g.RES(2, HLP)
		return 2
	}, 0x97: func(g *Gameboy) int {
		g.RES(2, A)
		return 2
	}, 0x98: func(g *Gameboy) int {
		g.RES(3, B)
		return 2
	}, 0x99: func(g *Gameboy) int {
		g.RES(3, C)
		return 2
	}, 0x9A: func(g *Gameboy) int {
		g.RES(3, D)
		return 2
	}, 0x9B: func(g *Gameboy) int {
		g.RES(3, E)
		return 2
	}, 0x9C: func(g *Gameboy) int {
		g.RES(3, H)
		return 2
	}, 0x9D: func(g *Gameboy) int {
		g.RES(3, L)
		return 2
	}, 0x9E: func(g *Gameboy) int {
		g.RES(3, HLP)
		return 2
	}, 0x9F: func(g *Gameboy) int {
		g.RES(3, A)
		return 2
	}, 0xA0: func(g *Gameboy) int {
		g.RES(4, B)
		return 2
	}, 0xA1: func(g *Gameboy) int {
		g.RES(4, C)
		return 2
	}, 0xA2: func(g *Gameboy) int {
		g.RES(4, D)
		return 2
	}, 0xA3: func(g *Gameboy) int {
		g.RES(4, E)
		return 2
	}, 0xA4: func(g *Gameboy) int {
		g.RES(4, H)
		return 2
	}, 0xA5: func(g *Gameboy) int {
		g.RES(4, L)
		return 2
	}, 0xA6: func(g *Gameboy) int {
		g.RES(4, HLP)
		return 2
	}, 0xA7: func(g *Gameboy) int {
		g.RES(4, A)
		return 2
	}, 0xA8: func(g *Gameboy) int {
		g.RES(5, B)
		return 2
	}, 0xA9: func(g *Gameboy) int {
		g.RES(5, C)
		return 2
	}, 0xAA: func(g *Gameboy) int {
		g.RES(5, D)
		return 2
	}, 0xAB: func(g *Gameboy) int {
		g.RES(5, E)
		return 2
	}, 0xAC: func(g *Gameboy) int {
		g.RES(5, H)
		return 2
	}, 0xAD: func(g *Gameboy) int {
		g.RES(5, L)
		return 2
	}, 0xAE: func(g *Gameboy) int {
		g.RES(5, HLP)
		return 2
	}, 0xAF: func(g *Gameboy) int {
		g.RES(5, A)
		return 2
	}, 0xB0: func(g *Gameboy) int {
		g.RES(6, B)
		return 2
	}, 0xB1: func(g *Gameboy) int {
		g.RES(6, C)
		return 2
	}, 0xB2: func(g *Gameboy) int {
		g.RES(6, D)
		return 2
	}, 0xB3: func(g *Gameboy) int {
		g.RES(6, E)
		return 2
	}, 0xB4: func(g *Gameboy) int {
		g.RES(6, H)
		return 2
	}, 0xB5: func(g *Gameboy) int {
		g.RES(6, L)
		return 2
	}, 0xB6: func(g *Gameboy) int {
		g.RES(6, HLP)
		return 2
	}, 0xB7: func(g *Gameboy) int {
		g.RES(6, A)
		return 2
	}, 0xB8: func(g *Gameboy) int {
		g.RES(7, B)
		return 2
	}, 0xB9: func(g *Gameboy) int {
		g.RES(7, C)
		return 2
	}, 0xBA: func(g *Gameboy) int {
		g.RES(7, D)
		return 2
	}, 0xBB: func(g *Gameboy) int {
		g.RES(7, E)
		return 2
	}, 0xBC: func(g *Gameboy) int {
		g.RES(7, H)
		return 2
	}, 0xBD: func(g *Gameboy) int {
		g.RES(7, L)
		return 2
	}, 0xBE: func(g *Gameboy) int {
		g.RES(7, HLP)
		return 2
	}, 0xBF: func(g *Gameboy) int {
		g.RES(7, A)
		return 2
	}, 0xC0: func(g *Gameboy) int {
		g.SET(0, B)
		return 2
	}, 0xC1: func(g *Gameboy) int {
		g.SET(0, C)
		return 2
	}, 0xC2: func(g *Gameboy) int {
		g.SET(0, D)
		return 2
	}, 0xC3: func(g *Gameboy) int {
		g.SET(0, E)
		return 2
	}, 0xC4: func(g *Gameboy) int {
		g.SET(0, H)
		return 2
	}, 0xC5: func(g *Gameboy) int {
		g.SET(0, L)
		return 2
	}, 0xC6: func(g *Gameboy) int {
		g.SET(0, HLP)
		return 2
	}, 0xC7: func(g *Gameboy) int {
		g.SET(0, A)
		return 2
	}, 0xC8: func(g *Gameboy) int {
		g.SET(1, B)
		return 2
	}, 0xC9: func(g *Gameboy) int {
		g.SET(1, C)
		return 2
	}, 0xCA: func(g *Gameboy) int {
		g.SET(1, D)
		return 2
	}, 0xCB: func(g *Gameboy) int {
		g.SET(1, E)
		return 2
	}, 0xCC: func(g *Gameboy) int {
		g.SET(1, H)
		return 2
	}, 0xCD: func(g *Gameboy) int {
		g.SET(1, L)
		return 2
	}, 0xCE: func(g *Gameboy) int {
		g.SET(1, HLP)
		return 2
	}, 0xCF: func(g *Gameboy) int {
		g.SET(1, A)
		return 2
	}, 0xD0: func(g *Gameboy) int {
		g.SET(2, B)
		return 2
	}, 0xD1: func(g *Gameboy) int {
		g.SET(2, C)
		return 2
	}, 0xD2: func(g *Gameboy) int {
		g.SET(2, D)
		return 2
	}, 0xD3: func(g *Gameboy) int {
		g.SET(2, E)
		return 2
	}, 0xD4: func(g *Gameboy) int {
		g.SET(2, H)
		return 2
	}, 0xD5: func(g *Gameboy) int {
		g.SET(2, L)
		return 2
	}, 0xD6: func(g *Gameboy) int {
		g.SET(2, HLP)
		return 2
	}, 0xD7: func(g *Gameboy) int {
		g.SET(2, A)
		return 2
	}, 0xD8: func(g *Gameboy) int {
		g.SET(3, B)
		return 2
	}, 0xD9: func(g *Gameboy) int {
		g.SET(3, C)
		return 2
	}, 0xDA: func(g *Gameboy) int {
		g.SET(3, D)
		return 2
	}, 0xDB: func(g *Gameboy) int {
		g.SET(3, E)
		return 2
	}, 0xDC: func(g *Gameboy) int {
		g.SET(3, H)
		return 2
	}, 0xDD: func(g *Gameboy) int {
		g.SET(3, L)
		return 2
	}, 0xDE: func(g *Gameboy) int {
		g.SET(3, HLP)
		return 2
	}, 0xDF: func(g *Gameboy) int {
		g.SET(3, A)
		return 2
	}, 0xE0: func(g *Gameboy) int {
		g.SET(4, B)
		return 2
	}, 0xE1: func(g *Gameboy) int {
		g.SET(4, C)
		return 2
	}, 0xE2: func(g *Gameboy) int {
		g.SET(4, D)
		return 2
	}, 0xE3: func(g *Gameboy) int {
		g.SET(4, E)
		return 2
	}, 0xE4: func(g *Gameboy) int {
		g.SET(4, H)
		return 2
	}, 0xE5: func(g *Gameboy) int {
		g.SET(4, L)
		return 2
	}, 0xE6: func(g *Gameboy) int {
		g.SET(4, HLP)
		return 2
	}, 0xE7: func(g *Gameboy) int {
		g.SET(4, A)
		return 2
	}, 0xE8: func(g *Gameboy) int {
		g.SET(5, B)
		return 2
	}, 0xE9: func(g *Gameboy) int {
		g.SET(5, C)
		return 2
	}, 0xEA: func(g *Gameboy) int {
		g.SET(5, D)
		return 2
	}, 0xEB: func(g *Gameboy) int {
		g.SET(5, E)
		return 2
	}, 0xEC: func(g *Gameboy) int {
		g.SET(5, H)
		return 2
	}, 0xED: func(g *Gameboy) int {
		g.SET(5, L)
		return 2
	}, 0xEE: func(g *Gameboy) int {
		g.SET(5, HLP)
		return 2
	}, 0xEF: func(g *Gameboy) int {
		g.SET(5, A)
		return 2
	}, 0xF0: func(g *Gameboy) int {
		g.SET(6, B)
		return 2
	}, 0xF1: func(g *Gameboy) int {
		g.SET(6, C)
		return 2
	}, 0xF2: func(g *Gameboy) int {
		g.SET(6, D)
		return 2
	}, 0xF3: func(g *Gameboy) int {
		g.SET(6, E)
		return 2
	}, 0xF4: func(g *Gameboy) int {
		g.SET(6, H)
		return 2
	}, 0xF5: func(g *Gameboy) int {
		g.SET(6, L)
		return 2
	}, 0xF6: func(g *Gameboy) int {
		g.SET(6, HLP)
		return 2
	}, 0xF7: func(g *Gameboy) int {
		g.SET(6, A)
		return 2
	}, 0xF8: func(g *Gameboy) int {
		g.SET(7, B)
		return 2
	}, 0xF9: func(g *Gameboy) int {
		g.SET(7, C)
		return 2
	}, 0xFA: func(g *Gameboy) int {
		g.SET(7, D)
		return 2
	}, 0xFB: func(g *Gameboy) int {
		g.SET(7, E)
		return 2
	}, 0xFC: func(g *Gameboy) int {
		g.SET(7, H)
		return 2
	}, 0xFD: func(g *Gameboy) int {
		g.SET(7, L)
		return 2
	}, 0xFE: func(g *Gameboy) int {
		g.SET(7, HLP)
		return 2
	}, 0xFF: func(g *Gameboy) int {
		g.SET(7, A)
		return 2
	},
}
