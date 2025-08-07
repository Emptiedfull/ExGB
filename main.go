package main

import (
	"fmt"
	"gbabot/cpu"
	"syscall/js"
)

type game struct {
	gb          *cpu.Gameboy
	viewerChan  chan [160][144]byte
	controlChan chan cpu.JoypadUpdate

	romLoaded bool
	running   bool

	doneChan  chan bool
	pauseChan chan bool
	endChan   chan bool
	oldFrame  [160][144]byte
}

var Game *game

func main() {
	js.Global().Set("initGameboy", js.FuncOf(InitGameboy))
	js.Global().Set("sendInput", js.FuncOf(sendInput))
	js.Global().Set("loadRom", js.FuncOf(loadRom))
	js.Global().Set("startGame", js.FuncOf(startGame))
	js.Global().Set("pauseGame", js.FuncOf(pauseGame))
	js.Global().Set("resumeGame", js.FuncOf(resumeGame))
	js.Global().Set("endGame", js.FuncOf(endGame))
	js.Global().Set("saveState", js.FuncOf(getState))
	js.Global().Set("loadState", js.FuncOf(loadState))
	js.Global().Set("changeSpeed", js.FuncOf(ChangeSpeed))
	js.Global().Set("checkGameState", js.FuncOf(CheckGameState))

	<-make(chan bool)
}

func CheckGameState(this js.Value, args []js.Value) interface{} {
	return Game == nil
}

func loadState(this js.Value, args []js.Value) interface{} {

	if len(args) < 1 {
		return "No state data provided"
	}

	jsArray := args[0]

	length := jsArray.Get("length").Int()
	if length == 0 {
		return "Empty state data"
	}

	stateData := make([]byte, length)
	js.CopyBytesToGo(stateData, jsArray)

	if len(stateData) != length {
		return fmt.Sprintf("Data copy failed: expected %d bytes, got %d", length, len(stateData))
	}

	err := Game.gb.LoadState(stateData)
	if err != nil {
		return fmt.Sprintf("Failed to load state: %v", err)
	}

	Game.romLoaded = true

	return "State loaded successfully"
}

func getState(this js.Value, args []js.Value) interface{} {

	if Game == nil || !Game.romLoaded {
		return "Gameboy not initialized or ROM not loaded"
	}

	state := Game.gb.SaveState()
	if state == nil {
		return "Failed to get state"
	}
	stateData := make([]byte, len(state))
	copy(stateData, state)
	uint8Array := js.Global().Get("Uint8Array").New(len(stateData))
	js.CopyBytesToJS(uint8Array, stateData)
	return uint8Array
}

func startGame(this js.Value, args []js.Value) interface{} {
	if Game.gb == nil || !Game.romLoaded {
		return "Gameboy not initialized or ROM not loaded"
	}

	if Game.running {
		return "Game already running"
	}
	Game.running = true

	Game.doneChan = make(chan bool)
	Game.pauseChan = make(chan bool)
	Game.endChan = make(chan bool)

	go Game.gb.Start(Game.doneChan, nil, Game.pauseChan)

	go func() {
		for {
			select {
			case frame := <-Game.viewerChan:
				if frame != Game.oldFrame {
					Game.oldFrame = frame
					frameData := PackFrameBuffer(frame)

					uint8Array := js.Global().Get("Uint8Array").New(len(frameData))
					js.CopyBytesToJS(uint8Array, frameData)

					js.Global().Call("onFrameUpdate", uint8Array)
				}
			case <-Game.endChan:
				return
			}
		}
	}()

	return "Game started"
}

func PackFrameBuffer(screen [160][144]byte) []byte {
	out := make([]byte, 160*144/4)
	i := 0
	for y := 0; y < 144; y++ {
		for x := 0; x < 160; x += 4 {
			var b byte
			b |= (screen[x][y] & 0x03) << 6
			b |= (screen[x+1][y] & 0x03) << 4
			b |= (screen[x+2][y] & 0x03) << 2
			b |= (screen[x+3][y] & 0x03)
			out[i] = b
			i++
		}
	}
	return out
}

func loadRom(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return "No ROM provided"
	}

	if Game.gb == nil {
		return "Gameboy not initialized"
	}

	if Game.romLoaded {
		return "ROM already loaded"
	}

	jsArray := args[0]
	if jsArray.Type() != js.TypeObject {
		return "Invalid ROM data type"
	}

	// Get the length of the array
	length := jsArray.Get("length").Int()
	if length == 0 {
		return "Empty ROM data"
	}

	// Create a Go byte slice and copy data from JavaScript
	romData := make([]byte, length)
	js.CopyBytesToGo(romData, jsArray)

	// Load ROM into gameboy
	Game.gb.LOADROM(romData)

	Game.romLoaded = true
	return "ROM loaded successfully"
}

func ChangeSpeed(this js.Value, args []js.Value) interface{} {
	if Game.gb == nil {
		return "Gameboy not initialized"
	}

	if len(args) < 1 {
		return "No speed multiplier provided"
	}

	mult := args[0].Int()
	Game.gb.ChangeSpeed(mult)
	return fmt.Sprintf("Game speed changed to %d", mult)
}

func InitGameboy(this js.Value, args []js.Value) interface{} {
	viewerChan := make(chan [160][144]byte)
	controlChan := make(chan cpu.JoypadUpdate)
	gb := cpu.GBInitDebug(viewerChan, controlChan)

	Game = &game{
		gb:          gb,
		viewerChan:  viewerChan,
		controlChan: controlChan,
	}

	return "gameboy initialized"
}

func pauseGame(this js.Value, args []js.Value) interface{} {
	if Game.gb == nil || !Game.romLoaded {
		return "Gameboy not initialized or ROM not loaded"
	}

	Game.pauseChan <- true
	return "Game paused"

}

func endGame(this js.Value, args []js.Value) interface{} {
	Game.endChan <- true
	Game = nil
	return "Game ended"
}

func resumeGame(this js.Value, args []js.Value) interface{} {
	if Game.gb == nil || !Game.romLoaded {
		return "Gameboy not initialized or ROM not loaded"
	}

	Game.pauseChan <- false
	return "Game resumed"
}

func sendInput(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		return "No input provided"
	}

	if Game.gb == nil || !Game.romLoaded {
		return "Gameboy not initialized"
	}

	pressed := args[0].Bool()
	key := args[1].Int()

	if key < 0 || key > 7 {
		return "Invalid key index"
	}

	update := cpu.JoypadUpdate{
		Key:     key,
		Pressed: pressed,
	}

	select {
	case Game.gb.ControlChan <- update:
		return "input sent"
	default:
		fmt.Println("Control channel is full, dropping input")
		return "control channel full, input dropped"
	}

}
