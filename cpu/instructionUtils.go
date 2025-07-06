package cpu

func (g *Gameboy) Load16r16n(r r16, n uint16) {
	g.cpu.registers.set16(r, n)
}

func (g *Gameboy) Loadr8r8(r1 r8, r2 r8) {
	val := g.get8(r2)
	g.set8(r1, val)
}

func (g *Gameboy) Loadr8HL(r r8) {
	val := g.memory.readAddr(g.cpu.registers.get16(HL))
	g.set8(r, val)
}

func (g *Gameboy) LoadHLr8(r r8) {
	val := g.get8(r)
	g.memory.writeAddr(g.cpu.registers.get16(HL), val)
}

func (g *Gameboy) Loadr8u8(r r8, n uint8) {
	g.set8(r, n)
}

func (g *Gameboy) LoadAn16(r r16) {
	val := g.memory.readAddr(g.cpu.registers.get16(r))
	g.set8(A, val)
}

func (g *Gameboy) LoadMEMA(r r16) {
	addr := g.cpu.registers.get16(r)
	g.memory.writeAddr(addr, g.cpu.registers.a)
}

func (g *Gameboy) Loadr16MEMA(r r16) {
	addr := g.cpu.registers.get16(r)
	val := g.memory.readAddr(addr)
	// fmt.Println("Loading value from address", addr, ":", val)
	g.set8(A, val)
}

func (g *Gameboy) Loadn16A(n uint16) {
	g.memory.writeAddr(n, g.cpu.registers.a)
}

func (g *Gameboy) inc16(r r16) {
	val := g.cpu.registers.get16(r)
	val++
	g.cpu.registers.set16(r, val)
}

func (g *Gameboy) inc8(r r8) {
	val := g.get8(r)
	res := (val + 1) & 0xFF

	g.cpu.registers.f.z = res == 0
	g.cpu.registers.f.n = false
	g.cpu.registers.f.h = (val & 0x0F) == 0x0F

	if res == 0 {
		g.cpu.registers.f.z = true

	}

	// fmt.Println("setting register to", res, "zero flag", g.cpu.registers.f.z)
	g.set8(r, res)
}

func (g *Gameboy) dec8(r r8) {
	val := g.get8(r)
	res := (val - 1) & 0xFF

	g.cpu.registers.f.z = res == 0
	g.cpu.registers.f.n = true
	g.cpu.registers.f.h = (val & 0x0F) == 0

	if res == 0 {
		g.cpu.registers.f.z = true
		// fmt.Println("Decrementing", r, "resulted in zero")
	}
	g.set8(r, res)
}

func (g *Gameboy) dec16(r r16) {
	val := g.cpu.registers.get16(r)
	val--
	g.cpu.registers.set16(r, val)
}

func (g *Gameboy) RLCA() {
	val := g.cpu.registers.a

	g.cpu.registers.f.c = (val & 0x80) != 0
	val = (val << 1) | (val >> 7)

	g.cpu.registers.f.z = false
	g.cpu.registers.f.n = false
	g.cpu.registers.f.h = false

	g.set8(A, val)
}

func (g *Gameboy) RLCr8(r r8) {
	val := g.get8(r)

	g.cpu.registers.f.c = (val & 0x80) != 0
	val = (val << 1) | (val >> 7)

	g.cpu.registers.f.z = val == 0
	g.cpu.registers.f.n = false
	g.cpu.registers.f.h = false

	g.set8(r, val)
}

func (g *Gameboy) RLr8(r r8) {

	val := g.get8(r)
	var carry uint8
	if g.cpu.registers.f.c {
		carry = 1
	} else {
		carry = 0
	}

	g.cpu.registers.f.n = false
	g.cpu.registers.f.h = false
	g.cpu.registers.f.c = (val & 0x80) != 0
	val = (val << 1) | (uint8(carry) & 0x01)
	g.cpu.registers.f.z = val == 0
	g.set8(r, val)

}

func (g *Gameboy) RLA() {

	val := g.get8(A)
	var carry uint8
	if g.cpu.registers.f.c {
		carry = 1
	} else {
		carry = 0
	}

	g.cpu.registers.f.n = false
	g.cpu.registers.f.h = false
	g.cpu.registers.f.c = (val & 0x80) != 0
	val = (val << 1) | (uint8(carry) & 0x01)
	g.cpu.registers.f.z = false

	g.set8(A, val)

}

