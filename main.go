package main

import (
	"fmt"
	"image/color"
	"os"

	"gbabot/cpu"

	"github.com/hajimehoshi/ebiten"
)

type Game struct {
	gb *cpu.Gameboy
}

func (g *Game) Update(screen *ebiten.Image) error {
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Clear the screen first
	screen.Fill(color.RGBA{0, 0, 0, 255})

	// Draw the GameBoy screen
	g.gb.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 160, 144
}

func main() {
	gb := cpu.GBInitDebug()
	rompath := "./cpu/individual/acid.gb"
	romData, err := os.ReadFile(rompath)
	if err != nil {
		fmt.Println("Error loading ROM:", err)
		return
	}
	gb.LOADROM(romData)
	fmt.Printf("Game Boy initialized: %+v\n", gb != nil)
	manual := make(chan bool)

	go func() {
		for {
			fmt.Scanln()
			manual <- true
		}
	}()

	go gb.Start(nil, manual) // Pass nil for the done channel in this example

	ebiten.SetWindowSize(160*2, 144*2) // 4x scale
	ebiten.SetWindowTitle("ExGB - GameBoy Emulator")

	game := &Game{gb: gb}
	if err := ebiten.RunGame(game); err != nil {
		fmt.Println("Error running game:", err)
	}

}
