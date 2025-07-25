package main

import (
	"crypto/rand"
	"fmt"
	"image/color"
	"net/http"
	"os"

	"gbabot/cpu"

	"github.com/gorilla/websocket"
	"github.com/hajimehoshi/ebiten"
)

type Game struct {
	gb  *cpu.Gameboy
	Oam *cpu.OAMViewer

	Viewers     []*websocket.Conn
	Control     *websocket.Conn
	Viewer_chan chan [160][144]byte

	ID string
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (g *Game) Update(screen *ebiten.Image) error {
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Clear the screen first
	screen.Fill(color.RGBA{0, 0, 0, 255})

	// Draw the GameBoy screen
	g.gb.Draw(screen)
	g.Oam.DrawOAMTable(screen, 170, 10, 2)

}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 400, 300
}

func main() {

	Games := make(map[string]*Game)

	setUpHttpServer(Games)

	// game := &Game{gb: gb, Oam: Oam}
	// if err := ebiten.RunGame(game); err != nil {
	// 	fmt.Println("Error running game:", err)
	// }

	// Games = append(Games, game)

}

func setUpHttpServer(Games map[string]*Game) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to ExGB - GameBoy Emulator!"))
	})

	http.HandleFunc("/ws/start", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println("Error upgrading connection:", err)
			return
		}

		createGame(conn, Games)
	})

	http.HandleFunc("/ws/view/", func(w http.ResponseWriter, r *http.Request) {

		path := r.URL.Path
		id := path[len("/ws/view/"):]

		if id == "" || len(id) != 6 {
			http.Error(w, "Game ID is required", http.StatusBadRequest)
			fmt.Println("Invalid game ID:", id)
			return
		}

		Game := Games[id]
		if Game == nil {
			http.Error(w, "Game not found", http.StatusNotFound)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println("Error upgrading connection:", err)
			return
		}

		Game.Viewers = append(Game.Viewers, conn)
		fmt.Printf("New viewer connected to game %s, total viewers: %d\n", id, len(Game.Viewers))

	})

	http.ListenAndServe(":8080", nil)
}

func createGame(user *websocket.Conn, Games map[string]*Game) {
	ViewerChan := make(chan [160][144]byte, 10)
	gb := cpu.GBInitDebug(ViewerChan)
	Oam := &cpu.OAMViewer{Gb: gb, SpriteSize: 8}

	Game := &Game{
		ID:      genRandomID(),
		gb:      gb,
		Oam:     Oam,
		Control: user,
	}

	Game.Viewer_chan = ViewerChan

	go UpdateScreen(Game)

	// rompath := "./cpu/individual/acid.gb"
	rompath := "./tet.gb"
	romData, err := os.ReadFile(rompath)
	if err != nil {
		fmt.Println("Error loading ROM:", err)
		return
	}
	gb.LOADROM(romData)
	fmt.Println("New game created with ID:", Game.ID)

	user.WriteJSON(map[string]string{"id": Game.ID})

	go gb.Start(nil, nil)

	Games[Game.ID] = Game

}

func UpdateScreen(g *Game) {
	for screen := range g.Viewer_chan {
		flatBytes := make([]byte, 0, 160*144)

		for y := 0; y < 144; y++ {
			for x := 0; x < 160; x++ {
				flatBytes = append(flatBytes, screen[x][y])
			}
		}

		activeViewers := g.Viewers[:0]
		for _, viewer := range g.Viewers {
			err := viewer.WriteMessage(websocket.BinaryMessage, flatBytes)
			if err != nil {
				viewer.Close()
			} else {
				activeViewers = append(activeViewers, viewer)
			}
		}
		g.Viewers = activeViewers
	}
}

func HandleControl(g *Game) {
	for {
		_, message, err := g.Control.ReadMessage()
		if err != nil {
			fmt.Printf("Error reading control message: %v\n", err)
			continue
		}

		fmt.Println(message)
	}

}

func genRandomID() string {
	letters := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	result := make([]byte, 6)

	randomBytes := make([]byte, 6)
	rand.Read(randomBytes)

	for i, b := range randomBytes {
		result[i] = letters[b%byte(len(letters))]
	}

	return string(result)
}
