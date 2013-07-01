// Copyright (c) 2013 Aubrey Barnard.  This is free software.  See
// LICENSE.txt for details.

// Go package that provides an interface to the Fortran implementation
// of the L-BFGS-B optimization algorithm.  The Fortran code is provided
// as a C-compatible library and this is the Go API for that library.
// There is a sliver of C code (in lbfgsb_go_interface.*) needed to
// connect this Go package to the Fortran library.

package lbfgsb

// Declarations for Cgo

// #cgo LDFLAGS: -L. -llbfgsb -lgfortran -lm
// #include "lbfgsb_go_interface.h"
import "C"

import (
	"fmt"
	"math"
	"reflect"
	"unsafe"
)

// Private constants
const (
	// 3 or so 80-character lines
	bufferSize = 250
)

// NewLbfgsb creates, initializes, and returns a new Lbfgsb solver
// object.  Equivalent to 'new(Lbfgsb).Init(dimensionality)'.  A
// zero-value Lbfgsb object is valid and needs no explicit construction.
// However, this constructor is convenient and explicit.
func NewLbfgsb(dimensionality int) *Lbfgsb {
	return new(Lbfgsb).Init(dimensionality)
}

// Lbfgsb provides the functionality of the Fortran L-BFGS-B optimizer
// as a Go object.  A Lbfgsb solver object contains the setup for an
// optimization problem of a particular dimensionality.  It stores
// bounds, parameters, and results so it is relatively lightweight
// (especially if no bounds are specified).  It can be re-used for other
// problems with the same dimensionality, but using a different solver
// object for each problem is probably better organization.  A
// zero-value Lbfgsb object is valid and needs no explicit construction.
// A solver object will perform unconstrained optimization unless bounds
// are set.
type Lbfgsb struct {
	// Dimensionality of the problem.  Zero is an invalid
	// dimensionality, so this also serves as an indicator of whether
	// this object has been initialized for computation.  Once the
	// dimensionality has been set it cannot be changed.  To be ready
	// for computation the following must be greater than zero:
	// dimensionality, approximationSize, fTolerance, gTolerance.
	dimensionality int

	// Problem specification.  Bounds may be nil or allocated fully.
	// Individual bounds may be omitted by placing NaNs or Infs.
	lowerBounds []float64
	upperBounds []float64

	// Parameters
	approximationSize int
	fTolerance        float64
	gTolerance        float64
	printControl      int

	// Statistics (do not embed or members will be public)
	statistics OptimizationStatistics
}

// Init initializes this Lbfgsb solver for problems of the given
// dimensionality.  Also sets default parameters that are not zero
// values.  Returns this for method chaining.  Ignores calls subsequent
// to the first (because a solver is intended for only a particular
// dimensionality).
func (lbfgsb *Lbfgsb) Init(dimensionality int) *Lbfgsb {
	// Only initialize if not previously initialized
	if lbfgsb.dimensionality == 0 {
		// Check for a valid dimensionality
		if dimensionality <= 0 {
			panic(fmt.Errorf("Lbfgsb: Optimization problem dimensionality %d <= 0.  Expected > 0.", dimensionality))
		}
		// Set up the solver.  Protect previous values so Init can be
		// called after other methods.
		lbfgsb.dimensionality = dimensionality
		if lbfgsb.approximationSize == 0 {
			lbfgsb.approximationSize = 5
		}
		if lbfgsb.fTolerance == 0.0 {
			lbfgsb.fTolerance = 1e-6
		}
		if lbfgsb.gTolerance == 0.0 {
			lbfgsb.gTolerance = 1e-6
		}
	}
	return lbfgsb
}

