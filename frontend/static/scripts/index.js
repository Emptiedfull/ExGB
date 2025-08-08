const { animate, scroll } = Motion

var pallete = {
    0: [155, 188, 15],
    1: [139, 172, 15],
    2: [48, 98, 48],
    3: [15, 56, 15],
}

let wasmReady = false
let wasmModule = null

let currentSpeed = 0

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

async function initWASM() {
    try {
        const go = new Go()
        const result = await WebAssembly.instantiateStreaming(fetch('static/scripts/main.wasm?v=233212'), go.importObject)
        wasmModule = result.instance
        go.run(wasmModule)
        wasmReady = true
        console.log("WASM module initialized")
    } catch (error) {
        console.error("Error initializing WASM module:", error)
    }
}
document.addEventListener('DOMContentLoaded', async () => {

    await initWASM()

    window.onFrameUpdate = (data) => {
        renderFrame(data)
    }




    keySelects = document.querySelectorAll('.currentKey')
    keySelects.forEach(select => {

        select.addEventListener('click', () => {
            resetControlHandler()
            const currentKey = select.getAttribute('id')
            select.textContent = `${currentKey}: ...`

            function keydownHandler(event) {
                select.textContent = `${currentKey}: ${event.key}`
                select.setAttribute('data-key', event.key)
                if (event.key in keyMap) {
                    delete keyMap[event.key]
                }
                keyMap[event.key] = currentKey
                document.removeEventListener('keydown', keydownHandler) 
                handleControls()
            }
            document.addEventListener('keydown', keydownHandler)
        })
    })



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
            if (file) {
                const reader = new FileReader()
                reader.onload = (e) => {
                    const arrayBuffer = e.target.result
                    const byteArray = new Uint8Array(arrayBuffer)
                    const romData = new Uint8ClampedArray(byteArray)

                    res = window.loadState(romData)
                    console.log("ROM loaded with response:", res)

                    res = window.startGame()
                    currentSpeed = document.querySelector('.current-speed')
                    if (currentSpeed) {
                        window.changeSpeed(parseInt(currentSpeed.textContent))
                    }
                    mode = "game"
                    Gameon = true
                    handleModeChange()
                    console.log("Game started with response:", res)
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
            if (Gameon) {
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

                        fetch(`http://localhost:8080/servers/status?id=${GameId}`).then(response => {
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
                handleControls()
            }

            endButton = document.getElementById('exitGameButton')
            endButton.addEventListener('click', () => {
                window.endGame()
                mode = "home"
                Gameon = false
                handleModeChange()
            })

            saveButton = document.getElementById('saveStateButton')
            saveButton.addEventListener('click', () => {
                state = window.saveState()
                if (state) {
                    const blob = new Blob([state], { type: 'application/octet-stream' })
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
                }
            })

            pauseButton = document.getElementById('pauseButton')
            pauseButton.addEventListener('click', () => {
                if (pauseButton.textContent === "Pause") {
                    pauseButton.textContent = "Resume"
                    window.pauseGame()
                } else {
                    pauseButton.textContent = "Pause"
                    window.resumeGame()
                }
            })

            loadstateButton = document.getElementById('loadStateButton')
            loadstateButton.addEventListener('click', () => {
                const input = document.createElement('input')
                input.type = 'file'
                input.style.display = 'none'
                input.accept = '.gb,.gbc,.bin'

                input.onchange = (event) => {
                    const file = event.target.files[0]
                    window.endGame()
                    window.initGameboy()
                    if (file) {
                        const reader = new FileReader()
                        reader.onload = (e) => {
                            const arrayBuffer = e.target.result
                            const byteArray = new Uint8Array(arrayBuffer)
                            const romData = new Uint8ClampedArray(byteArray)

                            res = window.loadState(romData)
                            console.log("ROM loaded with response:", res)

                            res = window.startGame()
                            currentSpeed = document.querySelector('.current-speed')
                    if (currentSpeed) {
                        window.changeSpeed(parseInt(currentSpeed.textContent))
                    }
                            mode = "game"
                            Gameon = true
                            handleModeChange()
                            console.log("Game started with response:", res)
                        }
                        reader.readAsArrayBuffer(file)
                    }
                }

                document.body.appendChild(input)
                input.click()
                document.body.removeChild(input)
            })

            break
        case "rom-selection":
            if (modeDivs['rom-selection']) {
                modeDivs['rom-selection'].style.display = "flex"
            }

            startbutton = document.querySelector('.rom-start')
            startbutton.addEventListener('click', () => {
                fetch(`static/roms/${rompath}.gb`).then(response => {
                    if (response.ok) {
                        response.arrayBuffer().then(buffer => {
                            const byteArray = new Uint8Array(buffer)
                            const romData = new Uint8ClampedArray(byteArray)
                            window.loadRom(romData)

                            res = window.startGame()
                            currentSpeed = document.querySelector('.current-speed')
                    if (currentSpeed) {
                        window.changeSpeed(parseInt(currentSpeed.textContent))
                    }
                            mode = "game"
                            Gameon = true
                            handleModeChange()
                            console.log("Game started with response:", res)
                        })
                    } else {
                        console.error("Failed to load ROM")
                    }
                })

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
    mode = "rom-selection"
    handleModeChange()
}

const controlMap = {
    "A": 0,        // A button
    "B": 1,        // B button
    "SELECT": 2,   // Select
    "START": 3,    // Start
    "RIGHT": 4,
    "LEFT": 5,
    "UP": 6,
    "DOWN": 7,
}

const keyMap = {
    "a": "A",
    "s": "B",
    "Escape": "SELECT",
    "Enter": "START",
    "ArrowUp": "UP",
    "ArrowDown": "DOWN",
    "ArrowLeft": "LEFT",
    "ArrowRight": "RIGHT",
}

function setGameId(id) {
    GameId = id
    const gameIdElement = document.getElementById('gameId')
    if (gameIdElement) {
        gameIdElement.textContent = GameId
    }
}

function handleControls() {
    document.addEventListener('keydown', (event) => HandleKeyDown(event))

    document.addEventListener('keyup', (event) => HandleKeyUp(event))

}

function HandleKeyUp(event) {
    if (event.key in keyMap) {
        key = controlMap[keyMap[event.key]]

        window.sendInput(false, key)

    } else {
        console.log("Unknown key released:", event.key)
        return
    }
}

function HandleKeyDown(event) {

    if (event.key in keyMap) {
        key = controlMap[keyMap[event.key]]

        window.sendInput(true, key)

    } else {

        return
    }


}

function resetControlHandler() {
    document.removeEventListener('keydown', HandleKeyDown)
    document.removeEventListener('keyup', HandleKeyUp)
}



function startNewGame() {
    result = window.initGameboy()
    console.log("Gameboy initialized:", result)

    setGameId("hi")
    startStream()
    console.log("Game ID received:", GameId)
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
}, {
    0: [113, 49, 65],
    1: [81, 40, 57],
    2: [49, 33, 55],
    3: [26, 33, 41],
}, {
    0: [207, 146, 85],
    1: [207, 113, 99],
    2: [176, 21, 83],
    3: [63, 23, 17],
}, {
    0: [169, 176, 179],
    1: [88, 97, 100],
    2: [32, 41, 63],
    3: [3, 12, 34],
}, {
    0: [158, 251, 227],
    1: [33, 175, 245],
    2: [30, 71, 147],
    3: [14, 31, 61]
}]

const Speeds = [1, 2,3, 4,5, 6,7, 8]
const AnimationModes = ["LOW", "HIGH"]

const setUpGearHandlers = () => {
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

    const speedLeft = document.querySelector('.speed-left')
    const speedRight = document.querySelector('.speed-right')
    const currentSpeed = document.querySelector('.current-speed')

    speedLeft.addEventListener('click', () => {
        currentIndex = currentSpeed.getAttribute('data-id')
        let newIndex = parseInt(currentIndex) - 1
        if (newIndex < 0) newIndex = Speeds.length - 1
        currentSpeed.setAttribute('data-id', newIndex)
        currentSpeed.textContent = Speeds[newIndex]
        // if(window.checkGameState()){
        //     window.changeSpeed(Speeds[newIndex])
        // }
        
    })

    speedRight.addEventListener('click', () => {
        currentIndex = currentSpeed.getAttribute('data-id')
        let newIndex = parseInt(currentIndex) + 1
        if (newIndex >= Speeds.length) newIndex = 0
        currentSpeed.setAttribute('data-id', newIndex)
        currentSpeed.textContent = Speeds[newIndex]
        // if(window.checkGameState()){
        //     window.changeSpeed(Speeds[newIndex])

        // }
    })



    

    const animationLeft = document.querySelector('.effect-left')
    const animationRight = document.querySelector('.effect-right')
    const currentAnimation = document.querySelector('.current-effect')
    console.log("Current animation mode:", currentAnimation.getAttribute('data-id'))

    animationLeft.addEventListener('click', () => {
        currentIndex = currentAnimation.getAttribute('data-id')
        let newIndex = parseInt(currentIndex) - 1
        if (newIndex < 0) newIndex = AnimationModes.length - 1
        currentAnimation.setAttribute('data-id', newIndex)
        console.log("Animation mode changed to:", AnimationModes[newIndex], currentIndex, newIndex)
        currentAnimation.textContent = AnimationModes[newIndex]
        if (newIndex === 1) {
            element.style.animationPlayState = 'running'
        } else {
            element.style.animationPlayState = 'paused'
        }
    })

    animationRight.addEventListener('click', () => {
        currentIndex = currentAnimation.getAttribute('data-id')
        let newIndex = parseInt(currentIndex) + 1
        if (newIndex >= AnimationModes.length) newIndex = 0
        currentAnimation.setAttribute('data-id', newIndex)
        console.log("Animation mode changed to:", AnimationModes[newIndex], newIndex)
        currentAnimation.textContent = AnimationModes[newIndex]
        if (newIndex === 1) {
            element.style.animationPlayState = 'running'
        } else {
            element.style.animationPlayState = 'paused'
        }
    })


    const feedbackButton = document.querySelector('.feedback-button')
    feedbackButton.addEventListener('click', () => {
        const textArea = document.getElementById("feedbackTextarea")
        const feedback = textArea.value.trim()
        if (feedback) {
            console.log("Feedback submitted:", feedback)
            textArea.value = ""
        }
    })

}