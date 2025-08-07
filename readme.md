# ExGB - WebAssembly GameBoy Emulator

A high-performance Game Boy emulator built in Go and compiled to WebAssembly, featuring a modern web interface with real-time gameplay, save states, and customizable controls.

## 🎮 Features

- **Full Game Boy Emulation**: Complete CPU, PPU, and memory management implementation
- **WebAssembly Performance**: Native-speed emulation running directly in your browser
- **Save States**: Save and load game progress instantly
- **Custom Controls**: Remap keyboard controls to your preference
- **Speed Control**: Adjust emulation speed for different gameplay experiences
- **ROM Library**: Pre-loaded with classic Game Boy titles
- **Real-time Rendering**: 60 FPS gameplay with accurate graphics emulation
- **Local Operation**: No server dependencies - runs entirely in your browser

## 🚀 Quick Start

### Prerequisites
- Go 1.21 or higher
- Modern web browser with WebAssembly support

### Building and Running

1. **Clone the repository**
   ```bash
   git clone https://github.com/Emptiedfull/ExGB.git
   cd ExGB
   ```

2. **Build the WebAssembly module**
   ```bash
   GOOS=js GOARCH=wasm go build -o frontend/static/scripts/main.wasm
   cp $(go env GOROOT)/misc/wasm/wasm_exec.js frontend/static/scripts/
   ```

3. **Serve the application**
   ```bash
   # Using Python
   cd frontend/static
   python3 -m http.server 8080
   
   # Using Caddy (recommended)
   cd /workspaces/ExGB
   caddy start --config Caddyfile
   ```

4. **Open in browser**
   ```
   http://localhost:8080
   ```

## 🎯 Usage

### Playing Games

1. Navigate to the **PLAY** section
2. Select a ROM from the available library:
   - **Tetris** - Classic puzzle game
   - **Super Mario Land** - Platform adventure
   - **The Legend of Zelda: Link's Awakening** - Action RPG
   - **Metroid II** - Sci-fi exploration
   - **Dr. Mario** - Puzzle game
   - **Kirby's Dream Land** - Platform adventure

3. Click **START** to begin playing

### Controls

Default keyboard mapping:
- **Arrow Keys**: D-Pad (↑↓←→)
- **A Key**: A Button
- **B Key**: B Button
- **Enter**: Start
- **Escape**: Select

### Save States

- **Save**: Press Select + A to quick save
- **Load**: Press Select + B to quick load
- **Reset**: Press Select + Start to reset game

### Settings (GEAR)

- **Speed Control**: Adjust emulation speed (0.5x to 2x)
- **Key Remapping**: Click any control to remap keys
- **Feedback**: Report issues or suggestions

## 🏗️ Architecture

### Core Components

```
ExGB/
├── main.go              # WASM entry point and JS bindings
├── cpu/                 # Game Boy emulation core
│   ├── cpu.go          # CPU implementation (LR35902)
│   ├── gameboy.go      # Main system controller
│   ├── memory.go       # Memory management unit
│   ├── ppu.go          # Picture Processing Unit
│   ├── joypad.go       # Input handling
│   └── opcodes.go      # CPU instruction set
└── frontend/
    └── static/
        ├── index.html   # Main application
        ├── scripts/     # JavaScript and WASM files
        └── styles/      # CSS styling
```

### WASM Functions

The emulator exposes these functions to JavaScript:

- `initGameboy()` - Initialize emulator
- `loadRom(romData)` - Load ROM from Uint8Array
- `startGame()` - Begin emulation
- `sendInput(pressed, key)` - Handle input
- `saveState()` - Create save state
- `loadState(stateData)` - Restore save state
- `changeSpeed(multiplier)` - Adjust emulation speed
- `endGame()` - Stop emulation

## 🔧 Development

### Building for Development

```bash
GOOS=js GOARCH=wasm go build -o frontend/static/scripts/main.wasm

### Adding New ROMs

1. Place `.gb` files in `frontend/static/roms`
2. Add ROM entry to `frontend/static/index.html`:
   ```html
   <img src="static/images/roms/your-rom.png" alt="Your ROM" 
        class="rom-item" data-rom="your-rom">
   ```
3. Add corresponding ROM image to `frontend/static/images/roms`

## 🎨 Technical Details

### Emulation Accuracy

- **CPU**: Full LR35902 instruction set implementation
- **Graphics**: Complete PPU with sprite rendering and background layers
- **Memory**: Accurate memory mapping and banking
- **Timing**: Cycle-accurate emulation for proper game compatibility
- **Input**: Real-time joypad state management

### Performance Features

- **60 FPS**: Maintains original Game Boy framerate
- **Low Latency**: WebGL support for minimal display lag
- **Memory Efficient**: Optimized WASM compilation with minimal overhead
- **Browser Compatible**: Works in all modern browsers


## 🐛 Known Issues

- Some ROM compatibility edge cases
- Audio emulation not yet implemented

## 🔗 Links

- **Live Demo**: [wasm.ex-gb.com](https://wasm.ex-gb.com)
- **GitHub**: [github.com/Emptiedfull/ExGB](https://github.com/Emptiedfull/ExGB)
- **Issues**: [Report bugs here](https://github.com/Emptiedfull/ExGB/issues)

---

**©2025 NOT NINTENDO** - This is an educational project for emulation learning purposes.