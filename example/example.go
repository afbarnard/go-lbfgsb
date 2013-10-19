// Copyright (c) 2013 Aubrey Barnard.  This is free software.  See
// LICENSE.txt for details.

// Main package providing some examples of how to use the lbfgsb
// package.

// TODO example program for Fortran
// TODO example program for C

// Install go-lbfgsb and build by running the following commands from
// this directory:
//
//     $ go get -d github.com/afbarnard/go-lbfgsb
//     $ pushd ~/go-pkgs/src/github.com/afbarnard/go-lbfgsb
//     $ make
//     $ popd
//     $ go get github.com/afbarnard/go-lbfgsb
//     $ go build
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
//
//     $ rm -R ~/go-pkgs/src/github.com/afbarnard/go-lbfgsb
//     $ find ~/go-pkgs -type d -empty -not -path '*/.*' -delete
package main

import (
	"fmt"
	"log"
	"math"
	"os"

	// Import go-lbfgsb as a "third-party" package (since it was
	// installed with 'go get').  Import as a local package if go-lbfgsb
	// was installed into your workspace.  For example, if you put the
	// code for this package in 'optim/go-lbfgsb', then 'import lbfgsb
	// "optim/go-lbfgsb"'.  To allow this example program to use this
	// package directly without installing anything and when this
	// package exists outside a Go workspace, use 'import lbfgsb ".."'.
	lbfgsb "github.com/afbarnard/go-lbfgsb"
	//lbfgsb "go-lbfgsb"
	//lbfsgb ".."
)

func main() {
	fmt.Printf("----- Go L-BFGS-B Example Program -----\n\n")

	////////////////////////////////////////
	// Example 1: Basic usage

	// Create default L-BFGS-B optimizer.  The solver adapts to any
	// initial dimensionality but then must stick with that
	// dimensionality.
	sphereOptimizer := new(lbfgsb.Lbfgsb)

	// Create sphere objective function as FunctionWithGradient object
	sphereObjective := new(SphereFunction)
	sphereMin := lbfgsb.PointValueGradient{
		X: []float64{0.0, 0.0, 0.0, 0.0, 0.0},
		F: 0.0,
		G: []float64{0.0, 0.0, 0.0, 0.0, 0.0},
	}

	// Minimize sphere function
	fmt.Printf("----- Sphere Function -----\n")
	x0_5d := []float64{10.0, -9.0, 8.0, -7.0, 6.0}
	minimum, exitStatus := sphereOptimizer.Minimize(sphereObjective, x0_5d)
	stats := sphereOptimizer.OptimizationStatistics()
	PrintResults(sphereMin, minimum, exitStatus, stats)

	////////////////////////////////////////
	// Example 2: Setting parameters

	// Create a new solver for a new problem with a different
	// dimensionality.  Make the tolerances strict.
	rosenOptimizer := lbfgsb.NewLbfgsb(2).
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

	// Minimize Rosenbrock
	fmt.Printf("----- Rosenbrock Function -----\n")
	x0_2d := []float64{10.0, 11.0}
	minimum, exitStatus = rosenOptimizer.Minimize(rosenObjective, x0_2d)
	stats = rosenOptimizer.OptimizationStatistics()
	PrintResults(rosenMin, minimum, exitStatus, stats)

	////////////////////////////////////////
	// Example 3: Bounds

	// Create bounds (box constraints) explicitly for each dimension.
	// Each pair represents an interval.
	bounds := [][2]float64{
		{1.0, 10.0},
		{-10.0, -1.0},
		{1.0, math.Inf(1.0)},
		{math.Inf(-1.0), -1.0},
		{math.Inf(-1.0), math.Inf(1.0)},
	}
	// Tell the optimizer about the bounds
	sphereOptimizer.SetBounds(bounds)
	// The constrained minimum is different
	sphereBoundsMin := lbfgsb.PointValueGradient{
		X: []float64{1.0, -1.0, 1.0, -1.0, 0.0},
		F: 4.0,
		G: []float64{2.0, -2.0, 2.0, -2.0, 0.0},
	}

	// Minimize sphere function subject to bounds
	fmt.Printf("----- Sphere Function with Bounds -----\n")
	minimum, exitStatus = sphereOptimizer.Minimize(sphereObjective, x0_5d)
	stats = sphereOptimizer.OptimizationStatistics()
	PrintResults(sphereBoundsMin, minimum, exitStatus, stats)

	// Clear bounds
	sphereOptimizer.ClearBounds()

	////////////////////////////////////////
	// Example 4: Logging

	// Minimize sphere function again, but with logging
	fmt.Printf("----- Sphere Function with Logging -----\n")
	// Create a logger
	logger := log.New(os.Stderr, "log: ", log.LstdFlags)
	// Make a closure for the logging function
	sphereOptimizer.SetLogger(
		func(info *lbfgsb.OptimizationIterationInformation) {
			LogOptimizationIteration(logger, info)
		})
	minimum, exitStatus = sphereOptimizer.Minimize(sphereObjective, x0_5d)
	stats = sphereOptimizer.OptimizationStatistics()
	PrintResults(sphereMin, minimum, exitStatus, stats)

	// Remove logger
	sphereOptimizer.SetLogger(nil)

	////////////////////////////////////////
	// Example 5: Usage errors

	// Create eventual error by reversing bounds
	sphereOptimizer.SetBoundsAll(5.0, -5.0)
	fmt.Printf("----- Sphere Function with Usage Error -----\n")
	minimum, exitStatus = sphereOptimizer.Minimize(sphereObjective, x0_5d)
	stats = sphereOptimizer.OptimizationStatistics()
	PrintResults(sphereMin, minimum, exitStatus, stats)
	sphereOptimizer.ClearBounds()
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

// Logs an optimization iteration to the given logger
func LogOptimizationIteration(logger *log.Logger,
	info *lbfgsb.OptimizationIterationInformation) {

	// Print a header every 10 lines
	if (info.Iteration-1)%10 == 0 {
		logger.Println(info.Header())
	}
	// Print the information
	logger.Println(info.String())
}