// SetBounds sets the upper and lower bounds on the individual
// dimensions to the given intervals resulting in a constrained
// optimization problem.  Individual bounds may be (+/-)Inf.
func (lbfgsb *Lbfgsb) SetBounds(bounds [][2]float64) *Lbfgsb {
	// Ensure object is initialized
	lbfgsb.Init(len(bounds))
	// Check dimensionality
	if lbfgsb.dimensionality != len(bounds) {
		panic(fmt.Errorf("Lbfgsb: Dimensionality of the bounds (%d) does not match the dimensionality of the solver (%d).", len(bounds), lbfgsb.dimensionality))
	}

	lbfgsb.lowerBounds = make([]float64, lbfgsb.dimensionality)
	lbfgsb.upperBounds = make([]float64, lbfgsb.dimensionality)
	for i, interval := range bounds {
		lbfgsb.lowerBounds[i] = interval[0]
		lbfgsb.upperBounds[i] = interval[1]
	}
	return lbfgsb
}

// SetBoundsAll sets the bounds of all the dimensions to [lower,upper].
// Init must be called first to set the dimensionality.
func (lbfgsb *Lbfgsb) SetBoundsAll(lower, upper float64) *Lbfgsb {
	// Check object has been initialized
	if lbfgsb.dimensionality == 0 {
		panic(fmt.Errorf("Lbfgsb: Init() must be called before SetAllBounds()."))
	}

	lbfgsb.lowerBounds = make([]float64, lbfgsb.dimensionality)
	lbfgsb.upperBounds = make([]float64, lbfgsb.dimensionality)
	for i := 0; i < lbfgsb.dimensionality; i++ {
		lbfgsb.lowerBounds[i] = lower
		lbfgsb.upperBounds[i] = upper
	}
	return lbfgsb
}

// SetBoundsSparse sets the bounds to only those in the given map;
// others are unbounded.  Each entry in the map is a (zero-based)
// dimension index mapped to a slice representing an interval.
// Individual bounds may be (+/-)Inf.  Init must be called first to set
// the dimensionality.
//
// The slice is interpreted as an interval as follows:
//
//     nil | []: [-Inf, +Inf]
//     [x]: [-|x|, |x|]
//     [l, u, ...]: [l, u]
func (lbfgsb *Lbfgsb) SetBoundsSparse(sparseBounds map[int][]float64) *Lbfgsb {
	// Check object has been initialized
	if lbfgsb.dimensionality == 0 {
		panic(fmt.Errorf("Lbfgsb: Init() must be called before SetAllBounds()."))
	}

	// If no bounds are given, clear the bounds
	if sparseBounds == nil || len(sparseBounds) == 0 {
		return lbfgsb.ClearBounds()
	}

	lbfgsb.lowerBounds = make([]float64, lbfgsb.dimensionality)
	lbfgsb.upperBounds = make([]float64, lbfgsb.dimensionality)
	nInf := math.Inf(-1)
	pInf := math.Inf(+1)
	for i := 0; i < lbfgsb.dimensionality; i++ {
		interval, exists := sparseBounds[i]
		if exists {
			if interval == nil || len(interval) == 0 {
				lbfgsb.lowerBounds[i] = nInf
				lbfgsb.upperBounds[i] = pInf
			} else if len(interval) == 1 {
				lbfgsb.upperBounds[i] = math.Abs(interval[0])
				lbfgsb.lowerBounds[i] = -lbfgsb.upperBounds[i]
			} else {
				lbfgsb.lowerBounds[i] = interval[0]
				lbfgsb.upperBounds[i] = interval[1]
			}
		} else {
			lbfgsb.lowerBounds[i] = nInf
			lbfgsb.upperBounds[i] = pInf
		}
	}
	return lbfgsb
}

// ClearBounds clears all bounds resulting in an unconstrained
// optimization problem.
func (lbfgsb *Lbfgsb) ClearBounds() *Lbfgsb {
	lbfgsb.lowerBounds = nil
	lbfgsb.upperBounds = nil
	return lbfgsb
}

// SetApproximationSize sets the amount of history (points and
// gradients) stored and used to approximate the inverse Hessian matrix.
// More history allows better approximation at the cost of more memory.
// The recommended range is [3,20].  Defaults to 5.
func (lbfgsb *Lbfgsb) SetApproximationSize(size int) *Lbfgsb {
	if size <= 0 {
		panic(fmt.Errorf("Lbfgsb: Approximation size %d <= 0.  Expected > 0.", size))
	}
	lbfgsb.approximationSize = size
	return lbfgsb
}

