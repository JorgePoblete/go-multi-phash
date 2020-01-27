package main

import (
	"math"
)

type Compute struct{}

func (c Compute) RGBRootMeanSquared(r, g, b []float64) float64 {
	var mean float64 = 0
	for i, _ := range r {
		mean += ((r[i] * r[i]) + (g[i] * g[i]) + (b[i] * b[i]))
	}
	return math.Sqrt(mean / float64(len(r)))
}

func (c Compute) RootMeanSquared(a []float64) float64 {
	var mean float64 = 0
	for i, _ := range a {
		mean += (a[i] * a[i])
	}
	return math.Sqrt(mean / float64(len(a)))
}

func (c Compute) Mean(a []float64) float64 {
	var mean float64 = 0
	for _, val := range a {
		mean += val
	}
	return mean / float64(len(a))
}

func (c Compute) IntegerDotMultiplication(a, b []int) int {
	var result int = 0
	for i, _ := range a {
		result += a[i] * b[i]
	}
	return result
}

func (c Compute) IntegerSum(a []int) int {
	var result int = 0
	for i, _ := range a {
		result += a[i]
	}
	return result
}

func (c Compute) Variance(a []float64) float64 {
	var variance float64 = 0
	mean := c.Mean(a)
	for _, val := range a {
		variance += math.Pow(val-mean, 2)

	}
	return variance / float64(len(a))
}

func (c Compute) Covariance(a, b []float64) float64 {
	var covariance float64 = 0
	meanA := c.Mean(a)
	meanB := c.Mean(b)
	for i, _ := range a {
		dA := a[i] - meanA
		dB := b[i] - meanB
		covariance += (dA*dB - covariance) / float64(i+1)
	}
	return covariance * float64(len(a)) / float64(len(a)-1)
}

func (c Compute) luminance(x, y []float64, C float64) float64 {
	ux := c.Mean(x)
	uy := c.Mean(y)
	return (2*ux*uy + C) / ((ux * ux) + (uy * uy) + C)
}

func (c Compute) contrast(x, y []float64, C float64) float64 {
	sx2 := c.Variance(x)
	sy2 := c.Variance(y)
	sx := math.Sqrt(sx2)
	sy := math.Sqrt(sy2)

	return (2*sx*sy + C) / (sx2 + sy2 + C)
}

func (c Compute) structure(x, y []float64, C float64) float64 {
	var sx = math.Sqrt(c.Variance(x))
	var sy = math.Sqrt(c.Variance(y))
	var sxy = c.Covariance(x, y)

	return (sxy + C) / (sx*sy + C)
}

func (c Compute) SSIM(a, b []float64) float64 {
	L, K1, K2 := 255.0, 0.01, 0.03
	C1 := (K1 * L) * (K1 * L)
	C2 := (K2 * L) * (K2 * L)
	C3 := C2 / 2

	luminance := c.luminance(a, b, C1)
	contrast := c.contrast(a, b, C2)
	structure := c.structure(a, b, C3)
	return luminance * contrast * structure
}

func (c Compute) Manhattan(a, b []float64) float64 {
	var distance float64 = 0

	for i, _ := range a {
		distance += math.Abs(float64(a[i] - b[i]))
	}

	return distance
}

func (c Compute) MSE(a, b []float64) float64 {
	err := 0.0
	for i, _ := range a {
		err = err + (math.Pow(a[i]-b[i], 2))
	}
	return err / float64(len(a))
}
