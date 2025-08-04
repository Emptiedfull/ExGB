const canvas = document.getElementById('gameCanvas')
const gl = canvas.getContext('webgl2')


const vertexSrc = `#version 300 es
in vec2 a_position;
out vec2 v_texCoord;
void main() {
   v_texCoord = vec2(a_position.x * 0.5 + 0.5, 1.0 - (a_position.y * 0.5 + 0.5));
    gl_Position = vec4(a_position, 0, 1);
}`


const fragmentSrc = `#version 300 es
precision mediump float;
precision mediump sampler2D;
uniform sampler2D u_texture;
uniform vec3 u_palette[4];
in vec2 v_texCoord;
out vec4 outColor;

void main() {
    float x = v_texCoord.x * 160.0;
    float y = v_texCoord.y * 144.0;
    int pixelIndex = int(floor(y)) * 160 + int(floor(x));
    int byteIndex = pixelIndex / 4;
    int shift = 6 - 2 * (pixelIndex % 4);

    float fx = float(byteIndex) + 0.5;
    float fy = 0.5;
    vec4 texel = texelFetch(u_texture, ivec2(byteIndex, 0), 0);
    int packedVal = int(texel.r * 255.0 + 0.5);
    int colorIndex = (packedVal >> shift) & 0x3;
    outColor = vec4(u_palette[colorIndex], 1.0);
}`

function createShader(gl, type, source) {
    const shader = gl.createShader(type)
    gl.shaderSource(shader, source)
    gl.compileShader(shader)
    if (!gl.getShaderParameter(shader, gl.COMPILE_STATUS)) {
        throw new Error(gl.getShaderInfoLog(shader))
    }
    return shader
}
function createProgram(gl, vsSource, fsSource) {
    const program = gl.createProgram()
    gl.attachShader(program, createShader(gl, gl.VERTEX_SHADER, vsSource))
    gl.attachShader(program, createShader(gl, gl.FRAGMENT_SHADER, fsSource))
    gl.linkProgram(program)
    if (!gl.getProgramParameter(program, gl.LINK_STATUS)) {
        throw new Error(gl.getProgramInfoLog(program))
    }
    return program
}

const program = createProgram(gl, vertexSrc, fragmentSrc)
gl.useProgram(program)
const positionBuffer = gl.createBuffer()
gl.bindBuffer(gl.ARRAY_BUFFER, positionBuffer)
gl.bufferData(gl.ARRAY_BUFFER, new Float32Array([
    -1, -1,  1, -1,  -1, 1,
    -1,  1,  1, -1,   1, 1,
]), gl.STATIC_DRAW)
const a_position = gl.getAttribLocation(program, "a_position")
gl.enableVertexAttribArray(a_position)
gl.vertexAttribPointer(a_position, 2, gl.FLOAT, false, 0, 0)


const u_palette = gl.getUniformLocation(program, "u_palette")
gl.uniform3fv(u_palette, [
    155/255, 188/255, 15/255,
    139/255, 172/255, 15/255,
    48/255, 98/255, 48/255,
    15/255, 56/255, 15/255
])

function setPalette(paletteObj) {
    var flatPalette = [0, 1, 2, 3].flatMap(i => paletteObj[i]);
    gl.useProgram(program)
    gl.uniform3fv(u_palette, flatPalette.map(v => v / 255))
}



const texture = gl.createTexture()
gl.bindTexture(gl.TEXTURE_2D, texture)
gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

function renderFrame(packedBuffer) {
    gl.bindTexture(gl.TEXTURE_2D, texture)
    gl.texImage2D(
        gl.TEXTURE_2D, 0, gl.LUMINANCE,
        packedBuffer.length, 1, 0,
        gl.LUMINANCE, gl.UNSIGNED_BYTE, packedBuffer
    )
    gl.drawArrays(gl.TRIANGLES, 0, 6)
}