// SetFTolerance sets the tolerance of the precision of the objective
// function required for convergence.  Defaults to 1e-6.
func (lbfgsb *Lbfgsb) SetFTolerance(fTolerance float64) *Lbfgsb {
	if fTolerance <= 0.0 {
		panic(fmt.Errorf("Lbfgsb: F tolerance %g <= 0.  Expected > 0.", fTolerance))
	}
	lbfgsb.fTolerance = fTolerance
	return lbfgsb
}

// SetGTolerance sets the tolerance of the precision of the objective
// gradient required for convergence.  Defaults to 1e-6.
func (lbfgsb *Lbfgsb) SetGTolerance(gTolerance float64) *Lbfgsb {
	if gTolerance <= 0.0 {
		panic(fmt.Errorf("Lbfgsb: G tolerance %g <= 0.  Expected > 0.", gTolerance))
	}
	lbfgsb.gTolerance = gTolerance
	return lbfgsb
}

// SetFortranPrintControl sets the level of output verbosity from the
// Fortran L-BFGS-B code.  Defaults to 0, no output.  Ranges from 0 to
// 102: 1 displays a summary, 100 displays details of each iteration,
// 102 adds vectors (X and G) to the output.
func (lbfgsb *Lbfgsb) SetFortranPrintControl(verbosity int) *Lbfgsb {
	if verbosity < 0 {
		panic(fmt.Errorf("Lbfgsb: Print control %d < 0.  Expected >= 0.", verbosity))
	}
	lbfgsb.printControl = verbosity
	return lbfgsb
}

