package main

import (
	"fmt"
	"math"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

const (
	windowWidth, windowHeight = 1280, 800
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

var Null Vector = Vector{0, 0}

type Vector struct {
	X, Y float64
}

func (v Vector) Add(a Vector) Vector {
	return Vector{v.X + a.X, v.Y + a.Y}
}

func (v Vector) AddX(a Vector) Vector {
	return Vector{v.X + a.X, v.Y}
}

func (v Vector) AddY(a Vector) Vector {
	return Vector{v.X, v.Y + a.Y}
}

func (v Vector) Scale(s float64) Vector {
	return Vector{s * v.X, s * v.Y}
}

func (v Vector) Norm() Vector {
	len := v.Len()
	if len != 0.0 {
		return v.Scale(1.0 / len)
	}
	return v
}

func (v Vector) Len() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

func (v Vector) ToPoint() sdl.Point {
	return sdl.Point{X: int32(v.X), Y: int32(v.Y)}
}

func (v Vector) Dot(a Vector) float64 {
	return v.X*a.X + v.Y*a.Y
}

func (v Vector) Det(a Vector) float64 {
	return v.X*a.Y - v.Y*a.X
}

func (v Vector) AngleBetween(a Vector) float64 {
	return math.Atan2(v.Det(a), v.Dot(a))
}

func (v Vector) RotateAround(a Vector, angle float64) Vector {
	return Vector{
		X: ((v.X - a.X) * math.Cos(angle)) + ((a.Y - v.Y) * math.Sin(angle)) + a.X,
		Y: ((a.Y - v.Y) * math.Cos(angle)) - ((v.X - a.X) * math.Sin(angle)) + a.Y,
	}
}

type Polygon []Vector

func NewPolygon(components ...Vector) Polygon {
	return Polygon(components)
}

func (p Polygon) Center() Vector {
	len := float64(len(p))
	if len == 0.0 {
		return Vector{}
	}
	r := Vector{}
	for _, c := range p {
		r = r.Add(c)
	}
	return r.Scale(1.0 / len)
}

func (p Polygon) Translate(v Vector) Polygon {
	components := make([]Vector, len(p))
	for i := range components {
		components[i] = p[i].Add(v)
	}
	return Polygon(components)
}

func (p Polygon) RotateAround(v Vector, angle float64) Polygon {
	components := make([]Vector, len(p))
	for i := range components {
		components[i] = p[i].RotateAround(v, angle)
	}
	return Polygon(components)
}

func (p Polygon) Points() []sdl.Point {
	points := make([]sdl.Point, len(p))
	for i := range points {
		points[i] = p[i].ToPoint()
	}
	if len(points) != 0 {
		points = append(points, points[0])
	}
	return points
}

type CarModel struct {
	Size, Position, Velocity, Acceleration Vector
	Bounds                                 Polygon
	Mass, WheelAngle                       float64
}

func (c *CarModel) ApplyForce(f Vector) {
	c.Acceleration = c.Acceleration.Add(f.Scale(1.0 / c.Mass))
}

func (c *CarModel) Accelerate(delta float64) {
	c.ApplyForce(Vector{1, 0}.Scale(delta))
}

func (c *CarModel) Break(delta float64) {
	c.ApplyForce(Vector{-1, 0}.Scale(delta))
}

func (c *CarModel) TurnLeft(delta float64) {
	c.WheelAngle += delta
	fmt.Println(c.WheelAngle)
}

func (c *CarModel) TurnRight(delta float64) {
	c.WheelAngle -= delta
	fmt.Println(c.WheelAngle)
}

func NewCar(mass, x, y float64) *CarModel {
	c := &CarModel{
		Size:         Vector{50, 100},
		Position:     Vector{100, 100},
		Velocity:     Vector{0, 0},
		Acceleration: Vector{0, 0},
		Mass:         mass,
		WheelAngle:   0,
	}
	c.Bounds = NewPolygon(Null, Null.AddX(c.Size), Null.Add(c.Size), Null.AddY(c.Size))
	return c
}

func (c *CarModel) Update(delta float64) {
	c.Acceleration = c.Acceleration.Scale(0.99)
	c.Velocity = c.Velocity.Scale(0.99).Add(c.Acceleration.Scale(delta))
	c.Position = c.Position.Add(c.Velocity.Scale(delta).RotateAround(Null, c.WheelAngle))
}

func (c *CarModel) Draw(r *sdl.Renderer) {
	center, angle := c.Bounds.Center(), c.Velocity.RotateAround(Null, c.WheelAngle).AngleBetween(Vector{0, 1})
	//fmt.Println(center, angle)
	transformed := c.Bounds.RotateAround(center, angle).Translate(c.Position)
	//fmt.Println(transformed)
	r.DrawLines(transformed.Points())
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

	car := NewCar(1, 100, 100)
	last := time.Now().UnixNano()
	for {
		renderer.SetDrawColor(0, 0, 0, 255)
		renderer.Clear()
		renderer.SetDrawColor(255, 0, 0, 255)
		car.Draw(renderer)

		delta := float64(time.Now().UnixNano()-last) / float64(time.Second)
		car.Update(delta)
		last = time.Now().UnixNano()
		//delta2 := float64(time.Now().UnixNano()-last) / float64(time.Second)

		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch et := event.(type) {
			case *sdl.KeyDownEvent:
				switch et.Keysym.Sym {
				case sdl.K_ESCAPE:
					return nil
				case sdl.K_w:
					car.Accelerate(100.0)
				case sdl.K_s:
					car.Break(100.0)
				case sdl.K_a:
					car.TurnLeft(delta)
				case sdl.K_d:
					car.TurnRight(delta)
				}
			}
		}
		renderer.Present()
		sdl.Delay(1)
	}
}
