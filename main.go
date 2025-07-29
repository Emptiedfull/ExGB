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
	Rom string `json:"rom"`
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

	http.HandleFunc("/ws/start", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println("Error upgrading connection:", err)
			return
		}

		createGame(conn, store)
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

func createGame(user *websocket.Conn, store *GameStore) {

	roms := map[string]string{
		"acid":   "./cpu/individual/acid.gb",
		"tetris": "./tet.gb",
		"sumar":  "./mar.gb",
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
	gb := cpu.GBInitDebug(ViewerChan, ControlChan)

	Game := &Game{
		ID:      genRandomID(),
		gb:      gb,
		Control: user,
	}

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
		return
	}
	fmt.Println("Received message from user:", string(rom.Rom))

	romPath, exists := roms[string(rom.Rom)]
	if !exists {
		user.WriteJSON(map[string]string{"error": "ROM not found"})
		fmt.Println("ROM not found for message:", string(rom.Rom))
		user.Close()
		return
	}

	romData, err := os.ReadFile(romPath)
	if err != nil {
		fmt.Println("Error loading ROM:", err)
		user.WriteJSON(map[string]string{"error": "Error loading ROM"})
		user.Close()
		return
	}

	gb.LOADROM(romData)

	user.WriteJSON(map[string]string{"success": "rom loaded"})

	go UpdateScreen(Game)
	go HandleControl(Game, endchan)
	go HandleGameEnd(store, Game, endchan)

	go gb.Start(nil, nil)

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

func UpdateScreen(g *Game) {
	for screen := range g.Viewer_chan {
		flatBytes := make([]byte, 0, 160*144)

		for y := range 144 {
			for x := range 160 {
				flatBytes = append(flatBytes, screen[x][y])
			}
		}

		activeViewers := g.Viewers[:0]
		for _, viewer := range g.Viewers {
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

func HandleControl(g *Game, endchan chan bool) {
	for {
		var update cpu.JoypadUpdate
		err := g.Control.ReadJSON(&update)
		if err != nil {
			fmt.Println("Error reading control update:", err)
			g.Control.Close()
			endchan <- true
			return
		}
		g.gb.ControlChan <- update
	}
}

func HandleGameEnd(store *GameStore, game *Game, endchan chan bool) {
	<-endchan
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