// Minimize optimizes the given objective using the L-BFGS-B algorithm.
// Implements OptimizationFunctionMinimizer.Minimize.
func (lbfgsb *Lbfgsb) Minimize(
	objective FunctionWithGradient,
	initialPoint []float64,
	parameters map[string]interface{}) (
		minimum PointValueGradient,
		exitStatus ExitStatus) {

	// Make sure object has been initialized
	lbfgsb.Init(len(initialPoint))

	// TODO OMG! split this out into some helper functions

	// Check dimensionality
	dim := len(initialPoint)
	dim_c := C.int(dim)
	if lbfgsb.dimensionality != dim {
		exitStatus.Code = USAGE_ERROR
		exitStatus.Message = fmt.Sprintf("Lbfgsb: Dimensionality of the initial point (%d) does not match the dimensionality of the solver (%d).", dim, lbfgsb.dimensionality)
		return
	}

	// Process the parameters.  Argument 'parameters' overrides but does
	// not replace class-level parameters.  The bounds are not part of
	// the parameters, they are part of the problem specification.
	var paramVal interface{}
	var ok bool
	// Approximation size
	approximationSize := lbfgsb.approximationSize
	if paramVal, ok = parameters["approximationSize"]; ok {
		approximationSize, ok = paramVal.(int)
		if !ok || approximationSize <= 0 {
			exitStatus.Code = USAGE_ERROR
			exitStatus.Message = fmt.Sprintf("Bad parameter value: approximationSize: %v.  Expected integer >= 1.", paramVal)
			return
		}
	}
	// F tolerance
	fTolerance := lbfgsb.fTolerance
	if paramVal, ok = parameters["fTolerance"]; ok {
		fTolerance, ok = paramVal.(float64)
		if !ok || fTolerance <= 0.0 {
			exitStatus.Code = USAGE_ERROR
			exitStatus.Message = fmt.Sprintf("Bad parameter value: fTolerance: %v.  Expected float >= 0.", paramVal)
			return
		}
	}
	// G tolerance
	gTolerance := lbfgsb.gTolerance
	if paramVal, ok = parameters["gTolerance"]; ok {
		gTolerance, ok = paramVal.(float64)
		if !ok || gTolerance <= 0.0 {
			exitStatus.Code = USAGE_ERROR
			exitStatus.Message = fmt.Sprintf("Bad parameter value: gTolerance: %v.  Expected float >= 0.", paramVal)
			return
		}
	}
	// Debug level
	printControl := lbfgsb.printControl
	if paramVal, ok = parameters["printControl"]; ok {
		printControl, ok = paramVal.(int)
		if !ok || printControl < 0 {
			exitStatus.Code = USAGE_ERROR
			exitStatus.Message = fmt.Sprintf("Bad parameter value: printControl: %v.  Expected integer >= 0.", paramVal)
			return
		}
	}

	// Set up bounds control.  Use a C-compatible type.
	boundsControl := make([]C.int, dim)
	if lbfgsb.lowerBounds != nil {
		for index, bound := range lbfgsb.lowerBounds {
			if !math.IsNaN(bound) && !math.IsInf(bound, -1) {
				boundsControl[index] = C.int(1)
			}
		}
	}
	if lbfgsb.upperBounds != nil {
		for index, bound := range lbfgsb.upperBounds {
			if !math.IsNaN(bound) && !math.IsInf(bound, -1) {
				// Map 0 -> 3, 1 -> 2
				boundsControl[index] = C.int(3 - boundsControl[index])
			}
		}
	}

	// Set up lower and upper bounds.  These must be different slices
	// than the ones in the Lbfgsb object because those must remain
	// unallocated if no bounds are specified.
	lowerBounds := makeCCopySlice_Float(lbfgsb.lowerBounds, dim)
	upperBounds := makeCCopySlice_Float(lbfgsb.upperBounds, dim)

	// Set up callbacks
	callbackData := &callbackData{objective: objective}
	callbackData_c := unsafe.Pointer(callbackData)
	doLogging_c := C.int(0)  // TODO
	logFunctionCallbackData_c := unsafe.Pointer(uintptr(0))  // TODO

	// Allocate arrays for return value
	minimum.X = make([]float64, dim)
	minimum.G = make([]float64, dim)

	// Convert parameters for C
	approximationSize_c := C.int(approximationSize)
	fTolerance_c := C.double(fTolerance)
	gTolerance_c := C.double(gTolerance)
	printControl_c := C.int(printControl)

	// Prepare buffers and arrays for C.  Avoid allocation in C land by
	// allocating compatible things in Go and passing their addresses.
	// The following arrays may not be iteroperably type-safe but this
	// is how they did it on the Cgo page: http://golang.org/cmd/cgo/.
	// (One could always allocate slices of C types, pass those, and
	// then copy out and convert the contents on return.)
	var boundsControl_c *C.int = &boundsControl[0]
	var lowerBounds_c *C.double = &lowerBounds[0]
	var upperBounds_c *C.double = &upperBounds[0]
	var x0_c *C.double = (*C.double)(&initialPoint[0])
	var minX_c *C.double = (*C.double)(&minimum.X[0])
	var minF_c *C.double = (*C.double)(&minimum.F)
	var minG_c *C.double = (*C.double)(&minimum.G[0])
	var iters_c, evals_c C.int
	// Status message
	statusMessageLength_c := C.int(bufferSize)
	var statusMessageBuffer [bufferSize]C.char
	statusMessage_c := (*C.char)(&statusMessageBuffer[0])

	// Call the actual L-BFGS-B procedure
	statusCode_c := C.lbfgsb_minimize_c(
		callbackData_c, dim_c,
		boundsControl_c, lowerBounds_c, upperBounds_c,
		approximationSize_c, fTolerance_c, gTolerance_c,
		x0_c, minX_c, minF_c, minG_c, &iters_c, &evals_c,
		printControl_c, doLogging_c, logFunctionCallbackData_c,
		statusMessage_c, statusMessageLength_c,
	)

	// Convert outputs
	// Exit status codes match between ExitStatusCode and the C enum
	exitStatus.Code = ExitStatusCode(statusCode_c)
	exitStatus.Message = C.GoString(statusMessage_c)
	// Minimum already populated because pointers to its members were
	// passed into C/Fortran

	// Save statistics
	lbfgsb.statistics.Iterations = int(iters_c)
	lbfgsb.statistics.FunctionEvaluations = int(evals_c)
	// Number of function and gradient evaluations is always the same
	lbfgsb.statistics.GradientEvaluations = lbfgsb.statistics.FunctionEvaluations

	return
}