func (g *Gameboy) RRr8(r r8) {
	val := g.get8(r)
	var carry uint8
	if g.cpu.registers.f.c {
		carry = 0x80
	} else {
		carry = 0
	}

	g.cpu.registers.f.n = false
	g.cpu.registers.f.h = false
	g.cpu.registers.f.c = (val & 0x01) != 0
	val = (val >> 1) | carry

	g.cpu.registers.f.z = val == 0

	g.set8(r, val)
}

func (g *Gameboy) RRA() {
	val := g.get8(A)
	var carry uint8
	if g.cpu.registers.f.c {
		carry = 0x80
	} else {
		carry = 0
	}

	g.cpu.registers.f.n = false
	g.cpu.registers.f.h = false
	g.cpu.registers.f.c = (val & 0x01) != 0
	val = (val >> 1) | carry
	g.cpu.registers.f.z = false
	g.set8(A, val)
}

func (g *Gameboy) RRCr8(r r8) {
	val := g.get8(r)

	g.cpu.registers.f.n = false
	g.cpu.registers.f.h = false
	g.cpu.registers.f.c = (val & 0x01) != 0
	val = (val >> 1) | (val << 7)
	g.cpu.registers.f.z = val == 0
	g.set8(r, val)
}

func (g *Gameboy) RRCA() {
	val := g.get8(A)

	g.cpu.registers.f.n = false
	g.cpu.registers.f.h = false
	g.cpu.registers.f.c = (val & 0x01) != 0
	val = (val >> 1) | (val << 7)
	g.cpu.registers.f.z = false

	g.set8(A, val)
}

func (g *Gameboy) AddHLr16(r r16) {
	add := g.cpu.registers.get16(r)
	val := g.cpu.registers.get16(HL)
	result32 := uint32(add) + uint32(val)
	result16 := uint16(result32)

	g.cpu.registers.f.n = false
	g.cpu.registers.f.h = ((add & 0x0FFF) + (val & 0x0FFF)) > 0x0FFF
	g.cpu.registers.f.c = result32 > 0xFFFF
	g.cpu.registers.set16(HL, result16)
}

func (g *Gameboy) Addr8r8(r1 r8, r2 r8) {
	value := g.get8(r2)
	originalValue := g.get8(r1)
	result16 := uint16(originalValue) + uint16(value)
	result8 := uint8(result16)

	g.cpu.registers.f.z = result8 == 0
	g.cpu.registers.f.n = false
	g.cpu.registers.f.h = (originalValue&0x0F)+(value&0x0F) > 0x0F
	g.cpu.registers.f.c = result16 > 0xFF

	g.set8(r1, result8)
}

func (g *Gameboy) Addr8u8(n uint8) {
	originalValue := g.cpu.registers.a
	result16 := uint16(originalValue) + uint16(n)
	result8 := uint8(result16)

	g.cpu.registers.f.z = result8 == 0
	g.cpu.registers.f.n = false
	g.cpu.registers.f.h = (originalValue&0x0F)+(uint8(n)&0x0F) > 0x0F
	g.cpu.registers.f.c = result16 > 0xFF

	g.set8(A, result8)
}

func (g *Gameboy) Adcr8r8(r1 r8, r2 r8) {
	value := g.get8(r2)
	originalValue := g.get8(r1)

	var carry uint8 = 0
	if g.cpu.registers.f.c {
		carry = 1
	}
	result16 := uint16(originalValue) + uint16(value) + uint16(carry)
	result8 := uint8(result16)

	g.cpu.registers.f.z = result8 == 0
	g.cpu.registers.f.n = false
	g.cpu.registers.f.h = (originalValue&0x0F)+(value&0x0F)+carry > 0x0F
	g.cpu.registers.f.c = result16 > 0xFF

	g.set8(r1, result8)

}

func (g *Gameboy) AdcAu8(n uint8) {
	value := n
	originalValue := g.get8(A)

	var carry uint8 = 0
	if g.cpu.registers.f.c {
		carry = 1
	}
	result16 := uint16(originalValue) + uint16(value) + uint16(carry)
	result8 := uint8(result16)

	g.cpu.registers.f.z = result8 == 0
	g.cpu.registers.f.n = false
	g.cpu.registers.f.h = (originalValue&0x0F)+(value&0x0F)+carry > 0x0F
	g.cpu.registers.f.c = result16 > 0xFF

	g.set8(A, result8)

}

