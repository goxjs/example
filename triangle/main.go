// Render a basic triangle.
package main

import (
	"encoding/binary"
	"fmt"
	"log"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/goxjs/gl"
	"github.com/goxjs/gl/glutil"
	"github.com/goxjs/glfw"
	"golang.org/x/mobile/exp/f32"
)

func main() {
	err := run()
	if err != nil {
		log.Fatalln(err)
	}
}

func run() error {
	err := glfw.Init(gl.ContextWatcher)
	if err != nil {
		return err
	}
	defer glfw.Terminate()

	var windowSize = [2]int{640, 480}
	glfw.WindowHint(glfw.Samples, 8) // Anti-aliasing.
	window, err := glfw.CreateWindow(windowSize[0], windowSize[1], "", nil, nil)
	if err != nil {
		return err
	}
	window.MakeContextCurrent()

	fmt.Printf("OpenGL: %s %s %s; %v samples.\n", gl.GetString(gl.VENDOR), gl.GetString(gl.RENDERER), gl.GetString(gl.VERSION), gl.GetInteger(gl.SAMPLES))
	fmt.Printf("GLSL: %s.\n", gl.GetString(gl.SHADING_LANGUAGE_VERSION))

	// Set callbacks.
	var cursorPos = [2]float32{200, 200}
	cursorPosCallback := func(_ *glfw.Window, x, y float64) {
		cursorPos[0], cursorPos[1] = float32(x), float32(y)
	}
	window.SetCursorPosCallback(cursorPosCallback)

	framebufferSizeCallback := func(w *glfw.Window, framebufferSize0, framebufferSize1 int) {
		gl.Viewport(0, 0, framebufferSize0, framebufferSize1)

		windowSize[0], windowSize[1] = w.GetSize()
	}
	window.SetFramebufferSizeCallback(framebufferSizeCallback)
	{
		var framebufferSize [2]int
		framebufferSize[0], framebufferSize[1] = window.GetFramebufferSize()
		framebufferSizeCallback(window, framebufferSize[0], framebufferSize[1])
	}

	// Set OpenGL options.
	gl.ClearColor(0.8, 0.3, 0.01, 1)

	// Init shaders.
	program, err := glutil.CreateProgram(vertexSource, fragmentSource)
	if err != nil {
		return err
	}

	gl.ValidateProgram(program)
	if gl.GetProgrami(program, gl.VALIDATE_STATUS) != gl.TRUE {
		return fmt.Errorf("gl validate status: %s", gl.GetProgramInfoLog(program))
	}

	gl.UseProgram(program)

	pMatrixUniform := gl.GetUniformLocation(program, "uPMatrix")
	mvMatrixUniform := gl.GetUniformLocation(program, "uMVMatrix")

	// Create VBO.
	triangleVertexPositionBuffer := gl.CreateBuffer()
	gl.BindBuffer(gl.ARRAY_BUFFER, triangleVertexPositionBuffer)
	vertices := f32.Bytes(binary.LittleEndian,
		0, 0, 0,
		300, 100, 0,
		0, 100, 0,
	)
	gl.BufferData(gl.ARRAY_BUFFER, vertices, gl.STATIC_DRAW)
	itemSize := 3
	itemCount := 3

	vertexPositionAttrib := gl.GetAttribLocation(program, "aVertexPosition")
	gl.EnableVertexAttribArray(vertexPositionAttrib)
	gl.VertexAttribPointer(vertexPositionAttrib, itemSize, gl.FLOAT, false, 0, 0)

	if err := gl.GetError(); err != 0 {
		return fmt.Errorf("gl error: %v", err)
	}

	for !window.ShouldClose() {
		gl.Clear(gl.COLOR_BUFFER_BIT)

		pMatrix := mgl32.Ortho2D(0, float32(windowSize[0]), float32(windowSize[1]), 0)

		mvMatrix := mgl32.Translate3D(cursorPos[0], cursorPos[1], 0)

		gl.UniformMatrix4fv(pMatrixUniform, pMatrix[:])
		gl.UniformMatrix4fv(mvMatrixUniform, mvMatrix[:])
		gl.DrawArrays(gl.TRIANGLES, 0, itemCount)

		window.SwapBuffers()
		glfw.PollEvents()
	}

	return nil
}

const (
	vertexSource = `//#version 120 // OpenGL 2.1.
//#version 100 // WebGL.

attribute vec3 aVertexPosition;

uniform mat4 uMVMatrix;
uniform mat4 uPMatrix;

void main() {
	gl_Position = uPMatrix * uMVMatrix * vec4(aVertexPosition, 1.0);
}
`
	fragmentSource = `//#version 120 // OpenGL 2.1.
//#version 100 // WebGL.

void main() {
	gl_FragColor = vec4(1.0, 1.0, 1.0, 1.0);
}
`
)
