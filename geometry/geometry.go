package geometry

import (
	"math"

	"github.com/veandco/go-sdl2/sdl"
)

var Null Vector = Vector{0, 0}
var NullVector Vector = Null
var X Vector = Vector{1, 0}
var Y Vector = Vector{0, 1}

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

func (p Polygon) Interpolate(v float64) Vector {
	pointCount := len(p)
	connections := make([]float64, pointCount+1)
	length := 0.0
	for i := 0; i < len(p); i++ {
		distance := p[(i+1)%pointCount].Add(p[i].Scale(-1)).Len()
		connections[i+1] = distance + connections[i]
		length += distance
	}
	target := math.Min(1.0, math.Max(0.0, v))
	for i := 1; i < pointCount+1; i++ {
		if target < connections[i]/length {
			a, b := connections[i-1]/length, connections[i]/length
			k := (target - a) / (b - a)
			t := p[i-1].Add(p[i%pointCount].Add(p[i-1].Scale(-1)).Scale(k))
			return t
		}
	}
	return Null
}
