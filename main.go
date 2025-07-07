package main

import (
	"fmt"
	"net/http"
	"os"

	"gbabot/cpu"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

func main() {
	gb := cpu.GBInitDebug()
	rompath := "./cpu/individual/boot.gb"
	romData, err := os.ReadFile(rompath)
	if err != nil {
		fmt.Println("Error loading ROM:", err)
		return
	}
	gb.LOADROM(romData)
	fmt.Printf("Game Boy initialized: %+v\n", gb != nil)

	go gb.Start(nil) // Pass nil for the done channel in this example

	http.HandleFunc("/", serveIndex)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		HandleWebSocket(w, r, gb)
	})

	fmt.Println("Starting server on :8080")
	http.ListenAndServe(":8080", nil)
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request, g *cpu.Gameboy) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading connection:", err)
		return
	}

	g.Soc = conn

	for {

		_, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading message:", err)
			break
		}

		fmt.Printf("Received message: %s\n", msg)
	}
	defer conn.Close()

}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>Game Boy Emulator</title>
    <style>
        body { margin: 0; padding: 20px; background: #000; color: #fff; font-family: monospace; }
        canvas { border: 1px solid #fff; image-rendering: pixelated; }
        #gameboy { text-align: center; }
    </style>
</head>
<body>
    <div id="gameboy">
        <h1>Game Boy Emulator</h1>
        <canvas id="screen" width="160" height="144"></canvas>
    </div>
    
    <script>
        const canvas = document.getElementById('screen');
        const ctx = canvas.getContext('2d');
        
        // Scale up the canvas for better visibility
        canvas.style.width = '640px';
        canvas.style.height = '576px';
        
        const ws = new WebSocket('ws://localhost:8080/ws');
        
        ws.onopen = function(event) {
            console.log('Connected to Game Boy emulator');
        };
        
        ws.onmessage = function(event) {
            const data = JSON.parse(event.data);
			console.log(data)
            
            // Create ImageData from pixel buffer
            const imageData = ctx.createImageData(data.width, data.height);
            
            // Convert RGB to RGBA
            for (let i = 0; i < data.pixels.length; i += 3) {
                const pixelIndex = (i / 3) * 4;
                imageData.data[pixelIndex] = data.pixels[i];     // R
                imageData.data[pixelIndex + 1] = data.pixels[i + 1]; // G
                imageData.data[pixelIndex + 2] = data.pixels[i + 2]; // B
                imageData.data[pixelIndex + 3] = 255; // A (fully opaque)
            }
            
            ctx.putImageData(imageData, 0, 0);
        };
        
        ws.onclose = function(event) {
            console.log('Disconnected from Game Boy emulator');
        };
        
        ws.onerror = function(error) {
            console.error('WebSocket error:', error);
        };
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}
