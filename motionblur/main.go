// Render a square with and without motion blur.
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

func run() error {
	err := glfw.Init(gl.ContextWatcher)
	if err != nil {
		return err
	}
	defer glfw.Terminate()

	var windowSize = [2]int{1024, 768}
	window, err := glfw.CreateWindow(windowSize[0], windowSize[1], "", nil, nil)
	if err != nil {
		return err
	}
	window.MakeContextCurrent()

	fmt.Printf("OpenGL: %s %s %s; %v samples.\n", gl.GetString(gl.VENDOR), gl.GetString(gl.RENDERER), gl.GetString(gl.VERSION), gl.GetInteger(gl.SAMPLES))
	fmt.Printf("GLSL: %s.\n", gl.GetString(gl.SHADING_LANGUAGE_VERSION))

	// Set callbacks.
	var cursorPos = [2]float32{float32(windowSize[0]) / 2, float32(windowSize[1]) / 2}
	var lastCursorPos = cursorPos
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
	gl.ClearColor(0, 0, 0, 1)
	gl.Enable(gl.CULL_FACE)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE)
	gl.Enable(gl.BLEND)

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

	tri0v0 := gl.GetUniformLocation(program, "tri0v0")
	tri0v1 := gl.GetUniformLocation(program, "tri0v1")
	tri0v2 := gl.GetUniformLocation(program, "tri0v2")
	tri1v0 := gl.GetUniformLocation(program, "tri1v0")
	tri1v1 := gl.GetUniformLocation(program, "tri1v1")
	tri1v2 := gl.GetUniformLocation(program, "tri1v2")

	vertexPositionAttrib := gl.GetAttribLocation(program, "aVertexPosition")
	gl.EnableVertexAttribArray(vertexPositionAttrib)

	triangleVertexPositionBuffer := gl.CreateBuffer()

	drawTriangle := func(triangle [9]float32, velocity mgl32.Vec3) {
		triangle0 := triangle
		for i := 0; i < 3*3; i++ {
			triangle0[i] -= velocity[i%3]
		}
		triangle1 := triangle

		gl.Uniform3f(tri0v0, triangle0[0], triangle0[1], triangle0[2])
		gl.Uniform3f(tri0v1, triangle0[3], triangle0[4], triangle0[5])
		gl.Uniform3f(tri0v2, triangle0[6], triangle0[7], triangle0[8])
		gl.Uniform3f(tri1v0, triangle1[0], triangle1[1], triangle1[2])
		gl.Uniform3f(tri1v1, triangle1[3], triangle1[4], triangle1[5])
		gl.Uniform3f(tri1v2, triangle1[6], triangle1[7], triangle1[8])

		{
			gl.BindBuffer(gl.ARRAY_BUFFER, triangleVertexPositionBuffer)
			vertices := f32.Bytes(binary.LittleEndian,
				triangle0[0], triangle0[1], triangle0[2],
				triangle0[3], triangle0[4], triangle0[5],
				triangle0[6], triangle0[7], triangle0[8],
				triangle1[0], triangle1[1], triangle1[2],
				triangle1[6], triangle1[7], triangle1[8],
				triangle1[3], triangle1[4], triangle1[5],
			)
			gl.BufferData(gl.ARRAY_BUFFER, vertices, gl.DYNAMIC_DRAW)
			itemSize := 3
			itemCount := 6

			gl.VertexAttribPointer(vertexPositionAttrib, itemSize, gl.FLOAT, false, 0, 0)
			gl.DrawArrays(gl.TRIANGLES, 0, itemCount)
		}

		{
			gl.BindBuffer(gl.ARRAY_BUFFER, triangleVertexPositionBuffer)
			vertices := f32.Bytes(binary.LittleEndian,
				triangle0[0], triangle0[1], triangle0[2],
				triangle1[0], triangle1[1], triangle1[2],
				triangle0[3], triangle0[4], triangle0[5],
				triangle1[3], triangle1[4], triangle1[5],
				triangle0[6], triangle0[7], triangle0[8],
				triangle1[6], triangle1[7], triangle1[8],
				triangle0[0], triangle0[1], triangle0[2],
				triangle1[0], triangle1[1], triangle1[2],
			)
			gl.BufferData(gl.ARRAY_BUFFER, vertices, gl.DYNAMIC_DRAW)
			itemSize := 3
			itemCount := 8

			gl.VertexAttribPointer(vertexPositionAttrib, itemSize, gl.FLOAT, false, 0, 0)
			gl.DrawArrays(gl.TRIANGLE_STRIP, 0, itemCount)
		}
	}

	if err := gl.GetError(); err != 0 {
		return fmt.Errorf("gl error: %v", err)
	}

	for !window.ShouldClose() {
		gl.Clear(gl.COLOR_BUFFER_BIT)

		pMatrix := mgl32.Ortho2D(0, float32(windowSize[0]), float32(windowSize[1]), 0)

		triangle0 := [9]float32{
			-50, -50, 0,
			50, -50, 0,
			-50, 50, 0}
		triangle1 := [9]float32{
			50, 50, 0,
			-50, 50, 0,
			50, -50, 0}

		// Square with motion blur on the left.
		{
			mvMatrix := mgl32.Translate3D(cursorPos[0]-200, cursorPos[1], 0)

			gl.UniformMatrix4fv(pMatrixUniform, pMatrix[:])
			gl.UniformMatrix4fv(mvMatrixUniform, mvMatrix[:])

			velocity := mgl32.Vec3{cursorPos[0] - lastCursorPos[0], cursorPos[1] - lastCursorPos[1], 0}

			drawTriangle(triangle0, velocity)
			drawTriangle(triangle1, velocity)
		}

		// Square without motion blur on the right.
		{
			mvMatrix := mgl32.Translate3D(cursorPos[0]+200, cursorPos[1], 0)

			gl.UniformMatrix4fv(pMatrixUniform, pMatrix[:])
			gl.UniformMatrix4fv(mvMatrixUniform, mvMatrix[:])

			drawTriangle(triangle0, mgl32.Vec3{})
			drawTriangle(triangle1, mgl32.Vec3{})
		}

		lastCursorPos = cursorPos

		window.SwapBuffers()
		glfw.PollEvents()
	}

	return nil
}

func main() {
	err := run()
	if err != nil {
		log.Fatalln(err)
	}
}
