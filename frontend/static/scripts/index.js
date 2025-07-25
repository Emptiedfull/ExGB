const { animate, scroll } = Motion

const pallete = {
    0:"#9bbc0f",
    1:"#8bac0f",
    2:"#306230",
    3:"#0f380f",
}

var GameId = null

document.addEventListener('DOMContentLoaded', async () => {

    gameCanvas = document.querySelector('.gameCanvas')

    sampledata = await fetch ('/static/data/sample.bin')
    const arrayBuffer = await sampledata.arrayBuffer()
    const data = new Uint8Array(arrayBuffer)

    console.log(data)
    UpdateCanvas(data)

    startNewGame()

})

const width = 160
const height = 144
const pixelSize = 2 

function startStream (){
    soc = new WebSocket(`ws://localhost:8080/ws/view/${GameId}`)

    soc.onopen = () => {
        console.log("WebSocket connection established")
    }

    soc.onmessage = async (event) => {
        console.log("Message received:", event.data)
         if (event.data instanceof Blob) {
            const arrayBuffer = await event.data.arrayBuffer()
            const byteArray = new Uint8Array(arrayBuffer)
            UpdateCanvas(byteArray)
        } else {
            // Handle text messages if needed
            console.log("Received text message:", event.data)
        }
    }
}

function startNewGame() {
    soc = new WebSocket(`ws://localhost:8080/ws/start`)

    soc.onopen = () => {
        console.log("WebSocket connection established")
    }

    soc.onmessage = (event) => {
        const message = JSON.parse(event.data)
        if (message.id){
            GameId = message.id
            startStream()
            console.log("Game ID received:", GameId)
        }
    }

    soc.onclose = function(event) {
        console.log('Game controller disconnected');
        setTimeout(() => {
            console.log('Attempting to reconnect...');
            startNewGame();
        }, 3000);
    };

    soc.onerror = (error) => {
        console.error("WebSocket error:", error)
    }
}


const UpdateCanvas = (data) => {
    gameCanvas = document.querySelector('.gameCanvas')
    ctx = gameCanvas.getContext('2d')
    


    gameCanvas.width = width * pixelSize
    gameCanvas.height = height * pixelSize
    
    let byteIndex = 0
    
    for (let y = 0; y < height; y++) {
        for (let x = 0; x < width; x++) {
            if (byteIndex < data.length) {
                // Get 2-bit pixel value (0, 1, 2, 3)
                const pixelValue = data[byteIndex] & 0x3
                
                // Set color from palette
                ctx.fillStyle = pallete[pixelValue]
                
                // Draw scaled pixel
                ctx.fillRect(x * pixelSize, y * pixelSize, pixelSize, pixelSize)
                
                byteIndex++
            }
        }
    }
}


