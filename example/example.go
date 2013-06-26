// Copyright (c) 2013 Aubrey Barnard.  This is free software.  See
// LICENSE.txt for details.

// Main package providing some examples of how to use the lbfgsb
// package.

// TODO example program for Fortran
// TODO example program for C

// TODO fix these build instructions
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

	lbfgsb ".."
)

func main() {
	fmt.Printf("----- Go L-BFGS-B Example Program -----\n\n")

	// Create default L-BFGS-B optimizer.  The solver adapts to any
	// initial dimensionality but then must stick with that
	// dimensionality.
	optimizer := new(lbfgsb.Lbfgsb)

	// Create sphere objective function as FunctionWithGradient object
	sphereObjective := new(SphereFunction)
	sphereMin := lbfgsb.PointValueGradient{
		X: []float64{0.0, 0.0, 0.0, 0.0, 0.0},
		F: 0.0,
		G: []float64{0.0, 0.0, 0.0, 0.0, 0.0},
	}

	// Minimize sphere function without additional parameters
	fmt.Printf("----- Sphere Function -----\n")
	x0 := []float64{10.0, 10.0, 10.0, 10.0, 10.0}
	minimum, exitStatus := optimizer.Minimize(sphereObjective, x0, nil)
	stats := optimizer.OptimizationStatistics()
	PrintResults(sphereMin, minimum, exitStatus, stats)

	// Create a new solver for a new problem with a different
	// dimensionality.  Make the tolerances strict.
	optimizer = lbfgsb.NewLbfgsb(2).
		SetFTolerance(1e-10).SetGTolerance(1e-10)

	// Create Rosenbrock objective function by composing individual
	// value and gradient functions
	rosenObjective := lbfgsb.GeneralObjectiveFunction{
		Function: RosenbrockFunction,
		Gradient: RosenbrockGradient,
	}
	rosenMin := lbfgsb.PointValueGradient{
		X: []float64{1.0, 1.0},
		F: 0.0,
		G: []float64{0.0, 0.0},
	}

	// Minimize Rosenbrock without additional parameters
	fmt.Printf("----- Rosenbrock Function -----\n")
	x0 = []float64{10.0, 10.0}
	minimum, exitStatus = optimizer.Minimize(rosenObjective, x0, nil)
	stats = optimizer.OptimizationStatistics()
	PrintResults(rosenMin, minimum, exitStatus, stats)

	// TODO example with bounds

	// TODO example with user errors

	// TODO example with optimization errors
}

// Sphere (multi-dimensional parabola) function as a FunctionWithGradient object
type SphereFunction struct{}

// Sphere function
func (sf SphereFunction) Evaluate(point []float64) (value float64) {
	for _, x := range point {
		value += x * x
	}
	return
}

// Sphere function gradient
func (sf SphereFunction) EvaluateGradient(point []float64) (gradient []float64) {
	gradient = make([]float64, len(point))
	for i, x := range point {
		gradient[i] = 2.0 * x
	}
	return
}

// Rosenbrock function value
func RosenbrockFunction(point []float64) (value float64) {
	for i := 0; i < len(point)-1; i++ {
		value += Pow2(1.0-point[i]) +
			100.0*Pow2(point[i+1]-Pow2(point[i]))
	}
	return
}

// Rosenbrock function gradient
func RosenbrockGradient(point []float64) (gradient []float64) {
	gradient = make([]float64, len(point))
	gradient[0] = -400.0*point[0]*(point[1]-Pow2(point[0])) -
		2.0*(1.0-point[0])
	var i int
	for i = 1; i < len(point)-1; i++ {
		gradient[i] = -400.0*point[i]*(point[i+1]-Pow2(point[i])) -
			2.0*(1.0-101.0*point[i]+100.0*Pow2(point[i-1]))
	}
	gradient[i] = 200.0 * (point[i] - Pow2(point[i-1]))
	return
}

// Simple square function rather than calling math.Pow
func Pow2(x float64) float64 {
	return x * x
}

// Displays expected and actual results of optimization
func PrintResults(expectedMin, actualMin lbfgsb.PointValueGradient,
	exitStatus lbfgsb.ExitStatus, stats lbfgsb.OptimizationStatistics) {

	fmt.Printf("expected: X: %v\n          F: %v\n          G: %v\n",
		expectedMin.X, expectedMin.F, expectedMin.G)
	fmt.Printf(" minimum: X: %v\n          F: %v\n          G: %v\n",
		actualMin.X, actualMin.F, actualMin.G)
	fmt.Printf("  status: %v\n", exitStatus)
	fmt.Printf("   stats: iters: %v; F evals: %v; G evals: %v\n\n",
		stats.Iterations, stats.FunctionEvaluations, stats.GradientEvaluations)
}
