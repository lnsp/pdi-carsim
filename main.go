package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/lnsp/pdi-carsim/geometry"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	windowWidth, windowHeight = 1200, 800
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

// CarModel is an abstract physics model of a car.
type CarModel struct {
	Size, Position, Velocity, Acceleration geometry.Vector
	Bounds                                 geometry.Polygon
	Mass, Rotation, Tension, Sensitivity   float64
}

// ApplyForce applies a force to the car.
func (car *CarModel) ApplyForce(f geometry.Vector) {
	car.Acceleration = car.Acceleration.Add(f.Scale(1.0 / car.Mass))
}

// Accelerate accelerates the car by the specified factor.
func (car *CarModel) Accelerate(delta float64) {
	car.ApplyForce(geometry.X.Scale(delta))
}

// Break stops the cars movement.
func (car *CarModel) Break(delta float64) {
	car.Acceleration = car.Acceleration.Scale(car.Tension / car.Sensitivity)
	car.Velocity = car.Velocity.Scale(car.Tension)
}

// Turn turns the wheel
func (car *CarModel) Turn(delta float64) {
	car.Rotation += delta * car.Sensitivity
}

// NewCar initializes a new car model.
func NewCar(mass, x, y, width, height float64) *CarModel {
	car := &CarModel{
		Size:         geometry.Vector{X: width, Y: height},
		Position:     geometry.Vector{X: x, Y: y},
		Velocity:     geometry.NullVector,
		Acceleration: geometry.NullVector,
		Mass:         mass,
		Rotation:     0,
		Tension:      0.9999,
		Sensitivity:  50.,
	}
	car.Bounds = geometry.NewPolygon(geometry.NullVector, geometry.NullVector.AddX(car.Size), car.Size, geometry.NullVector.AddY(car.Size))
	return car
}

func (car *CarModel) TurnCenter() geometry.Vector {
	return car.Bounds.Translate(geometry.Null.AddX(car.Size.Scale(-0.2))).Center()
}

// Update updates the car model.
func (car *CarModel) Update(delta float64) {
	car.Acceleration = car.Acceleration.Scale(car.Tension)
	car.Velocity = car.Velocity.Add(car.Acceleration.Scale(delta))
	car.Position = car.Position.Add(car.Velocity.Scale(delta).RotateAround(geometry.Null, car.Rotation))
}

// Draw renders the model onto the screen.
func (car *CarModel) Draw(r *sdl.Renderer) {
	// Rotated = Model.RotateAround(Center, Rotation)
	// Translated = Rotated.Translate(Car.Position)
	vertices := car.Bounds.RotateAround(car.TurnCenter(), car.Rotation).Translate(car.Position).Points()
	r.DrawLines(vertices)
}

type CarController interface {
	Feed(float64, geometry.Vector)
}

type SimpleCarControl struct {
	*CarModel
}

func (ctrl *SimpleCarControl) Feed(delta float64, p geometry.Vector) {
	diffVector := p.Add(ctrl.Position.Add(ctrl.TurnCenter()).Scale(-1))
	diffAngle := diffVector.AngleBetween(ctrl.Velocity.Norm().RotateAround(geometry.Null, ctrl.Rotation))
	fmt.Println(diffAngle)

	if diffAngle > 0 {
		ctrl.Turn(delta)
	} else if diffAngle < 0 {
		ctrl.Turn(-delta)
	}

	br := (-diffVector.Len() + 100)
	of := math.Log(diffVector.Len()-30) / 10

	if of > 0 {
		ctrl.Accelerate(0.1)
	}
	if br > 0 {
		ctrl.Break(br)
	}
}

func generateRandomPath(c int, x, y, width, height float64) geometry.Polygon {
	rand.Seed(time.Now().Unix())
	vertices := make([]geometry.Vector, c)
	for i := range vertices {
		vertices[i] = geometry.Vector{
			X: rand.Float64()*width + x,
			Y: rand.Float64()*height + y,
		}
	}
	return geometry.NewPolygon(vertices...)
}

func run() error {
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		return err
	}
	window, renderer, err := sdl.CreateWindowAndRenderer(windowWidth, windowHeight, sdl.WINDOW_SHOWN)
	if err != nil {
		return err
	}
	defer window.Destroy()
	defer renderer.Destroy()

	targetPath := generateRandomPath(8, 100, 100, 1000, 600)
	car := NewCar(1, 100, 100, 100, 50)
	path := []sdl.Point{car.TurnCenter().ToPoint()}
	lastFrame, lastPathUpdate := time.Now(), time.Now()
	ctrl := SimpleCarControl{car}

	progress := 2.0
	ownControl := false
	for {
		renderer.SetDrawColor(0, 0, 0, 255)
		renderer.Clear()
		renderer.SetDrawColor(0, 0, 255, 255)
		renderer.DrawLines(targetPath.Points())
		renderer.SetDrawColor(0, 255, 0, 255)
		renderer.DrawLines(path)
		renderer.SetDrawColor(255, 0, 0, 255)
		car.Draw(renderer)

		delta := float64(time.Since(lastFrame)) / float64(time.Second)
		car.Update(delta)
		lastFrame = time.Now()

		if time.Since(lastPathUpdate) > time.Second/10 {
			path = append(path, car.Position.Add(car.TurnCenter()).ToPoint())
			lastPathUpdate = time.Now()
		}

		progress += delta / 60
		if progress > 1.0 {
			targetPath = generateRandomPath(8, 100, 100, 1000, 600)
			progress = 0.0
			car.Position = targetPath.Interpolate(0.0).Add(car.TurnCenter().Scale(-1))
			path = []sdl.Point{car.Position.ToPoint()}
		}
		interpol := targetPath.Interpolate(progress).ToPoint()
		renderer.DrawPoints([]sdl.Point{interpol})

		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch et := event.(type) {
			case *sdl.KeyDownEvent:
				switch et.Keysym.Sym {
				case sdl.K_ESCAPE:
					return nil
				case sdl.K_w:
					car.Accelerate(10.0)
				case sdl.K_s:
					car.Break(10.0)
				case sdl.K_a:
					car.Turn(math.Pi * 4 * delta)
				case sdl.K_d:
					car.Turn(-math.Pi * 4 * delta)
				case sdl.K_o:
					ownControl = !ownControl
				}
			}
		}
		if !ownControl {
			ctrl.Feed(delta, targetPath.Interpolate(progress))
		}

		renderer.Present()
	}
}
