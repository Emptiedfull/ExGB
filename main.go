package main

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"os"
	"sync"

	"gbabot/cpu"

	"github.com/gorilla/websocket"
)

type Game struct {
	gb *cpu.Gameboy

	Viewers     []*Viewer
	Control     *websocket.Conn
	Viewer_chan chan [160][144]byte

	Public bool

	ID string
}

type Viewer struct {
	oldscreen [160][144]byte
	conn      *websocket.Conn
}

type GameStore struct {
	sync.RWMutex
	Games map[string]*Game
}

type RomUpdate struct {
	Rom   string `json:"rom"`
	State []byte `json:"state,omitempty"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {

	Games := make(map[string]*Game)
	GamesStore := &GameStore{
		Games: Games,
	}

	setUpHttpServer(GamesStore)

	// game := &Game{gb: gb, Oam: Oam}
	// if err := ebiten.RunGame(game); err != nil {
	// 	fmt.Println("Error running game:", err)
	// }

	// Games = append(Games, game)

}

func setUpHttpServer(store *GameStore) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to ExGB - GameBoy Emulator!"))
	})

	http.HandleFunc("/servers/count", func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		store.RLock()
		count := 10 - len(store.Games)
		store.RUnlock()
		fmt.Println("Received request for server count", count)
		w.Write([]byte(fmt.Sprintf("%d", count)))
	})

	http.HandleFunc("/servers/status", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Received request for server status")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		id := r.URL.Query().Get("id")
		if id == "" || len(id) != 5 {
			http.Error(w, "Game ID is required", http.StatusBadRequest)
			fmt.Println("Invalid game ID:", id)
			return
		}

		store.RLock()
		game := store.Games[id]
		store.RUnlock()

		if game == nil {
			http.Error(w, "Game not found", http.StatusNotFound)
			fmt.Println("Game not found for ID:", id)
			return
		}

		w.Write([]byte("Game is active"))
	})

	http.HandleFunc("/ws/start", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		public := r.URL.Query().Get("public") == "true"
		if err != nil {
			fmt.Println("Error upgrading connection:", err)
			return
		}

		createGame(conn, store, public)
	})

	http.HandleFunc("/ws/view/", func(w http.ResponseWriter, r *http.Request) {

		path := r.URL.Path
		id := path[len("/ws/view/"):]

		if id == "" || len(id) != 5 {
			http.Error(w, "Game ID is required", http.StatusBadRequest)
			fmt.Println("Invalid game ID:", id)
			return
		}

		store.RLock()
		Game := store.Games[id]
		store.RUnlock()
		if Game == nil {
			http.Error(w, "Game not found", http.StatusNotFound)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println("Error upgrading connection:", err)
			return
		}

		Game.Viewers = append(Game.Viewers, &Viewer{conn: conn})
		fmt.Printf("New viewer connected to game %s, total viewers: %d\n", id, len(Game.Viewers))

	})

	http.ListenAndServe(":8080", nil)
}

func createGame(user *websocket.Conn, store *GameStore, public bool) {

	roms := map[string]string{
		"tetris":    "./games/tet.gb",
		"sumar":     "./games/sumar.gb",
		"pok green": "./games/pok2.gb",
		"mar":       "./games/mar.gb",
		"zelda":     "./games/zelda.gb",
		"link":      "./games/link.gb",
		"poke red":  "./games/red.gb",
		"kirby":     "./games/kirby.gb",
		"metroid":   "./games/metroid.gb",
	}

	store.Lock()
	if len(store.Games) >= 10 {
		user.WriteJSON(map[string]string{"error": "Maximum number of games reached"})
		user.Close()
		return
	}
	store.Unlock()

	ViewerChan := make(chan [160][144]byte, 10)
	ControlChan := make(chan cpu.JoypadUpdate, 10)
	endchan := make(chan bool)
	donechan := make(chan bool)
	pauseChan := make(chan bool)

	gb := cpu.GBInitDebug(ViewerChan, ControlChan)

	Game := &Game{
		ID:      genRandomID(),
		gb:      gb,
		Control: user,
		Public:  public,
	}
	go HandleGameEnd(store, Game, endchan, donechan)

	Game.Viewer_chan = ViewerChan
	store.Lock()
	store.Games[Game.ID] = Game
	store.Unlock()
	user.WriteJSON(map[string]string{"id": Game.ID})
	fmt.Println("waiting for user to send rom path...")

	var rom RomUpdate
	err := user.ReadJSON(&rom)
	if err != nil {
		fmt.Println("Error reading ROM path from user:", err)
		user.WriteJSON(map[string]string{"error": "Invalid ROM path"})
		user.Close()
		endchan <- true
		return
	}
	fmt.Println("Received message from user:", string(rom.Rom))

	if rom.Rom == "load" {
		romdata := rom.State
		gb.LoadState(romdata)
	} else {
		romPath, exists := roms[string(rom.Rom)]
		if !exists {
			user.WriteJSON(map[string]string{"error": "ROM not found"})
			fmt.Println("ROM not found for message:", string(rom.Rom))
			user.Close()
			endchan <- true
			return
		}

		romData, err := os.ReadFile(romPath)
		if err != nil {
			fmt.Println("Error loading ROM:", err)
			user.WriteJSON(map[string]string{"error": "Error loading ROM"})
			user.Close()
			endchan <- true
			return
		}

		gb.LOADROM(romData)
	}

	user.WriteJSON(map[string]string{"success": "rom loaded"})

	go UpdateScreen(Game)
	go HandleControl(Game, endchan, pauseChan)

	go gb.Start(donechan, nil, pauseChan)

	// rompath := "./cpu/individual/acid.gb"
	// rompath := "./tet.gb"
	// romData, err := os.ReadFile(rompath)
	// if err != nil {
	// 	fmt.Println("Error loading ROM:", err)
	// 	return
	// }
	// gb.LOADROM(romData)
	// fmt.Println("New game created with ID:", Game.ID)

	// user.WriteJSON(map[string]string{"id": Game.ID})

	// go gb.Start(nil, nil)

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

func UpdateScreen(g *Game) {
	for screen := range g.Viewer_chan {
		flatBytes := PackFrameBuffer(screen)

		activeViewers := g.Viewers[:0]
		for _, viewer := range g.Viewers {
			if !g.Public {
				if viewer.conn != g.Control {
					continue
				}
			}
			err := viewer.conn.WriteMessage(websocket.BinaryMessage, flatBytes)
			if err != nil {
				viewer.conn.Close()
			} else {
				activeViewers = append(activeViewers, viewer)
			}
		}
		g.Viewers = activeViewers
	}
}

func HandleControl(g *Game, endchan chan bool, pauseChan chan bool) {
	for {
		var update cpu.JoypadUpdate
		err := g.Control.ReadJSON(&update)
		if err != nil {
			fmt.Println("Error reading control update:", err)
			g.Control.Close()
			endchan <- true
			return
		}
		if update.Meta == "" {
			g.gb.ControlChan <- update
		}

		if update.Meta == "savestate" && update.SaveState {
			state := g.gb.SaveState()
			g.Control.WriteMessage(websocket.BinaryMessage, state)
		}

		if update.Meta == "pause" {
			pauseChan <- update.Pause
		}

	}
}

func HandleGameEnd(store *GameStore, game *Game, endchan, donechan chan bool) {
	<-endchan
	donechan <- true
	fmt.Println("Game ended:", game.ID)

	for _, viewer := range game.Viewers {
		viewer.conn.Close()
	}
	store.Lock()
	delete(store.Games, game.ID)
	store.Unlock()
	game.Control.Close()
	fmt.Println("Game removed from active games:", game.ID)
}

func genRandomID() string {
	letters := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	result := make([]byte, 5)

	randomBytes := make([]byte, 5)
	rand.Read(randomBytes)

	for i, b := range randomBytes {
		result[i] = letters[b%byte(len(letters))]
	}

	return string(result)
}
