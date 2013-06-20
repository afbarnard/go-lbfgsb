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

// Lbfgsb provides the functionality of the Fortran L-BFGS-B optimizer
// as a Go object.
type Lbfgsb struct {
	// Problem specification.  Bounds may be nil or allocated fully.
	// Individual bounds may be omitted by placing NaNs or Infs.
	lowerBounds []float64
	upperBounds []float64
	// Parameters
	approximationSize int
	fTolerance        float64
	gTolerance        float64
	debug             int
}

// Minimize optimizes the given objective using the L-BFGS-B algorithm.
// Implements OptimizationFunctionMinimizer.Minimize.
func (lbfgsb Lbfgsb) Minimize(
	objective FunctionWithGradient,
	initialPoint []float64,
	parameters map[string]interface{}) (
		minimum PointValueGradient,
		exitStatus ExitStatus) {

	// TODO OMG! split this out into some helper functions

	// Check everyone agrees on the dimensionality
	dim := len(initialPoint)
	dim_c := C.int(dim)
	if lbfgsb.lowerBounds != nil && len(lbfgsb.lowerBounds) != dim {
		exitStatus.Code = USAGE_ERROR
		exitStatus.Message = fmt.Sprintf("Dimensionality disagreement: initialPoint: %v, lowerBounds: %v", dim, len(lbfgsb.lowerBounds))
	}
	if lbfgsb.upperBounds != nil && len(lbfgsb.upperBounds) != dim {
		exitStatus.Code = USAGE_ERROR
		exitStatus.Message = fmt.Sprintf("Dimensionality disagreement: initialPoint: %v, upperBounds: %v", dim, len(lbfgsb.upperBounds))
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
	// Debug
	debug := lbfgsb.debug
	if paramVal, ok = parameters["debug"]; ok {
		debug, ok = paramVal.(int)
		if !ok || debug < 0 {
			exitStatus.Code = USAGE_ERROR
			exitStatus.Message = fmt.Sprintf("Bad parameter value: debug: %v.  Expected integer >= 0.", paramVal)
			return
		}
	}

	// Set up bounds control.  Use a C-compatible type.
	boundsControl := make([]int32, dim)
	if lbfgsb.lowerBounds != nil {
		for index, bound := range lbfgsb.lowerBounds {
			if !math.IsNaN(bound) && !math.IsInf(bound, -1) {
				boundsControl[index] = 1
			}
		}
	}
	if lbfgsb.upperBounds != nil {
		for index, bound := range lbfgsb.upperBounds {
			if !math.IsNaN(bound) && !math.IsInf(bound, -1) {
				// Map 0 -> 3, 1 -> 2
				boundsControl[index] = 3 - boundsControl[index]
			}
		}
	}

	// Set up callback
	callbackData := &callbackData{objective: objective}
	callbackData_c := unsafe.Pointer(callbackData)

	// Allocate arrays for return value
	minimum.X = make([]float64, dim)
	minimum.G = make([]float64, dim)

	// Convert parameters for C
	approximationSize_c := C.int(approximationSize)
	fTolerance_c := C.double(fTolerance)
	gTolerance_c := C.double(gTolerance)
	debug_c := C.int(debug)

	// Prepare buffers and arrays for C.  Avoid allocation in C land by
	// allocating compatible things in Go and passing their addresses.
	// The following arrays may not be iteroperably type-safe but this
	// is how they did it on the Cgo page: http://golang.org/cmd/cgo/.
	// (One could always allocate slices of C types, pass those, and
	// then copy out and convert the contents on return.)
	var boundsControl_c *C.int = (*C.int)(&boundsControl[0])
	var lowerBounds_c *C.double = (*C.double)(&lbfgsb.lowerBounds[0])
	var upperBounds_c *C.double = (*C.double)(&lbfgsb.upperBounds[0])
	var x0_c *C.double = (*C.double)(&initialPoint[0])
	var minX_c *C.double = (*C.double)(&minimum.X[0])
	var minF_c *C.double = (*C.double)(&minimum.F)
	var minG_c *C.double = (*C.double)(&minimum.G[0])
	// Status message
	statusMessageLength_c := C.int(bufferSize)
	var statusMessageBuffer [bufferSize]C.char
	statusMessage_c := (*C.char)(&statusMessageBuffer[0])

	// Call the actual L-BFGS-B procedure
	statusCode_c := C.lbfgsb_minimize_c(
		callbackData_c, dim_c,
		boundsControl_c, lowerBounds_c, upperBounds_c,
		approximationSize_c, fTolerance_c, gTolerance_c,
		x0_c, minX_c, minF_c, minG_c,
		statusMessageLength_c, statusMessage_c, debug_c,
	)

	// Convert outputs
	// Exit status codes match between ExitStatusCode and the C enum
	exitStatus.Code = ExitStatusCode(statusCode_c)
	exitStatus.Message = C.GoStringN(statusMessage_c, statusMessageLength_c)
	// Minimum already populated because pointers to its members were
	// passed into C/Fortran

	return
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
	statusMessageLength_c C.int, statusMessage_c *C.char) (
		statusCode_c C.int) {

	var point []float64

	// Convert inputs
	dim := int(dim_c)
	wrapCArrayAsGoSlice_Float64(point_c, dim, &point)
	cbData := (*callbackData)(callbackData_c)

	// Evaluate the objective function
	value := cbData.objective.Evaluate(point)

	// Convert outputs
	*value_c = C.double(value)

	return
}

// go_objective_gradient_callback is an adapter between the C callback
// and the Go callback for evaluating the objective gradient.  Exported
// to C for use as a function pointer.  Must match the signature of
// objective_gradient_type in gd_c.h.
//
//export go_objective_gradient_callback
func go_objective_gradient_callback(
	dim_c C.int, point_c, gradient_c *C.double,
	callbackData_c unsafe.Pointer,
	statusMessageLength_c C.int, statusMessage_c *C.char) (
		statusCode_c C.int) {

	var point, gradient, gradRet []float64

	// Convert inputs
	dim := int(dim_c)
	wrapCArrayAsGoSlice_Float64(point_c, dim, &point)
	cbData := (*callbackData)(callbackData_c)

	// Evaluate the gradient of the objective function
	gradRet = cbData.objective.EvaluateGradient(point)

	// Convert outputs
	wrapCArrayAsGoSlice_Float64(gradient_c, dim, &gradient)
	copy(gradient, gradRet)

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