// makeCCopySlice_Float creates a C copy of a Go slice.  If the Go slice
// is nil, then a slice of the given length is created.
func makeCCopySlice_Float(slice []float64, sliceLen int) (
	slice_c []C.double) {

	slice_c = make([]C.double, sliceLen)
	// Copy the Go slice to the C slice, converting elements
	if slice != nil {
		for i := 0; i < sliceLen; i++ {
			slice_c[i] = C.double(slice[i])
		}
	}
	return
}

// Statistics returns some statistics about the most recent
// minimization: the total number of iterations and the total numbers of
// function and gradient evaluations.
func (lbfgsb *Lbfgsb) OptimizationStatistics() OptimizationStatistics {
	return lbfgsb.statistics
}

// callbackData is a container for the actual objective function and
// related data.
type callbackData struct {
	objective FunctionWithGradient
}

// go_objective_function_callback is an adapter between the C callback
// and the Go callback for evaluating the objective function.  Exported
// to C for use as a function pointer.  Must match the signature of
// objective_function_type in lbfgsb_c.h.
//
//export go_objective_function_callback
func go_objective_function_callback(
	dim_c C.int, point_c, value_c *C.double,
	callbackData_c unsafe.Pointer,
	statusMessage_c *C.char, statusMessageLength_c C.int) (
		statusCode_c C.int) {

	var point []float64

	// Convert inputs
	dim := int(dim_c)
	wrapCArrayAsGoSlice_Float64(point_c, dim, &point)
	cbData := (*callbackData)(callbackData_c)

	// Evaluate the objective function
	// TODO handle panics
	value := cbData.objective.Evaluate(point)

	// Convert outputs
	*value_c = C.double(value)

	//fmt.Printf("go_objective_function_callback: %v; %v;\n", point, value)

	return
}

// go_objective_gradient_callback is an adapter between the C callback
// and the Go callback for evaluating the objective gradient.  Exported
// to C for use as a function pointer.  Must match the signature of
// objective_gradient_type in lbfgsb_c.h.
//
//export go_objective_gradient_callback
func go_objective_gradient_callback(
	dim_c C.int, point_c, gradient_c *C.double,
	callbackData_c unsafe.Pointer,
	statusMessage_c *C.char, statusMessageLength_c C.int) (
		statusCode_c C.int) {

	var point, gradient, gradRet []float64

	// Convert inputs
	dim := int(dim_c)
	wrapCArrayAsGoSlice_Float64(point_c, dim, &point)
	cbData := (*callbackData)(callbackData_c)

	// Evaluate the gradient of the objective function
	// TODO handle panics
	gradRet = cbData.objective.EvaluateGradient(point)

	// Convert outputs
	wrapCArrayAsGoSlice_Float64(gradient_c, dim, &gradient)
	copy(gradient, gradRet)

	//fmt.Printf("go_objective_gradient_callback: %v; %v;\n", point, gradient)

	return
}

// go_log_function_callback is an adapter between the C callback and the
// Go callback for logging information about each iteration.  Exported
// to C for use as a function pointer.  Must match the signature of
// lbfgsb_log_function_type in lbfgsb_c.h.
//
//export go_log_function_callback
func go_log_function_callback(
	logCallBackData_c unsafe.Pointer,
	iteration_c, fgEvals_c, fgEvalsTotal_c C.int, stepLength_c C.double,
	dim_c C.int, x *C.double, f C.double, g *C.double,
	fDelta, fDeltaBound, gNorm, gNormBound C.double) (
		statusCode_c C.int) {

	// TODO go_log_function_callback
	return
}

// wrapCArrayAsGoSlice_Float64 allows a C array to be treated as a Go
// slice.  Based on https://code.google.com/p/go-wiki/wiki/cgo.  This
// only works if the Go and C types happen to be interoperable (binary
// compatible), but that seems to be the case so far.
func wrapCArrayAsGoSlice_Float64(array *C.double, length int,
	slice *[]float64) {

	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(slice))
	sliceHeader.Cap = length
	sliceHeader.Len = length
	sliceHeader.Data = uintptr(unsafe.Pointer(array))
}