func (g *Gameboy) Adcr8HL(r1 r8) {
	Addr := g.cpu.registers.get16(HL)
	value := g.memory.readAddr(Addr)
	originalValue := g.get8(r1)

	var carry uint8 = 0
	if g.cpu.registers.f.c {
		carry = 1
	}
	result16 := uint16(originalValue) + uint16(value) + uint16(carry)
	result8 := uint8(result16)

	g.cpu.registers.f.z = result8 == 0
	g.cpu.registers.f.n = false
	g.cpu.registers.f.h = (originalValue&0x0F)+(value&0x0F)+carry > 0x0F
	g.cpu.registers.f.c = result16 > 0xFF

	g.set8(r1, result8)

}

func (g *Gameboy) Addr8HL(r1 r8) {
	Addr := g.cpu.registers.get16(HL)
	value := g.memory.readAddr(Addr)
	originalValue := g.get8(r1)
	result16 := uint16(originalValue) + uint16(value)
	result8 := uint8(result16)

	g.cpu.registers.f.z = result8 == 0
	g.cpu.registers.f.n = false
	g.cpu.registers.f.h = (originalValue&0x0F)+(value&0x0F) > 0x0F
	g.cpu.registers.f.c = result16 > 0xFF

	g.set8(r1, result8)
}

func (g *Gameboy) Subr8r8(r1 r8, r2 r8) {
	value := g.get8(r2)
	originalValue := g.get8(r1)
	result16 := uint16(originalValue) - uint16(value)
	result8 := uint8(result16)

	g.cpu.registers.f.z = result8 == 0
	g.cpu.registers.f.n = true
	g.cpu.registers.f.h = (originalValue & 0x0F) < (value & 0x0F)
	g.cpu.registers.f.c = originalValue < value

	g.set8(r1, result8)
}
func (g *Gameboy) SubAu8(n uint8) {
	value := n
	originalValue := g.get8(A)
	result16 := uint16(originalValue) - uint16(value)
	result8 := uint8(result16)

	g.cpu.registers.f.z = result8 == 0
	g.cpu.registers.f.n = true
	g.cpu.registers.f.h = (originalValue & 0x0F) < (value & 0x0F)
	g.cpu.registers.f.c = originalValue < value

	g.set8(A, result8)
}

func (g *Gameboy) Subr8HL(r1 r8) {
	Addr := g.cpu.registers.get16(HL)
	value := g.memory.readAddr(Addr)
	originalValue := g.get8(r1)
	result16 := uint16(originalValue) - uint16(value)
	result8 := uint8(result16)

	g.cpu.registers.f.z = result8 == 0
	g.cpu.registers.f.n = true
	g.cpu.registers.f.h = (originalValue & 0x0F) < (value & 0x0F)
	g.cpu.registers.f.c = originalValue < value

	g.set8(r1, result8)
}

func (g *Gameboy) Sbcr8r8(r1 r8, r2 r8) {
	value := g.get8(r2)
	originalValue := g.get8(r1)

	var carry uint8 = 0
	if g.cpu.registers.f.c {
		carry = 1
	}

	result := int16(originalValue) - int16(value) - int16(carry)
	result8 := uint8(result)

	g.cpu.registers.f.z = result8 == 0
	g.cpu.registers.f.n = true

	g.cpu.registers.f.h = int16(originalValue&0x0F)-int16(value&0x0F)-int16(carry) < 0
	g.cpu.registers.f.c = result < 0

	g.set8(r1, result8)
}

func (g *Gameboy) SbcAu8(n uint8) {
	value := n
	originalValue := g.get8(A)

	var carry uint8 = 0
	if g.cpu.registers.f.c {
		carry = 1
	}

	result := int16(originalValue) - int16(value) - int16(carry)
	result8 := uint8(result)

	g.cpu.registers.f.z = result8 == 0
	g.cpu.registers.f.n = true

	g.cpu.registers.f.h = int16(originalValue&0x0F)-int16(value&0x0F)-int16(carry) < 0
	g.cpu.registers.f.c = result < 0

	g.set8(A, result8)
}

func (g *Gameboy) Sbcr8HL(r1 r8) {
	Addr := g.cpu.registers.get16(HL)
	value := g.memory.readAddr(Addr)
	originalValue := g.get8(r1)

	var carry uint8 = 0
	if g.cpu.registers.f.c {
		carry = 1
	}

	result := int16(originalValue) - int16(value) - int16(carry)
	result8 := uint8(result)

	g.cpu.registers.f.z = result8 == 0
	g.cpu.registers.f.n = true

	g.cpu.registers.f.h = int16(originalValue&0x0F)-int16(value&0x0F)-int16(carry) < 0
	g.cpu.registers.f.c = result < 0

	g.set8(r1, result8)
}

