package cpu

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten"
)

type OAMViewer struct {
	Gb         *Gameboy
	SpriteSize int
}

type SpriteData struct {
	Ypos       uint8
	Xpos       uint8
	Tile       uint8
	Attributes uint8
	Priority   bool
	FlipX      bool
	FlipY      bool
	Palette    uint8
}

func (o *OAMViewer) GetSpriteData(spriteIndex int) SpriteData {
	mem := o.Gb.memory.mem
	spriteBase := 0xFE00 + (spriteIndex * 4)

	return SpriteData{
		Ypos:       mem[spriteBase],
		Xpos:       mem[spriteBase+1],
		Tile:       mem[spriteBase+2],
		Attributes: mem[spriteBase+3],
		Priority:   testBit(mem[spriteBase+3], 7),
		FlipX:      testBit(mem[spriteBase+3], 6),
		FlipY:      testBit(mem[spriteBase+3], 5),
		Palette:    mem[spriteBase+3] & 0x10,
	}
}

func (o *OAMViewer) RenderSpriteTile(spriteIndex int, scale int) *ebiten.Image {
	sprite := o.GetSpriteData(spriteIndex) // Example for sprite 0

	if testBit(o.Gb.memory.readAddr(LCDC), 2) {
		o.SpriteSize = 16
	} else {
		o.SpriteSize = 8
	}

	width := 8 * scale
	height := o.SpriteSize * scale
	img, err := ebiten.NewImage(width, height, ebiten.FilterDefault)
	if err != nil {
		fmt.Println("Error creating image:", err)
		return nil
	}

	img.Fill(color.RGBA{0, 0, 0, 255}) // Fill with black background

	tileNum := sprite.Tile
	if o.SpriteSize == 16 {
		tileNum &= 0xFE
	}

	for y := 0; y < o.SpriteSize; y++ {
		var dataAddr uint16

		if o.SpriteSize == 16 {
			if y < 8 {
				dataAddr = uint16(tileNum)*16 + 0x8000 + uint16(y)*2
			} else {
				dataAddr = uint16(tileNum+1)*16 + 0x8000 + uint16(y-8)*2
			}

		} else {
			dataAddr = uint16(tileNum)*16 + 0x8000 + uint16(y)*2
		}

		data1 := o.Gb.memory.readAddr(dataAddr)
		data2 := o.Gb.memory.readAddr(dataAddr + 1)

		for x := 0; x < 8; x++ {
			colorBit := 7 - x
			if sprite.FlipX {
				colorBit = x
			}

			colorNum := ((data1 >> colorBit) & 1) | (((data2 >> colorBit) & 1) << 1)

			var paletteAddr uint16
			if sprite.Palette == 0 {
				paletteAddr = 0xFF48 // BG Palette
			} else {
				paletteAddr = 0xFF49 // Sprite Palette
			}

			r, g, b, _ := o.Gb.getColor(colorNum, paletteAddr)

			var pixelColor color.RGBA
			if colorNum == 0 {
				pixelColor = color.RGBA{255, 255, 255, 128} // Transparent white
			} else {
				pixelColor = color.RGBA{r, g, b, 255}
			}

			yPos := y
			if sprite.FlipY {
				yPos = o.SpriteSize - 1 - y
			}

			for sy := 0; sy < scale; sy++ {
				for sx := 0; sx < scale; sx++ {
					img.Set(x*scale+sx, yPos*scale+sy, pixelColor)
				}
			}

		}
	}

	return img
}

func (o *OAMViewer) DrawOAMTable(screen *ebiten.Image, x, y, scale int) {
	const spritesPerRow = 10

	const spriteSpacing = 2

	for i := 0; i < 40; i++ {
		row := i / spritesPerRow
		col := i % spritesPerRow

		spriteImg := o.RenderSpriteTile(i, scale)

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x+col*(8*scale+spriteSpacing)), float64(y+row*(o.SpriteSize*scale+spriteSpacing)))
		screen.DrawImage(spriteImg, op)

	}
}
