// Copyright (c) 2013 Aubrey Barnard.  This is free software.  See
// LICENSE.txt for details.

// Go version of Fortran main.  Runs the Fortran gradient descent via C
// with Go callbacks for evaluation of objective function and gradient.

package main

import (
	"fmt"
)

// Class to combine objective function and gradient into one object.
// The sphere function is just a multi-dimensional parabola.
type SphereFunction struct{}

// Sphere function
func (sf *SphereFunction) Evaluate(point []float64) (value float64) {
	for _, coord := range point {
		value += coord * coord
	}
	return
}

// Sphere function gradient
func (sf *SphereFunction) EvaluateGradient(point []float64) (grad []float64) {
	grad = make([]float64, len(point))
	for i, coord := range point {
		grad[i] = 2.0 * coord
	}
	return
}

// Runs gradient descent on the objective and displays results
func main() {
	fmt.Printf("Go-Fortran optimization interface prototype\n\n")

	objective := &SphereFunction{}
	x0 := []float64{7.0, -8.0, 9.0}
	min, err := GradientDescent(objective, x0, 1e-1, 100)

	fmt.Printf("maing(\n")
	fmt.Printf("     x0: %v\n", x0)
	if err == nil {
		fmt.Printf("  min.X: %v\n", min.X)
		fmt.Printf("  min.G: %v\n", min.G)
		fmt.Printf("  min.F: %v\n", min.F)
	}
	fmt.Printf("  error: %v\n", err)
	fmt.Printf(")\n")
}
