self.onmessage = function(e) {
    const { data, width, height, pallete } = e.data
    const buf = new Uint8ClampedArray(width * height * 4)
    let idx = 0
    for (let y = 0; y < height; y++) {
        for (let x = 0; x < width; x += 4) {
            const b = data[idx / 4]
            for (let i = 0; i < 4; i++) {
                const pixelValue = (b >> (6 - 2 * i)) & 0x3
                const [r, g, b_] = pallete[pixelValue]
                const bufIdx = 4 * (y * width + x + i)
                buf[bufIdx] = r
                buf[bufIdx + 1] = g
                buf[bufIdx + 2] = b_
                buf[bufIdx + 3] = 255
            }
            idx += 4
        }
    }
    self.postMessage(buf, [buf.buffer])
}