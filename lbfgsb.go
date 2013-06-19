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
	buffer_size = 250
)

// Lbfgsb provides the functionality of the Fortran L-BFGS-B optimizer
// as a Go object.
type Lbfgsb struct {
	// Problem specification.  Bounds may be nil or allocated fully.
	// Individual bounds may be omitted by placing NaNs or Infs.
	lower_bounds []float64
	upper_bounds []float64
	// Parameters
	approximation_size int
	f_tolerance        float64
	g_tolerance        float64
	debug              int
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
	if lbfgsb.lower_bounds != nil && len(lbfgsb.lower_bounds) != dim {
		exitStatus.Code = USAGE_ERROR
		exitStatus.Message = fmt.Sprintf("Dimensionality disagreement: initial_point: %v, lower_bounds: %v", dim, len(lbfgsb.lower_bounds))
	}
	if lbfgsb.upper_bounds != nil && len(lbfgsb.upper_bounds) != dim {
		exitStatus.Code = USAGE_ERROR
		exitStatus.Message = fmt.Sprintf("Dimensionality disagreement: initial_point: %v, upper_bounds: %v", dim, len(lbfgsb.upper_bounds))
	}

	// Process the parameters.  Argument 'parameters' overrides but does
	// not replace class-level parameters.  The bounds are not part of
	// the parameters, they are part of the problem specification.
	var paramVal interface{}
	var ok bool
	// Approximation size
	approximation_size := lbfgsb.approximation_size
	if paramVal, ok = parameters["approximation_size"]; ok {
		approximation_size, ok = paramVal.(int)
		if !ok || approximation_size <= 0 {
			exitStatus.Code = USAGE_ERROR
			exitStatus.Message = fmt.Sprintf("Bad parameter value: approximation_size: %v.  Expected integer >= 1.", paramVal)
			return
		}
	}
	// F tolerance
	f_tolerance := lbfgsb.f_tolerance
	if paramVal, ok = parameters["f_tol"]; ok {
		f_tolerance, ok = paramVal.(float64)
		if !ok || f_tolerance <= 0.0 {
			exitStatus.Code = USAGE_ERROR
			exitStatus.Message = fmt.Sprintf("Bad parameter value: f_tol: %v.  Expected float >= 0.", paramVal)
			return
		}
	}
	// G tolerance
	g_tolerance := lbfgsb.g_tolerance
	if paramVal, ok = parameters["g_tol"]; ok {
		g_tolerance, ok = paramVal.(float64)
		if !ok || g_tolerance <= 0.0 {
			exitStatus.Code = USAGE_ERROR
			exitStatus.Message = fmt.Sprintf("Bad parameter value: g_tol: %v.  Expected float >= 0.", paramVal)
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
	bounds_control := make([]int32, dim)
	if lbfgsb.lower_bounds != nil {
		for index, bound := range lbfgsb.lower_bounds {
			if !math.IsNaN(bound) && !math.IsInf(bound, -1) {
				bounds_control[index] = 1
			}
		}
	}
	if lbfgsb.upper_bounds != nil {
		for index, bound := range lbfgsb.upper_bounds {
			if !math.IsNaN(bound) && !math.IsInf(bound, -1) {
				// Map 0 -> 3, 1 -> 2
				bounds_control[index] = 3 - bounds_control[index]
			}
		}
	}

	// Set up callback
	callbackData := &callbackData{objective: objective}
	callback_data_c := unsafe.Pointer(callbackData)

	// Allocate arrays for return value
	minimum.X = make([]float64, dim)
	minimum.G = make([]float64, dim)

	// Convert parameters for C
	approximation_size_c := C.int(approximation_size)
	f_tolerance_c := C.double(f_tolerance)
	g_tolerance_c := C.double(g_tolerance)
	debug_c := C.int(debug)

	// Prepare buffers and arrays for C.  Avoid allocation in C land by
	// allocating compatible things in Go and passing their addresses.
	// The following arrays may not be iteroperably type-safe but this
	// is how they did it on the Cgo page: http://golang.org/cmd/cgo/.
	// (One could always allocate slices of C types, pass those, and
	// then copy out and convert the contents on return.)
	var bounds_control_c *C.int = (*C.int)(&bounds_control[0])
	var lower_bounds_c *C.double = (*C.double)(&lbfgsb.lower_bounds[0])
	var upper_bounds_c *C.double = (*C.double)(&lbfgsb.upper_bounds[0])
	var x0_c *C.double = (*C.double)(&initialPoint[0])
	var min_x_c *C.double = (*C.double)(&minimum.X[0])
	var min_f_c *C.double = (*C.double)(&minimum.F)
	var min_g_c *C.double = (*C.double)(&minimum.G[0])
	// Status message
	status_message_length_c := C.int(buffer_size)
	var status_message_buffer [buffer_size]C.char
	status_message_c := (*C.char)(&status_message_buffer[0])

	// Call the actual L-BFGS-B procedure
	status_code_c := C.lbfgsb_minimize_c(
		callback_data_c, dim_c,
		bounds_control_c, lower_bounds_c, upper_bounds_c,
		approximation_size_c, f_tolerance_c, g_tolerance_c,
		x0_c, min_x_c, min_f_c, min_g_c,
		status_message_length_c, status_message_c, debug_c,
	)

	// Convert outputs
	// Exit status codes match between ExitStatusCode and the C enum
	exitStatus.Code = ExitStatusCode(status_code_c)
	exitStatus.Message = C.GoStringN(status_message_c, status_message_length_c)
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
	callback_data_c unsafe.Pointer,
	status_message_length C.int, status_message *C.char) (
		status_code_c C.int) {

	var point []float64

	// Convert inputs
	dim := int(dim_c)
	wrapCArrayAsGoSlice_Float64(point_c, dim, &point)
	cbdata := (*callbackData)(callback_data_c)

	// Evaluate the objective function
	value := cbdata.objective.Evaluate(point)

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
	callback_data_c unsafe.Pointer,
	status_message_length C.int, status_message *C.char) (
		status_code_c C.int) {

	var point, gradient, grad_ret []float64

	// Convert inputs
	dim := int(dim_c)
	wrapCArrayAsGoSlice_Float64(point_c, dim, &point)
	cbdata := (*callbackData)(callback_data_c)

	// Evaluate the gradient of the objective function
	grad_ret = cbdata.objective.EvaluateGradient(point)

	// Convert outputs
	wrapCArrayAsGoSlice_Float64(gradient_c, dim, &gradient)
	copy(gradient, grad_ret)

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
