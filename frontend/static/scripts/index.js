const { animate, scroll } = Motion

var pallete = {
    0: [155, 188, 15],   
    1: [139, 172, 15],  
    2: [48, 98, 48],     
    3: [15, 56, 15],     
}

var modeDivs = {}

var GameId = null
var mode = "home"

var Gameon = false

var rompath = null

var controlSoc = null

var public = true


var gameCanvas = null
var ctx = null
var Imagedata = null
var buf = null

const worker = new Worker('static/scripts/pixelworker.js')

function uint8ToBase64(buffer) {
    const uint8Array = buffer instanceof Uint8Array ? buffer : new Uint8Array(buffer);
    let binary = '';
    const chunkSize = 0x8000; 
    for (let i = 0; i < uint8Array.length; i += chunkSize) {
        binary += String.fromCharCode.apply(
            null,
            uint8Array.subarray(i, i + chunkSize)
        );
    }
    return btoa(binary);
}

document.addEventListener('DOMContentLoaded', async () => {

    modeDivs = {
        "home": document.querySelector('.home'),
        "game": document.querySelector('.game'),
        "rom-selection": document.querySelector('.rom-selection'),
        "play": document.querySelector('.play'),
        "link": document.querySelector('.link'),
        "gear": document.querySelector('.gear'),
    }

    gameCanvas = document.querySelector('.gameCanvas')

    loadbutton = document.querySelector('.rom-load')
            loadbutton.addEventListener('click', () => {
                const input = document.createElement('input')
                input.type = 'file'
                input.style.display = 'none'
                input.accept = '.gb,.gbc,.bin'

                input.onchange = (event) => {
                    const file = event.target.files[0]
                    if (file){
                        const reader = new FileReader()
                        reader.onload = (e) => {
                            const romData = e.target.result
                            console.log(romData.byteLength)
                            const base64Data = uint8ToBase64(romData);
                             console.log("Base64 length:", base64Data.length);
                            controlSoc.send(JSON.stringify({
                                "rom": "load",
                                "state": base64Data,
                            }))
                            console.log("ROM loaded from file")
                            mode = "game"
                            handleModeChange()
                        }
                        reader.readAsArrayBuffer(file)
                    }
                }

                document.body.appendChild(input)
                input.click()
                document.body.removeChild(input)

                
                
            })

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


})




