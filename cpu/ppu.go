package cpu

import (
	"fmt"
)

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
	LYSHADOW       bool
	WindowLine     int
}

func (g *Gameboy) UpdateGraphics(Mcycles int) {

	Cycles := Mcycles * 4

	if !g.lcdstatus() {
		g.SetLcdStatus(0)
		return
	}

	g.ppu.ScalineCounter -= Cycles

	if g.ppu.ScalineCounter <= 0 {
		g.memory.mem[LY]++
		line := g.memory.readAddr(LY)
		g.SetLcdStatus(int(line))

		line = g.memory.readAddr(LY)

		g.ppu.ScalineCounter += SCANLINECYCLES

		if line == 144 {
			g.memory.EFSet(INTERRUPT_FLAG_VBLANK)
		}

		if line > 153 {
			g.memory.mem[LY] = 0
		}

		if line < 144 {
			g.DrawScanline(g.memory.readAddr(LCDC), int(line))
		}
	} else {
		g.SetLcdStatus(int(g.memory.mem[LY]))
	}

}

func (g *Gameboy) DrawScanline(control uint8, Scaline int) {

	if testBit(control, 0) {
		g.RenderBackground(control, Scaline)
	}
	if testBit(control, 1) {
		g.RenderSprite()
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
		xPos := g.memory.readAddr(0xFE00+index+1) - 8
		tileNum := g.memory.readAddr(0xFE00 + index + 2)
		attributes := g.memory.readAddr(0xFE00 + index + 3)

		if yPos > g.memory.readAddr(LY) || yPos+byte(ysize) <= g.memory.readAddr(LY) {
			continue
		}

		if linesprites >= 10 {
			break
		}
		linesprites++

		yflip := testBit(attributes, 6)
		xflip := testBit(attributes, 5)
		priority := testBit(attributes, 7)

		line := uint16(g.memory.readAddr(LY) - yPos)
		if yflip {
			line = uint16(ysize - int(line) - 1)
		}

		var dataAddr uint16
		if ysize == 16 {
			if line < 8 {
				dataAddr = uint16(tileNum&0xFE)*16 + 0x8000 + line*2
			} else {
				dataAddr = uint16((tileNum&0xFE)+1)*16 + 0x8000 + (line-8)*2
			}
		} else {
			dataAddr = uint16(tileNum)*16 + 0x8000 + line*2
		}
		data1 := g.memory.readAddr(dataAddr)
		data2 := g.memory.readAddr(dataAddr + 1)

		for x := 7; x >= 0; x-- {
			colorbit := x

			if xflip {
				colorbit = 7 - x
			}

			colorNum := ((data2 >> colorbit) & 1) << 1
			colorNum |= (data1 >> colorbit) & 1

			if colorNum == 0 {
				continue
			}

			var colorAddr uint16
			if testBit(attributes, 4) {
				colorAddr = 0xFF49
			} else {
				colorAddr = 0xFF48
			}

			_, _, _, bs := g.getColor(colorNum, colorAddr)

			xPix := 7 - uint8(x) + xPos

			if xPix >= 160 {
				continue
			}

			if priority {

				if g.tileScanline[xPix] != 0 {
					continue
				}
			}

			g.SetPixel(priority, int(xPix), int(g.memory.readAddr(LY)), bs)

		}

	}

}

func (g *Gameboy) RenderBackground(control uint8, Scanline int) {
	Scrolly := g.memory.readAddr(SCROLLY)
	Scrollx := g.memory.readAddr(SCROLLX)

	Windowy := g.memory.readAddr(WINDOWY)
	Windowx := uint8(int(g.memory.readAddr(WINDOWX)) - 7)

	windowEnabled, Signed, tileData, _ := g.getTileSettings(control)

	windowthisline := windowEnabled && (uint8(Scanline) >= Windowy) && (Windowx <= 166)

	if windowthisline {
		if Scanline == int(Windowy) {
			g.ppu.WindowLine = 0
		}
	}

	for x := 0; x < 160; x++ {
		pixel := uint8(x)

		usingWindow := windowEnabled && pixel >= Windowx && uint8(Scanline) >= Windowy

		var xPos, yPos uint8
		var backgroundMem uint16

		if usingWindow {

			xPos = pixel - Windowx
			yPos = uint8(g.ppu.WindowLine)
			if testBit(control, 6) {
				backgroundMem = 0x9C00
			} else {
				backgroundMem = 0x9800
			}
		} else {
			// Background coordinates - with scrolling
			xPos = pixel + Scrollx
			yPos = Scrolly + uint8(Scanline)
			// Use background tile map
			if testBit(control, 3) {
				backgroundMem = 0x9C00
			} else {
				backgroundMem = 0x9800
			}
		}

		tileRow := uint16(yPos/8) * 32
		tileCol := uint16(xPos / 8)
		tileAddr := backgroundMem + tileRow + tileCol

		tileLoc := tileData

		if Signed {
			tileNum := int16(int8(g.memory.readAddr(tileAddr)))
			tileLoc = uint16(int32(tileLoc) + int32(tileNum)*16)
		} else {
			tileNum := int16(g.memory.readAddr(tileAddr))
			tileLoc = tileLoc + uint16(tileNum)*16
		}

		line := yPos % 8
		line *= 2

		data1 := g.memory.readAddr(tileLoc + uint16(line))
		data2 := g.memory.readAddr(tileLoc + uint16(line) + 1)

		colorBit := 7 - (xPos % 8)
		colorNum := ((data2 >> colorBit) & 1) << 1
		colorNum |= (data1 >> colorBit) & 1

		g.tileScanline[x] = colorNum

		_, _, _, bs := g.getColor(colorNum, 0xFF47)

		g.SetPixel(false, x, Scanline, bs)
	}

	if windowthisline {
		g.ppu.WindowLine++
	}
}

