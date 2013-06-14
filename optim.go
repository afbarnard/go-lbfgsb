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
	//
	// 'parameters' is a map containing parameters for the optimization
	// algorithm.  The parameters are optional and may be nil.  If
	// specified, the parameters override any previously-set parameters,
	// but only for the duration of this invocation.  Parameters whose
	// names appear exactly as keys in the map will be interpreted.  The
	// interpretation may fail if the given value cannot be converted to
	// a valid parameter value.  Any other contents of the parameter map
	// will be ignored.
	Minimize(objective FunctionWithGradient, initialPoint []float64, parameters map[string]interface{}) (
		minimum PointValueGradient, exitStatus ExitStatus)
}

////////////////////////////////////////
// Optimization inputs

// FunctionWithGradient is the interface for a function (f: R**n -> R)
// and its gradient (f': R**n -> R**n) suitable for use as an objective
// function for optimization.
type FunctionWithGradient interface {
	// Evaluate returns the value of the function at the given point.
	Evaluate(point []float64) float64
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

func (gof GeneralObjectiveFunction) Evaluate(point []float64) float64 {
	return gof.Function(point)
}

func (gof GeneralObjectiveFunction) EvaluateGradient(point []float64) []float64 {
	return gof.Gradient(point)
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
	return fmt.Sprintf("Exit status: %v; Message: %v", es.Code, es.Message)
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
