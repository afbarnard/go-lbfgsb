// Copyright (c) 2014 Aubrey Barnard.  This is free software.  See
// LICENSE.txt for details.

// General interface to optimization algorithms.

package lbfgsb

import (
	"fmt"
)

////////////////////////////////////////
// Optimization algorithms

// ObjectiveFunctionMinimizer is the interface for all n-dimensional
// numerical minimization optimization algorithms that use gradients.
// Finds a local minimum.  The idea is that an optimization algorithm
// object will provide this method as well as methods for setting
// parameters, getting performance results, logging, etc., methods that
// are specific to the implementation.  This interface should specify
// the minimum necessary to be a useful optimization algorithm.
type ObjectiveFunctionMinimizer interface {
	// Minimize finds a numerically-approximate minimum of the given
	// objective function starting from the given point.  Returns the
	// minimum (or the best point found) and the status of the algorithm
	// at exit.
	Minimize(objective FunctionWithGradient, initialPoint []float64) (
		minimum PointValueGradient, exitStatus ExitStatus)
}

////////////////////////////////////////
// Optimization inputs

// FunctionWithGradient is the interface for a function (f: R**n -> R)
// and its gradient (f': R**n -> R**n) suitable for use as an objective
// function for optimization.
type FunctionWithGradient interface {
	// EvaluateFunction returns the value of the function at the given
	// point.
	EvaluateFunction(point []float64) float64
	// EvaluateGradient returns the gradient of the function at the
	// given point.
	EvaluateGradient(point []float64) []float64
}

// GeneralObjectiveFunction is a utility object that combines individual
// Go functions into a FunctionWithGradient.
type GeneralObjectiveFunction struct {
	Function func([]float64) float64
	Gradient func([]float64) []float64
}

func (gof GeneralObjectiveFunction) EvaluateFunction(point []float64) float64 {
	return gof.Function(point)
}

func (gof GeneralObjectiveFunction) EvaluateGradient(point []float64) []float64 {
	return gof.Gradient(point)
}

// OptimizationIterationLogger is the type of function that
// logs/records/processes information about a single iteration in an
// optimization run.
type OptimizationIterationLogger func(info *OptimizationIterationInformation)

// OptimizationIterationInformation is a container for information about
// an optimization iteration.
type OptimizationIterationInformation struct {
	Iteration   int
	FEvals      int
	GEvals      int
	FEvalsTotal int
	GEvalsTotal int
	StepLength  float64
	X           []float64
	F           float64
	G           []float64
	FDelta      float64
	FDeltaBound float64
	GNorm       float64
	GNormBound  float64
}

// Header returns a string with descriptions for the fields returned by
// String().
func (oii *OptimizationIterationInformation) Header() string {
	return "iter, f(x), step, df(x) <1?, ||f'(x)|| <1?, #f(), #g()"
}

// String formats the iteration information fields as a row in a table.
func (oii *OptimizationIterationInformation) String() string {
	// Close to convergence for F?
	fConvRatio := oii.FDelta / oii.FDeltaBound
	fConvIndicator := "F"
	if fConvRatio < 1.0 {
		fConvIndicator = "T"
	}
	// Close to convergence for G?
	gConvRatio := oii.GNorm / oii.GNormBound
	gConvIndicator := "F"
	if gConvRatio < 1.0 {
		gConvIndicator = "T"
	}
	// Put all the fields together
	return fmt.Sprintf("%d %g %g %g %.2g%v %g %.2g%v %d %d",
		oii.Iteration, oii.F, oii.StepLength,
		oii.FDelta, fConvRatio, fConvIndicator,
		oii.GNorm, gConvRatio, gConvIndicator,
		oii.FEvals, oii.GEvals)
}

////////////////////////////////////////
// Optimization outputs

// PointValueGradient is a point in optimization space as well as the
// result of optimization: a point (x), its function value (f), and its
// gradient (g).  The lengths of X and G must agree.
type PointValueGradient struct {
	X []float64
	F float64
	G []float64
}

// ExitStatusCode describes the exit status of an optimization
// algorithm.  Multiple statuses are necessary because success in
// optimization is not binary and so a simple error is not adequate.
// There are four exit statuses to distinguish:
//
// 1. Success.  Normal termination having converged.
//
// 2. Approximate success.  Normal operation resulting in a more
// approximate answer.  For example, unable to meet termination
// tolerances.
//
// 3. Warning.  The result could be OK, but there were some issues that
// may have reduced the quality of the result and require examination.
// For example, slight numerical problems, exceeding iteration or time
// bounds.
//
// 4. Failure of optimization.  For example, a necessary condition of
// the algorithm was not met, severe numerical problems.  (This status
// includes failures to evaluate the objective function or objective
// gradient.)
//
// There are also the typical runtime errors due to usage, bugs, etc.:
//
// 5. Usage error.  For example, invalid constraints, bad parameters.
// Responsibility is on the caller.
//
// 6. Internal error.  Other runtime or programming/logic error which
// may be a bug.  Responsibility is on this package.
type ExitStatusCode uint8

// ExitStatusCode values.
const (
	SUCCESS ExitStatusCode = iota
	APPROXIMATE
	WARNING
	FAILURE
	USAGE_ERROR
	INTERNAL_ERROR
)

// String returns a word for each ExitStatusCode.
func (esc ExitStatusCode) String() string {
	switch esc {
	case SUCCESS:
		return "SUCCESS"
	case APPROXIMATE:
		return "APPROXIMATE"
	case WARNING:
		return "WARNING"
	case FAILURE:
		return "FAILURE"
	case USAGE_ERROR:
		return "USAGE_ERROR"
	case INTERNAL_ERROR:
		return "INTERNAL_ERROR"
	default:
		return "UNKNOWN"
	}
}

// ExitStatus is the exit status of an optimization algorithm.  Includes
// a status code and a message explaining the situation.
type ExitStatus struct {
	Code    ExitStatusCode
	Message string
}

// String returns the exit status code and message as text.
func (es ExitStatus) String() string {
	return fmt.Sprintf("Exit status: %v; Message: %v;", es.Code, es.Message)
}

// Error allows this ExitStatus to be treated like an error object.
func (es ExitStatus) Error() string {
	return es.String()
}

// AsError returns an error representing this exit status.  If the exit
// status code is 'SUCCESS' then AsError returns nil.  Otherwise returns
// an error object (which happens to be this object).
func (es ExitStatus) AsError() error {
	if es.Code != SUCCESS {
		return &es
	}
	return nil
}

// OptimizationStatistics is a container for basic statistics about an
// optimization run.  Values can be negative to indicate they were not
// tracked.
type OptimizationStatistics struct {
	Iterations          int
	FunctionEvaluations int
	GradientEvaluations int
}

// OptimizationStatisticser is an object that can supply statistics
// about an optimization run.
type OptimizationStatisticser interface {
	OptimizationStatistics() OptimizationStatistics
}
