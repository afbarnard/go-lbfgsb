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
	"reflect"
	"unsafe"
)

// Lbfgsb provides the functionality of the Fortran L-BFGS-B optimizer
// as a Go object.
type Lbfgsb struct {
	// TODO
}

// Minimize optimizes the given objective using the L-BFGS-B algorithm.
// Implements OptimizationFunctionMinimizer.Minimize.
func (bfgs Lbfgsb) Minimize(
	objective FunctionWithGradient,
	initialPoint []float64,
	parameters map[string]interface{}) (
		minimum PointValueGradient,
		exitStatus ExitStatus) {

	// TODO process the parameters

	dim := len(initialPoint)

	// Set up callback
	callbackData := &callbackData{objective: objective}
	// Allocate return value
	minimum = PointValueGradient{
		X: make([]float64, dim),
		G: make([]float64, dim),
	}
	// Allocate arrays for bounds
	bounds_control := make([]int32, dim)
	lower_bounds := make([]float64, dim)
	upper_bounds := make([]float64, dim)

	// Convert inputs TODO!
	// Use some dummy values just to get things connected
	callback_data_c := unsafe.Pointer(callbackData)
	dim_c := C.int(dim)
	approximation_size_c := C.int(5)
	f_tolerance_c := C.double(1e-6)
	g_tolerance_c := C.double(1e-6)
	error_message_length_c := C.int(100)
	var dummy_message [101]C.char
	error_message_c := (*C.char)(&dummy_message[0])
	debug_c := C.int(0)
	// The following arrays may not be iteroperably type-safe but this
	// is how they did it on the Cgo page: http://golang.org/cmd/cgo/
	var bounds_control_c *C.int = (*C.int)(&bounds_control[0])
	var lower_bounds_c *C.double = (*C.double)(&lower_bounds[0])
	var upper_bounds_c *C.double = (*C.double)(&upper_bounds[0])
	var x0_c *C.double = (*C.double)(&initialPoint[0])
	var min_x_c *C.double = (*C.double)(&minimum.X[0])
	var min_f_c *C.double = (*C.double)(&minimum.F)
	var min_g_c *C.double = (*C.double)(&minimum.G[0])

	// Call the actual L-BFGS-B procedure
	error_code_c := C.lbfgsb_minimize_c(
		callback_data_c, dim_c,
		bounds_control_c, lower_bounds_c, upper_bounds_c,
		approximation_size_c, f_tolerance_c, g_tolerance_c,
		x0_c, min_x_c, min_f_c, min_g_c,
		error_message_length_c, error_message_c, debug_c,
	)

	// Convert outputs
	// TODO properly translate exit status using enumerated error codes
	error_code := int(error_code_c)
	if error_code != 0 {
		panic(fmt.Errorf("Gradient descent failed with error code: %d", error_code))
	}

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
	error_message_length C.int, error_message *C.char) (
		error_code_c C.int) {

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
	error_message_length C.int, error_message *C.char) (
		error_code_c C.int) {

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