function handleModeChange() {
    Object.values(modeDivs).forEach(div => {
        if (div) {
            div.style.display = "none"
            console.log("Hiding div:", div)
        }
    })


    switch (mode) {
        case "gear":
            if (modeDivs.gear) {
                modeDivs.gear.style.display = "flex"
            }

            fetchSettings()
            setUpGearHandlers()

            break
        case "home":
            if (modeDivs.home) {
                modeDivs.home.style.display = "flex"
            }

            break
        case "spectator":
            modeDivs.game.style.display = "flex"
            startStream()
            break
        case "play":
            if (Gameon){
                mode = "game"
                handleModeChange()
                return
            }
            
            if (modeDivs.play) {
                modeDivs.play.style.display = "flex"
                const playButton = document.getElementById('playButton')

                playButton.addEventListener('click', () => {
                    startNewGame()
                })

                serverCountElement = document.getElementById('serverCount')
                serverErrorElement = modeDivs.play.querySelector('.play-error')

                fetch('https://ex-gb.com/api/servers/count')
                    .then(response => response.text())
                    .then(count => {
                        serverCountElement.textContent = count
                        if (parseInt(count) > 0) {
                            modeDivs.play.querySelector('.play-error').style.display = 'none'
                        } else {
                            serverErrorElement.textContent = "All Servers Busy"
                            modeDivs.play.querySelector('.play-error').style.display = 'block'
                        }
                    })
                    .catch(error => {
                        console.error("Error fetching server count:", error)
                        serverCountElement.textContent = "0"
                        serverErrorElement.textContent = "Error connecting to server, Pls inform admin"
                        modeDivs.play.querySelector('.play-error').style.display = 'block'
                    }
                    )
            }
            break
        case "link":
            if (modeDivs.link) {
                modeDivs.link.style.display = "flex"

                console.log("Link mode activated")

                codeInput = document.getElementById('serverInput')
                connectButton = document.getElementById('connectButton')

                connectButton.addEventListener('click', () => {
                    const serverCode = codeInput.value.trim()
                    if (serverCode.length === 5) {
                        mode = "spectator"
                        GameId = serverCode

                        console.log("Connecting to server with code:", GameId)

                        fetch(`https://ex-gb.com/api/servers/status?id=${GameId}`).then(response => {
                            if (!response.ok) {
                                linkError = modeDivs.link.querySelector('.link-error')
                                linkError.textContent = "Invalid server code"

                            } else {
                                handleModeChange()
                            }
                        }
                        ).catch(error => {
                            console.error("Error connecting to server:", error)
                            linkError = modeDivs.link.querySelector('.link-error')
                            linkError.textContent = "Error connecting to server, Pls inform admin"
                        })


                    } else {
                        console.error("Invalid server code. Please enter a 5-character code.")
                    }
                })
            }


            break
        case "game":
            if (modeDivs.game) {
                modeDivs.game.style.display = "flex"
                handleControls(controlSoc)
            }

            endButton = document.getElementById('exitGameButton')
            endButton.addEventListener('click', () => {
                controlSoc.close()
                resetControlHandler()
                mode = "play"
                Gameon = false
                handleModeChange()
            })

            saveButton = document.getElementById('saveStateButton')
            saveButton.addEventListener('click', () => {
                controlSoc.send(JSON.stringify({
                    "meta":"savestate",
                    "saveState": true,
                }))
                console.log("Save game request sent")
            })

            pauseButton = document.getElementById('pauseButton')
            pauseButton.addEventListener('click', () => {
                if (pauseButton.textContent === "PAUSE"){
                    controlSoc.send(JSON.stringify({
                        "meta":"pause",
                        "pause": true,
                    }))
                    pauseButton.textContent = "RESUME"
                    console.log("Game paused")
                }else{
                    controlSoc.send(JSON.stringify({
                        "meta":"pause",
                        "pause": false,
                    }))
                    pauseButton.textContent = "PAUSE"
                    console.log("Game resumed")
                }
            })

            break
        case "rom-selection":
            if (modeDivs['rom-selection']) {
                modeDivs['rom-selection'].style.display = "flex"
            }

            startbutton = document.querySelector('.rom-start')
            startbutton.addEventListener('click', () => {
                if (rompath) {
                    console.log("Starting game with ROM:", rompath)
                    mode = "game"
                    handleModeChange()
                    controlSoc.send(JSON.stringify({
                        "rom": rompath,
                    }))
                } else {
                    console.warn("No ROM selected")
                }
            })

            

            romlist = document.querySelector('.roms-list')
            if (romlist) {
                roms = romlist.querySelectorAll('.rom-item')
                roms.forEach(rom => {
                    rom.addEventListener('click', () => {


                      
                            rompath = rom.getAttribute('data-rom')
                            roms.forEach(r => { r.classList.remove('active') })
                            rom.classList.add('active')

                            romName = document.getElementById('romName')
                            if (romName) {
                                romName.textContent = rom.getAttribute('data-rom')
                            }

                            controlButtons = document.querySelector('.rom-controls')
                            if (controlButtons) {
                                controlButtons.querySelector('.rom-start').disabled = false
                                controlButtons.querySelector('.rom-load').disabled = false
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
    soc = new WebSocket(`wss://ex-gb.com/api/ws/view/${GameId}`)

    soc.onopen = () => {
        console.log("WebSocket connection established")
        Gameon = true
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
        
    } else {
        
        return
    }


}

function resetControlHandler() {
    document.removeEventListener('keydown', HandleKeyDown)
    document.removeEventListener('keyup', HandleKeyUp)
}



function startNewGame() {
    controlSoc = new WebSocket(`wss://ex-gb.com/api/ws/start?public=${public}`)

    controlSoc.onopen = () => {
        console.log("WebSocket connection established")
        mode = "rom-selection"
        handleModeChange()
    }

    controlSoc.onmessage = (event) => {
        

        if (event.data instanceof Blob){
            console.log("Received binary data, saving ROM...")
            const blob = new Blob([event.data], { type: 'application/octet-stream' })
            const url = URL.createObjectURL(blob)

            const now = new Date()
           const timeString = now.toISOString().replace(/[:.]/g, '-')

            const a = document.createElement('a')
            a.href = url
            a.download = rompath + timeString + ".bin"
            document.body.appendChild(a)
            a.click()
            document.body.removeChild(a)
            URL.revokeObjectURL(url)
            return
        }
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
    };

    controlSoc.onerror = (error) => {
        console.error("WebSocket error:", error)
    }
}

function UpdateCanvas(data) {
     renderFrame(data);

}

const fetchSettings = () => {
    currentPalette = document.querySelector('.current-pallete')
    currentBlocks = currentPalette.querySelectorAll('span')
    currentBlocks.forEach((block, index) => {
        block.style.backgroundColor = `rgb(${pallete[index].join(',')})`
    })
}

const palletes = [{
    0: [155, 188, 15],   
    1: [139, 172, 15],  
    2: [48, 98, 48],     
    3: [15, 56, 15],     
},{
    0: [113 ,49 ,65],   
    1: [81, 40 ,57],  
    2: [49 ,33 ,55],     
    3: [26 ,33 ,41],     
},{
    0: [155, 188, 15],   
    1: [139, 172, 15],  
    2: [48, 98, 48],     
    3: [15, 56, 15],     
}]

const Modes = ["PUBLIC", "PRIVATE"]


const setUpGearHandlers = () =>{
    console.log("Setting up gear handlers")
    const palleteLeft = document.querySelector('.pallete-left')
    const palleteRight = document.querySelector('.pallete-right')
    const currentPalette = document.querySelector('.current-pallete')


    palleteLeft.addEventListener('click', () => {
        currentIndex = currentPalette.getAttribute('data-id')
        let newIndex = parseInt(currentIndex) - 1
        if (newIndex < 0) newIndex = palletes.length - 1
        currentPalette.setAttribute('data-id', newIndex)
        pallete = palletes[newIndex]
        setPalette(pallete)
        fetchSettings()
    })

    palleteRight.addEventListener('click', () => {
        currentIndex = currentPalette.getAttribute('data-id')
        let newIndex = parseInt(currentIndex) + 1
        if (newIndex >= palletes.length) newIndex = 0
        pallete = palletes[newIndex]
        setPalette(pallete)
        currentPalette.setAttribute('data-id', newIndex)
        fetchSettings()
    })

    const modeLeft = document.querySelector('.mode-left')
    const modeRight = document.querySelector('.mode-right')
    const currentMode = document.querySelector('.current-mode')

    modeLeft.addEventListener('click', () => {
        currentIndex = currentMode.getAttribute('data-id')
        let newIndex = parseInt(currentIndex) - 1
        if (newIndex < 0) newIndex = Modes.length - 1
        currentMode.setAttribute('data-id', newIndex)
        public = (newIndex === 0)
        currentMode.textContent = Modes[newIndex]
    })

    modeRight.addEventListener('click', () => {
        currentIndex = currentMode.getAttribute('data-id')
        let newIndex = parseInt(currentIndex) + 1
        if (newIndex >= Modes.length) newIndex = 0
        public = (newIndex === 0)
        currentMode.setAttribute('data-id', newIndex)
        currentMode.textContent = Modes[newIndex]
    })

}