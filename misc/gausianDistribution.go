package misc

import (
	"math"
)

func GenerageGausianDistribution(size int, o float64) []float64 {
	var result = make([]float64, size*size)

	for x := 0; x < size; x++ {
		for y := 0; y < size; y++ {
			result[size*x+y] = gaussianEquation(x-size/2, y-size/2, o)
		}
	}
	return result
}

func gaussianEquation(x int, y int, o float64) float64 {
	exponent := float64((x*x)+(y*y)) / (2 * o * o)
	base := (1 / (2 * math.Pi * o * o)) * math.Pow(math.E, -exponent)

	return base
}
