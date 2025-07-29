package cpu

func (g *Gameboy) controlHandler() {
	for update := range g.ControlChan {
		if update.Pressed {
			g.HandleKeyPress(update.Key)
		} else {

			g.memory.joypadState = setBit(g.memory.joypadState, uint8(update.Key))

		}
	}
}

func (g *Gameboy) HandleKeyPress(key int) {

	g.memory.joypadState = resetBit(g.memory.joypadState, key)

	selectByte := g.memory.mem[0xFF00] & 0x30
	if (key < 4 && selectByte&0x10 == 0) || (key >= 4 && selectByte&0x20 == 0) {
		g.memory.EFSet(INTERRUPT_FLAG_JOYPAD)
	}
}

func resetBit(val uint8, bit int) uint8 {
	if bit < 0 || bit > 7 {
		return val
	}
	return val & ^(1 << bit)
}

func (m *Memory) getJoyPadState() uint8 {
	selectByte := m.mem[0xFF00] & 0x30
	result := selectByte | 0xCF

	if selectByte&0x10 == 0 {

		result &= (m.joypadState >> 4) | 0xF0

	}
	if selectByte&0x20 == 0 {

		result &= (m.joypadState & 0x0F) | 0xF0
	}
	return result
}