func (gb *Gameboy) getColor(colorNum uint8, pallete uint16) (byte, byte, byte, byte) {

	palette := gb.memory.readAddr(pallete)

	// Extract the 2-bit color value from the palette
	colorIndex := (palette >> (colorNum * 2)) & 0x03

	// Convert to RGB values (Game Boy grayscale)
	switch colorIndex {
	case 0:
		return 255, 255, 255, 0 // White
	case 1:
		return 192, 192, 192, 1 // Light gray
	case 2:
		return 96, 96, 96, 2 // Dark gray
	case 3:
		return 0, 0, 0, 3 // Black
	default:
		fmt.Println("Invalid color index:", colorIndex)
		return 255, 255, 255, 0 // Default to white
	}
}

func (gb *Gameboy) SetPixel(priority bool, x int, y int, bs byte) {
	if x >= 0 && x < 160 && y >= 0 && y < 144 {
		gb.WebScreen[x][y] = bs
	}

}

func (g *Gameboy) getTileSettings(control uint8) (usingWindow bool, Signed bool, tileData uint16, backgroundMem uint16) {
	Signed = false

	tileData = uint16(0x8800)
	backgroundMem = 0x9800
	usingWindow = false

	if testBit(control, 5) {
		windowY := g.memory.readAddr(WINDOWY)
		currentLine := g.memory.readAddr(LY)

		// Window is potentially active if current scanline >= window Y position
		if currentLine >= windowY {
			usingWindow = true
		}
	}

	if testBit(control, 4) {
		tileData = 0x8000
	} else {
		tileData = 0x9000
		Signed = true
	}
	return usingWindow, Signed, tileData, backgroundMem
}

func (g *Gameboy) lcdstatus() bool {
	return g.memory.readAddr(LCDC)&0x80 != 0
}

func (g *Gameboy) SetLcdStatus(Scanline int) {
	status := g.memory.readAddr(LCDS)
	currentMode := status & 0x03
	newStatus := status & 0xFC
	if !g.lcdstatus() {
		g.ppu.Scanline = 0
		g.ppu.ScalineCounter = SCANLINECYCLES
		g.memory.mem[LY] = 0
		newStatus &= 0xFC
		newStatus |= 0x00
		g.memory.writeAddr(LCDS, newStatus)

		return
	}

	reqInterrupt := false

	if g.memory.readAddr(LY) >= 144 {
		newStatus |= 0x01
		reqInterrupt = testBit(status, 4)
	} else {

		currentCycle := SCANLINECYCLES - g.ppu.ScalineCounter

		if currentCycle < 80 {
			// Mode 2: OAM Search
			newStatus |= 0x02
			reqInterrupt = testBit(status, 5)
		} else if currentCycle < 252 {
			// Mode 3: Drawing
			newStatus |= 0x03
			// if currentMode != 0x03 {
			// 	g.DrawScanline(g.memory.readAddr(LCDC), Scanline)
			// }
		} else {
			// Mode 0: H-Blank
			newStatus |= 0x00
			reqInterrupt = testBit(status, 3)
		}
	}

	if reqInterrupt && currentMode != (newStatus&0x03) {

		g.memory.EFSet(INTERRUPT_FLAG_LCD)
	}

	if g.memory.readAddr(LY) == g.memory.readAddr(LYC) {
		newStatus |= 0x04
	} else {
		newStatus &= 0xFB
	}

	if testBit(status, 6) && (newStatus&0x04) != 0 {
		if !g.ppu.LYSHADOW {

			g.memory.EFSet(INTERRUPT_FLAG_LCD)
			g.ppu.LYSHADOW = true
		}
	} else {
		g.ppu.LYSHADOW = false
	}
	g.memory.writeAddr(LCDS, newStatus)
}

func testBit(status uint8, bit int) bool {
	return (status & (1 << bit)) != 0
}
