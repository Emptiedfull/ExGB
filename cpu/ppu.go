package cpu

const (
	LY             = 0xFF44
	LYC            = 0xFF45
	SCANLINECYCLES = 456

	LCDC = 0xFF40
	LCDS = 0xFF41

	SCROLLX = 0xFF43
	SCROLLY = 0xFF42
	WINDOWY = 0xFF4A
	WINDOWX = 0xFF4B
)

type mode int

const (
	Hblank mode = iota
	Vblank
	OamSearch
	Drawing
)

type Ppu struct {
	Scanline       int
	ScalineCounter int
}

func (g *Gameboy) UpdateGraphics(Mcycles int) {
	Cycles := Mcycles * 4

	if !g.lcdstatus() {
		return
	}

	g.ppu.ScalineCounter -= Cycles

	if g.ppu.ScalineCounter <= 0 {
		g.memory.mem[LY]++
		line := g.memory.readAddr(LY)

		g.ppu.ScalineCounter = SCANLINECYCLES

		if line == 144 {
			g.memory.EFSet(INTERRUPT_FLAG_VBLANK)
		}

		if line > 153 {
			g.memory.mem[LY] = 0
		}

		if line < 144 {
			g.DrawScanline()
		}
	}

}

func (g *Gameboy) DrawScanline() {

	control := g.memory.readAddr(LCDC)

	if control&0x01 == 0 {
		return
	}

	if testBit(control, 1) {
		g.RenderSprite()
	}

	if testBit(control, 0) {
		g.RenderBackground()
	}

}

func (g *Gameboy) RenderSprite() {

}

func (g *Gameboy) RenderBackground() {
	Scrolly := g.memory.readAddr(SCROLLY)
	Scrollx := g.memory.readAddr(SCROLLX)
	Windowy := g.memory.readAddr(WINDOWY)
	Windowx := g.memory.readAddr(WINDOWX) - 7

	usingWindow, Signed, tileData, backgroundMem := g.getTileSettings()

}

func (g *Gameboy) getTileSettings() (usingWindow bool, Signed bool, tileData uint16, backgroundMem uint16) {
	control := g.memory.readAddr(LCDC)
	Signed = false

	tileData = uint16(0x8800)
	backgroundMem = 0x9800
	windowY := g.memory.readAddr(WINDOWY)
	if testBit(control, 5) {
		if windowY <= g.memory.readAddr(LY) {
			usingWindow = true

		}
	}

	if testBit(control, 4) {
		tileData = 0x8000
	} else {
		tileData = 0x8800
		Signed = true
	}

	if usingWindow {
		if testBit(control, 6) {
			backgroundMem = 0x9C00
		} else {
			backgroundMem = 0x9800
		}
	} else {
		if testBit(control, 3) {
			backgroundMem = 0x9C00
		} else {
			backgroundMem = 0x9800
		}
	}

	return usingWindow, Signed, tileData, backgroundMem
}

func (g *Gameboy) getTileAddress() {
	if g.memory.readAddr(LCDC)&0x08 != 0 {

	}
}

func (g *Gameboy) lcdstatus() bool {
	return g.memory.readAddr(LCDC)&0x80 != 0
}

func (g *Gameboy) SetLcdStatus() {
	status := g.memory.readAddr(LCDS)

	if !g.lcdstatus() {
		g.ppu.Scanline = 0
		g.ppu.ScalineCounter = SCANLINECYCLES
		g.memory.mem[LY] = 0

		g.memory.setLcdMode(Vblank)

		return
	}

	currentLine := g.memory.readAddr(LY)
	currentMode := status & 0x03
	reqInterrupt := false

	if currentLine >= 144 {
		g.memory.setLcdMode(1)
		reqInterrupt = testBit(status, 4)
	} else {
		mode2bound := SCANLINECYCLES - 80
		mode3bound := SCANLINECYCLES - 172

		if currentLine >= uint8(mode2bound) {
			g.memory.setLcdMode(2)
			reqInterrupt = testBit(status, 5)
		} else {
			if currentLine >= uint8(mode3bound) {
				g.memory.setLcdMode(3)
			} else {
				g.memory.setLcdMode(0)
				reqInterrupt = testBit(status, 3)
			}
		}
	}

	if reqInterrupt && currentMode != g.memory.readAddr(LCDS)&0x03 {
		g.memory.EFSet(INTERRUPT_FLAG_LCD)
	}

	if g.memory.readAddr(LY) == g.memory.readAddr(LYC) {
		g.memory.writeAddr(LCDS, status|0x04)
		if testBit(status, 2) {
			g.memory.EFSet(INTERRUPT_FLAG_LCD)
		} else {
			g.memory.writeAddr(LCDS, status&^0x04)
		}
	}

	g.memory.writeAddr(LCDS, status)

}

func (m *Memory) setLcdMode(mo mode) {
	status := m.readAddr(LCDS) & 0xFC

	switch mo {
	case Hblank:
		status |= 0x00
	case Vblank:
		status |= 0x01
	case OamSearch:
		status |= 0x02
	case Drawing:
		status |= 0x03
	}

	m.writeAddr(LCDS, status)
}

func testBit(status uint8, bit int) bool {
	return (status & (1 << bit)) != 0
}