func (g *Gameboy) Andr8r8(r1 r8, r2 r8) {
	value := g.get8(r2)
	originalValue := g.get8(r1)
	result := originalValue & value

	g.cpu.registers.f.z = result == 0
	g.cpu.registers.f.n = false
	g.cpu.registers.f.h = true
	g.cpu.registers.f.c = false

	g.set8(r1, result)
}

func (g *Gameboy) AndAu8(n uint8) {
	value := n
	originalValue := g.get8(A)
	result := originalValue & value

	g.cpu.registers.f.z = result == 0
	g.cpu.registers.f.n = false
	g.cpu.registers.f.h = true
	g.cpu.registers.f.c = false

	g.set8(A, result)
}

func (g *Gameboy) Andr8HL(r1 r8) {
	Addr := g.cpu.registers.get16(HL)
	value := g.memory.readAddr(Addr)
	originalValue := g.get8(r1)
	result := originalValue & value

	g.cpu.registers.f.z = result == 0
	g.cpu.registers.f.n = false
	g.cpu.registers.f.h = true
	g.cpu.registers.f.c = false

	g.set8(r1, result)
}

func (g *Gameboy) AddSPi8(n int8) {
	sp := g.cpu.registers.sp

	result := int32(sp) + int32(n)

	g.cpu.registers.f.z = false
	g.cpu.registers.f.n = false

	spLow := sp & 0xFF
	nUnsigned := uint16(uint8(n))

	g.cpu.registers.f.h = ((spLow & 0x0F) + (nUnsigned & 0x0F)) > 0x0F
	g.cpu.registers.f.c = (spLow + nUnsigned) > 0xFF

	g.cpu.registers.sp = uint16(result)

}

func (g *Gameboy) Xorr8r8(r1 r8, r2 r8) {
	value := g.get8(r2)
	originalValue := g.get8(r1)
	result := originalValue ^ value

	g.cpu.registers.f.z = result == 0
	g.cpu.registers.f.n = false
	g.cpu.registers.f.h = false
	g.cpu.registers.f.c = false

	g.set8(r1, result)
}

func (g *Gameboy) XorAu8(n uint8) {
	value := n
	originalValue := g.get8(A)
	result := originalValue ^ value

	g.cpu.registers.f.z = result == 0
	g.cpu.registers.f.n = false
	g.cpu.registers.f.h = false
	g.cpu.registers.f.c = false

	g.set8(A, result)
}

func (g *Gameboy) Xorr8HL(r1 r8) {
	Addr := g.cpu.registers.get16(HL)
	value := g.memory.readAddr(Addr)
	originalValue := g.get8(r1)
	result := originalValue ^ value

	g.cpu.registers.f.z = result == 0
	g.cpu.registers.f.n = false
	g.cpu.registers.f.h = false
	g.cpu.registers.f.c = false

	g.set8(r1, result)
}

func (g *Gameboy) Or8r8(r1 r8, r2 r8) {
	value := g.get8(r2)
	originalValue := g.get8(r1)
	result := originalValue | value

	g.cpu.registers.f.z = result == 0
	g.cpu.registers.f.n = false
	g.cpu.registers.f.h = false
	g.cpu.registers.f.c = false

	g.set8(r1, result)
}

func (g *Gameboy) OrAu8(n uint8) {
	value := n
	originalValue := g.get8(A)
	result := originalValue | value

	g.cpu.registers.f.z = result == 0
	g.cpu.registers.f.n = false
	g.cpu.registers.f.h = false
	g.cpu.registers.f.c = false

	g.set8(A, result)
}

func (g *Gameboy) Or8HL(r1 r8) {
	Addr := g.cpu.registers.get16(HL)
	value := g.memory.readAddr(Addr)
	originalValue := g.get8(r1)
	result := originalValue | value

	g.cpu.registers.f.z = result == 0
	g.cpu.registers.f.n = false
	g.cpu.registers.f.h = false
	g.cpu.registers.f.c = false

	g.set8(r1, result)
}

func (g *Gameboy) CPr8r8(r1 r8, r2 r8) {
	comparer := g.get8(r1)
	value := g.get8(r2)
	result := comparer - value

	g.cpu.registers.f.z = result == 0
	g.cpu.registers.f.n = true
	g.cpu.registers.f.h = (comparer & 0x0F) < (value & 0x0F)
	g.cpu.registers.f.c = comparer < value

}

