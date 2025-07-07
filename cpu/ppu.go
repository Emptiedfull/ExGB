package cpu

import "fmt"

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

		// fmt.Println("Updating scanline", line)

		g.ppu.ScalineCounter += SCANLINECYCLES

		if line == 144 {
			g.memory.EFSet(INTERRUPT_FLAG_VBLANK)
		}

		if line > 153 {
			g.memory.mem[LY] = 0
		}

		if line < 144 {
			fmt.Println("Drawing scanline", line)
			g.DrawScanline()
		}
	}

}

func (g *Gameboy) DrawScanline() {

	control := g.memory.readAddr(LCDC)

	if control&0x01 == 0 {
		fmt.Println("LCD is off, skipping rendering")
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

	ysize := 8
	if testBit(g.memory.readAddr(LCDC), 2) {
		ysize = 16
	}

	linesprites := 0
	for sprite := 0; sprite < 40; sprite++ {

		index := uint16(sprite * 4)
		yPos := g.memory.readAddr(0xFE00+index) - 16
		if yPos > g.memory.readAddr(LY) || yPos+byte(ysize) <= g.memory.readAddr(LY) {
			continue
		}

		if linesprites >= 10 {
			break
		}
		linesprites++

		xPos := g.memory.readAddr(0xFE00+index+1) - 8
		tileNum := g.memory.readAddr(0xFE00 + index + 2)
		attributes := g.memory.readAddr(0xFE00 + index + 3)

		yflip := testBit(attributes, 6)
		xflip := testBit(attributes, 5)
		priority := testBit(attributes, 7)

		line := uint16(g.memory.readAddr(LY) - yPos)
		if yflip {
			line = uint16(ysize - int(line) - 1)
		}

		line *= 2

		dataAddr := uint16(tileNum)*16 + 0x8000 + line
		data1 := g.memory.readAddr(dataAddr)
		data2 := g.memory.readAddr(dataAddr + 1)

		for x := 7; x >= 0; x-- {
			colorbit := x

			if xflip {
				colorbit = 7 - x
			}

			colorNum := ((data2 >> colorbit) & 1) << 1
			colorNum |= (data1 >> colorbit) & 1
			var colorAddr uint16
			if testBit(attributes, 4) {
				colorAddr = 0xFF49
			} else {
				colorAddr = 0xFF48
			}

			r, gr, b := g.getColor(colorNum, colorAddr)

			if r == 0 && gr == 0 && b == 0 {
				continue
			}

			xPix := 7 - uint8(x) + xPos
			g.SetPixel(priority, int(xPix), int(g.memory.readAddr(LY)), r, gr, b)

		}

	}

}

func (g *Gameboy) RenderBackground() {
	Scrolly := g.memory.readAddr(SCROLLY)
	Scrollx := g.memory.readAddr(SCROLLX)
	Windowy := g.memory.readAddr(WINDOWY)
	Windowx := g.memory.readAddr(WINDOWX) - 7

	usingWindow, Signed, tileData, backgroundMem := g.getTileSettings()

	var yPos byte
	if !usingWindow {
		yPos = Scrolly + g.memory.readAddr(LY)

	} else {
		yPos = g.memory.readAddr(LY) - Windowy
	}

	tileRow := uint16(yPos/8) * 32

	for x := 0; x < 160; x++ {
		pixel := uint8(x)
		xPos := pixel + Scrollx

		if usingWindow {
			if pixel >= Windowx {
				xPos = pixel - Windowx
			}
		}

		tileCol := uint16(xPos / 8)
		tileAddr := backgroundMem + tileRow + tileCol

		tileLoc := tileData

		if Signed {
			tileNum := int16(int8(g.memory.readAddr(tileAddr)))
			tileLoc = uint16(int32(tileLoc) + int32(tileNum+128)*16)
		} else {
			tileNum := int16(g.memory.readAddr(tileAddr))
			tileLoc = tileLoc + uint16(tileNum)*16
		}

		line := yPos % 8
		line *= 2

		data1 := g.memory.readAddr(tileLoc + uint16(line))
		data2 := g.memory.readAddr(tileLoc + uint16(line) + 1)

		colorBit := 7 - (xPos % 8)

		// Extract the color bits from both bytes
		colorNum := ((data2 >> colorBit) & 1) << 1
		colorNum |= (data1 >> colorBit) & 1

		g.tileScanline[x] = colorNum

		r, gr, b := g.getColor(colorNum, 0xFF47)
		g.SetPixel(false, x, int(g.memory.readAddr(LY)), r, gr, b)
	}

}

func (gb *Gameboy) getColor(colorNum uint8, pallete uint16) (byte, byte, byte) {

	palette := gb.memory.readAddr(pallete)

	// Extract the 2-bit color value from the palette
	colorIndex := (palette >> (colorNum * 2)) & 0x03

	// Convert to RGB values (Game Boy grayscale)
	switch colorIndex {
	case 0:
		return 255, 255, 255 // White
	case 1:
		return 192, 192, 192 // Light gray
	case 2:
		return 96, 96, 96 // Dark gray
	case 3:
		return 0, 0, 0 // Black
	default:
		fmt.Println("Invalid color index:", colorIndex)
		return 255, 255, 255 // Default to white
	}
}

func (gb *Gameboy) SetPixel(priority bool, x int, y int, r, g, b byte) {

	if gb.tileScanline[x] != 0 && !priority {
		fmt.Println("Skipping pixel at", x, y, "due to priority")
		return
	}

	fmt.Println("Setting pixel at", x, y, "to color", r, g, b)

	gb.Screen[x][y][0] = r
	gb.Screen[x][y][1] = g
	gb.Screen[x][y][2] = b

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
