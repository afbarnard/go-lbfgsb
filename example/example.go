// Copyright (c) 2013 Aubrey Barnard.  This is free software.  See
// LICENSE.txt for details.

// Main package providing some examples of how to use the lbfgsb
// package.

// Build by running the following two commands:
//
//     $ go get github.com/afbarnard/go-lbfgsb
//     $ go build
//
// The 'go get' command only needs to be run once.
//
// Then run.  Fun!
//
//     $ ./example
//
// To uninstall:
//
//     $ go clean -i github.com/afbarnard/go-lbfgsb
//
// You will have to remove the sources and extraneous directories
// yourself.  Unfortunately 'go clean' does not appear to be that
// sophisticated yet.
package main

import (
	"fmt"

	lbfgsb "github.com/afbarnard/go-lbfgsb"
)

func main() {
	fmt.Printf("----- Go L-BFGS-B Example Program -----\n\n")

	// Create default L-BFGS-B optimizer
	optimizer := new(lbfgsb.Lbfgsb)

	// Create Rosenbrock objective function
	rosenObjective := new(RosenbrockFunction)
	rosenMin := lbfgsb.PointValueGradient{
		X: []float64{1.0, 1.0},
		F: 0.0,
		G: []float64{0.0, 0.0},
	}

	// Minimize Rosenbrock without additional parameters
	x0 := []float64{10.0, 10.0}
	minimum, exitStatus := optimizer.Minimize(rosenObjective, x0, nil)
	fmt.Printf("----- Rosenbrock Function -----\nexpected: %v\n minimum: %v\n  status: %v\n\n",
		rosenMin, minimum, exitStatus)

	// Create sphere objective function using an objective function
	// object to combine arbitrary functions
	sphereObjective := lbfgsb.GeneralObjectiveFunction{
		Function: SphereFunction,
		Gradient: SphereGradient,
	}
	sphereMin := lbfgsb.PointValueGradient{
		X: []float64{0.0, 0.0, 0.0, 0.0, 0.0},
		F: 0.0,
		G: []float64{0.0, 0.0, 0.0, 0.0, 0.0},
	}

	// Minimize sphere function
	x0 = []float64{10.0, 10.0, 10.0, 10.0, 10.0}
	minimum, exitStatus = optimizer.Minimize(sphereObjective, x0, nil)
	fmt.Printf("----- Sphere Function -----\nexpected: %v\n minimum: %v\n  status: %v\n\n",
		sphereMin, minimum, exitStatus)

	// TODO example with bounds
}

// Famous Rosenbrock function as a FunctionWithGradient object
type RosenbrockFunction struct {}

// Rosenbrock function value
func (rf RosenbrockFunction) Evaluate(point []float64) (value float64) {
	for i := 0; i < len(point) - 1; i++ {
		value += Pow2(1.0 - point[i]) +
			100.0 * Pow2(point[i + 1] - Pow2(point[i]))
	}
	return
}

// Rosenbrock function gradient
func (rf RosenbrockFunction) EvaluateGradient(point []float64) (gradient []float64) {
	gradient = make([]float64, len(point))
	gradient[0] = -400.0 * point[0] * (point[1] - Pow2(point[0])) -
		2.0 * (1.0 - point[0])
	var i int
	for i = 1; i < len(point) - 1; i++ {
		gradient[i] = -400.0 * point[i] * (point[i + 1] - Pow2(point[i])) -
			2.0 * (1.0 - 101.0 * point[i] + 100.0 * Pow2(point[i - 1]))
	}
	gradient[i] = 200.0 * (point[i] - Pow2(point[i - 1]))
	return
}

// Simple square function rather than calling math.Pow
func Pow2(x float64) float64 {
	return x * x
}

// Sphere function
func SphereFunction(point []float64) (value float64) {
	for _, x := range point {
		value += x * x
	}
	return
}

// Sphere function gradient
func SphereGradient(point []float64) (gradient []float64) {
	gradient = make([]float64, len(point))
	for i, x := range point {
		gradient[i] = 2.0 * x
	}
	return
}