func (g *Gameboy) CPAu8(n uint8) {
	comparer := g.get8(A)
	value := n
	result := comparer - value

	g.cpu.registers.f.z = result == 0
	g.cpu.registers.f.n = true
	g.cpu.registers.f.h = (comparer & 0x0F) < (value & 0x0F)
	g.cpu.registers.f.c = comparer < value

}

func (g *Gameboy) CPr8HL(r1 r8) {
	Addr := g.cpu.registers.get16(HL)
	comparer := g.get8(r1)
	value := g.memory.readAddr(Addr)
	result := comparer - value

	g.cpu.registers.f.z = result == 0
	g.cpu.registers.f.n = true
	g.cpu.registers.f.h = (comparer & 0x0F) < (value & 0x0F)
	g.cpu.registers.f.c = comparer < value
}

func (g *Gameboy) ret() {
	g.cpu.pc = g.popStack16()
}

func (g *Gameboy) popStack16() uint16 {
	low := g.memory.readAddr(g.cpu.registers.sp)
	g.cpu.registers.sp++
	high := g.memory.readAddr(g.cpu.registers.sp)
	g.cpu.registers.sp++
	return uint16(high)<<8 | uint16(low)
}

func (g *Gameboy) pushStack16(r uint16) {
	val := r
	g.cpu.registers.sp--
	g.memory.writeAddr(g.cpu.registers.sp, uint8(val>>8))
	g.cpu.registers.sp--
	g.memory.writeAddr(g.cpu.registers.sp, uint8(val&0xFF))
}

func (g *Gameboy) push16(r r16) {
	val := g.cpu.registers.get16(r)
	g.pushStack16(val)
}

func (g *Gameboy) pop16(r r16) {
	val := g.popStack16()
	g.cpu.registers.set16(r, val)
}

func (g *Gameboy) jumpu16(n uint16) {
	g.cpu.pc = n
}

func (g *Gameboy) callu16(n uint16) {
	g.pushStack16(g.cpu.pc)
	g.jumpu16(n)
}

func (g *Gameboy) EI() {
	g.cpu.IMESet = true
}

func (g *Gameboy) enableInterrupts() {
	g.cpu.IME = true
	g.cpu.IMESet = false
}

func (g *Gameboy) Slar8(r r8) {
	val := g.get8(r)

	g.cpu.registers.f.c = (val & 0x80) != 0
	val = val << 1

	g.cpu.registers.f.z = val == 0
	g.cpu.registers.f.n = false
	g.cpu.registers.f.h = false

	g.set8(r, val)
}

func (g *Gameboy) Srar8(r r8) {
	val := g.get8(r)
	bit7 := val & 0x80

	g.cpu.registers.f.n = false
	g.cpu.registers.f.h = false
	g.cpu.registers.f.c = (val & 0x01) != 0
	val = (val >> 1) | bit7
	g.cpu.registers.f.z = val == 0
	g.set8(r, val)
}

func (g *Gameboy) Srlr8(r r8) {
	val := g.get8(r)

	g.cpu.registers.f.c = (val & 0x01) != 0
	val = val >> 1

	g.cpu.registers.f.z = val == 0
	g.cpu.registers.f.n = false
	g.cpu.registers.f.h = false

	g.set8(r, val)
}

func (g *Gameboy) swap(r r8) {
	val := g.get8(r)
	upper := (val & 0xF0) >> 4
	lower := (val & 0x0F) << 4
	swapped := upper | lower

	g.cpu.registers.f.z = swapped == 0
	g.cpu.registers.f.n = false
	g.cpu.registers.f.h = false
	g.cpu.registers.f.c = false

	g.set8(r, swapped)
}

func (g *Gameboy) BIT(u int, r r8) {
	val := g.get8(r)
	mask := uint8(1 << u)

	g.cpu.registers.f.z = (val & mask) == 0
	g.cpu.registers.f.n = false
	g.cpu.registers.f.h = true
}

func (g *Gameboy) RES(u int8, r r8) {
	val := g.get8(r)
	mask := uint8(1 << u)

	g.set8(r, val&^mask)

}

func (g *Gameboy) SET(u int8, r r8) {
	val := g.get8(r)
	mask := uint8(1 << u)

	g.set8(r, val|mask)
}
