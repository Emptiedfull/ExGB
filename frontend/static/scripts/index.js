const { animate, scroll } = Motion

const pallete = {
    0: "#9bbc0f",
    1: "#8bac0f",
    2: "#306230",
    3: "#0f380f",
}

var modeDivs = {}

var GameId = null
var mode = "home"

var rompath = null

var controlSoc = null

document.addEventListener('DOMContentLoaded', async () => {

    modeDivs = {
        "home": document.querySelector('.home'),
        "game": document.querySelector('.game'),
        "rom-selection": document.querySelector('.rom-selection'),
        "play": document.querySelector('.play'),
        "link": document.querySelector('.link'),
    }

    menuControls = document.querySelectorAll('.menu-control')
    menuControls.forEach(control => {
        control.addEventListener('click', () => {


            const selectedMode = control.getAttribute('data-mode')
            if (selectedMode) {
                mode = selectedMode
                handleModeChange()
                console.log("Mode changed to:", mode)


                menuControls.forEach(c => {
                    c.classList.remove('active')
                })
                control.classList.add('active')
            } else {
                console.warn("No mode set for this control")
            }
        })
    })

    codeInput = document.getElementById('serverInput')
    connectButton = document.getElementById('connectButton')

    connectButton.addEventListener('click', () => {
        const serverCode = codeInput.value.trim()
        if (serverCode.length === 5) {
            mode = "spectator"
            GameId = serverCode
            handleModeChange()
            
        } else {
            console.error("Invalid server code. Please enter a 5-character code.")
        }
    })
})




function handleModeChange() {
    Object.values(modeDivs).forEach(div => {
        if (div) {
            div.style.display = "none"
            console.log("Hiding div:", div)
        }
    })

    

    switch (mode) {
        case "spectator":
              modeDivs.game.style.display = "flex"
              startStream()
              break
        case "play":
            if (modeDivs.play) {
                modeDivs.play.style.display = "flex"
                const playButton = document.getElementById('playButton')

                playButton.addEventListener('click', () => {
                    startNewGame()
                })
            }
            break
        case "link":
            if (modeDivs.link) {
                modeDivs.link.style.display = "flex"
            }
            break
        case "game":
            if (modeDivs.game) {
                modeDivs.game.style.display = "flex"
                handleControls(controlSoc)
            }
            break
        case "rom-selection":
            if (modeDivs['rom-selection']) {
                modeDivs['rom-selection'].style.display = "flex"
            }

            romlist = document.querySelector('.roms-list')
            if (romlist) {
                roms = romlist.querySelectorAll('.rom-item')
                roms.forEach(rom => {
                    rom.addEventListener('click', () => {


                        if (rom.classList.contains('active')) {
                            console.log("Starting game")
                            rompath = rom.getAttribute('data-rom')
                            mode = "game"
                            handleModeChange()
                            console.log(soc)
                            controlSoc.send(JSON.stringify({
                                "rom": rompath,
                            }))

                        } else {
                            roms.forEach(r => { r.classList.remove('active') })
                            rom.classList.add('active')

                        }


                    })
                })
            }
            break
    }

}

function setromactive(rom, roms) {
    roms.forEach(r => {
        r.classList.remove('active')
        if (r === rom) {
            if (!r.classList.contains('active')) {
                r.classList.add('active')
            }

        }
    })


}

const width = 160
const height = 144
const pixelSize = 2

function startStream() {
    soc = new WebSocket(`ws://localhost:8080/ws/view/${GameId}`)

    soc.onopen = () => {
        console.log("WebSocket connection established")


    }

    soc.onmessage = async (event) => {
        if (event.data instanceof Blob) {
            const arrayBuffer = await event.data.arrayBuffer()
            const byteArray = new Uint8Array(arrayBuffer)
            UpdateCanvas(byteArray)
        } else {
            console.log("Received text message:", event.data)
        }
    }
}

const controlMap = {
    "a": 0,        // A button
    "s": 1,        // B button
    "Escape": 2,   // Select
    "Enter": 3,    // Start
    "ArrowRight": 4,
    "ArrowLeft": 5,
    "ArrowUp": 6,
    "ArrowDown": 7,
}

function setGameId(id) {
    GameId = id
    const gameIdElement = document.getElementById('gameId')
    if (gameIdElement) {
        gameIdElement.textContent = GameId
    }
}

function handleControls(soc) {
    document.addEventListener('keydown', (event) => HandleKeyDown(event, soc))

    document.addEventListener('keyup', (event) => HandleKeyUp(event, soc))

}

function HandleKeyUp(event, soc) {
    if (event.key in controlMap) {
        key = controlMap[event.key]

        soc.send(JSON.stringify({
            "pressed": false,
            "key": key,
        }))
        console.log("Key released:", event.key, "Mapped to:", key)
    } else {
        console.log("Unknown key released:", event.key)
        return
    }
}

function HandleKeyDown(event, soc) {

    if (event.key in controlMap) {
        key = controlMap[event.key]

        soc.send(JSON.stringify({
            "pressed": true,
            "key": key,
        }))
        console.log("Key pressed:", event.key, "Mapped to:", key)
    } else {
        console.log("Unknown key pressed:", event.key)
        return
    }


}

function resetControlHandler() {
    document.removeEventListener('keydown', HandleKeyDown)
    document.removeEventListener('keyup', HandleKeyUp)
}



function startNewGame() {
    controlSoc = new WebSocket(`ws://localhost:8080/ws/start`)

    controlSoc.onopen = () => {
        console.log("WebSocket connection established")
        mode = "rom-selection"
        handleModeChange()
    }

    controlSoc.onmessage = (event) => {
        const message = JSON.parse(event.data)
        if (message.id) {
            setGameId(message.id)
            startStream()
            console.log("Game ID received:", GameId)
        }
    }

    controlSoc.onclose = function (event) {
        console.log('Game controller disconnected');
        resetControlHandler()
        setTimeout(() => {
            console.log('Attempting to reconnect...');
            startNewGame();
        }, 3000);
    };

    controlSoc.onerror = (error) => {
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


