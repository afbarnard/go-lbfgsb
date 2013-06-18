// Copyright (c) 2013 Aubrey Barnard.  This is free software.  See
// LICENSE.txt for details.

// Go-Fortran gradient descent package (Go-Fortran optimization
// interface prototype)

package main

// #cgo LDFLAGS: -L. -lgd -lm -lgfortran
// #include "gd_go_inter.h"
import "C"

import (
	"fmt"
	"reflect"
	"unsafe"
)

// A function f: R**n -> R.
type Function interface {
	// Returns the value of the function at a point.
	Evaluate(point []float64) float64
}

// A function f: R**n -> R and its gradient f': R**n -> R**n.
type FunctionWithGradient interface {
	Function
	// Returns the gradient of the function at the given point.
	EvaluateGradient(point []float64) []float64
}

// A point and its function value.
type PointValue struct {
	X []float64
	F float64
}

// A point, its function value, and its gradient.
type PointValueGradient struct {
	PointValue
	G []float64
}

// N-dimensional numerical minimization.  Finds a local minimum.
type FunctionMinimizer interface {
	// Minimizes the given function starting from the given point.  The
	// function may be a Function or a FunctionWithGradient.
	Minimize(function Function, initialPoint []float64) (
		result PointValue, success bool, iterations int)
}

// Implements a simplified version of gradient descent by calling
// Fortran implementation (via C).  This gradient descent runs for a
// fixed number of iterations with a fixed step size just to make things
// easy.
func GradientDescent(
	function FunctionWithGradient,
	initialPoint []float64,
	stepSize float64,
	maxIterations int) (
		result *PointValueGradient, err error){

	// Create data
	callbackData := &CallbackData{objective: function}
	min_x := make([]float64, len(initialPoint))
	min_g := make([]float64, len(initialPoint))
	var min_f float64

	// Convert inputs
	callback_data_c := unsafe.Pointer(callbackData)
	stepsize_c := C.double(stepSize)
	iters_c := C.int(maxIterations)
	dim_c := C.int(len(initialPoint))
	// The following arrays may not be iteroperably type-safe but this
	// is how they did it on the Cgo page: http://golang.org/cmd/cgo/
	var x0_c *C.double = (*C.double)(&initialPoint[0])
	var min_x_c *C.double = (*C.double)(&min_x[0])
	var min_f_c *C.double = (*C.double)(&min_f)
	var min_g_c *C.double = (*C.double)(&min_g[0])

	// Call the actual gradient descent procedure
	error_code_c := C.gradient_descent_c(callback_data_c,
		stepsize_c, iters_c, dim_c, x0_c, min_x_c, min_f_c, min_g_c)

	// Convert outputs
	error_code := int(error_code_c)
	if error_code == 0 {
		result = &PointValueGradient{
			PointValue: PointValue{
				X: min_x,
				F: min_f,
			},
			G: min_g,
		}
	} else {
		err = fmt.Errorf("Gradient descent failed with error code: %d", error_code)
	}

	return
}

// Container for actual objective callback
type CallbackData struct {
	objective FunctionWithGradient
}

// Adapter between C callback and Go callback (for objective function).
// Exported to C for use as a function pointer.  Must match the
// signature of objective_function_type in gd_c.h.
//
//export go_objective_function_callback
func go_objective_function_callback(
	dim_c C.int, point_c, value_c *C.double,
	callback_data_c unsafe.Pointer) (error_code_c C.int) {

	var point []float64

	//fmt.Printf("go_objective_function_callback(\n")
	//fmt.Printf("  dim: %v\n", dim_c)
	//fmt.Printf("  point[0]: %v; &point[0]: %p\n", *point_c, point_c)
	//fmt.Printf("  point: %v;\n", point)
	//fmt.Printf("  value: %v; &value: %p\n", *value_c, value_c)
	//fmt.Printf("  cbdata: %v; &cbdata: %p\n", (*CallbackData)(callback_data_c), callback_data_c)
	//fmt.Printf("  error: %v;\n", error_code_c)
	//fmt.Printf(")\n")

	// Convert inputs
	dim := int(dim_c)
	WrapCArrayAsGoSlice_Float64(point_c, dim, &point)
	cbdata := (*CallbackData)(callback_data_c)

	// Evaluate the objective function
	value := cbdata.objective.Evaluate(point)

	// Convert outputs
	*value_c = C.double(value)

	//fmt.Printf("go_objective_function_callback(\n")
	//fmt.Printf("  dim: %v\n", dim_c)
	//fmt.Printf("  point[0]: %v; &point[0]: %p\n", *point_c, point_c)
	//fmt.Printf("  point: %v;\n", point)
	//fmt.Printf("  value: %v; &value: %p\n", *value_c, value_c)
	//fmt.Printf("  cbdata: %v; &cbdata: %p\n", (*CallbackData)(callback_data_c), callback_data_c)
	//fmt.Printf("  error: %v;\n", error_code_c)
	//fmt.Printf(")\n")

	return
}

// Adapter between C callback and Go callback (for objective gradient).
// Exported to C for use as a function pointer.  Must match the
// signature of objective_gradient_type in gd_c.h.
//
//export go_objective_gradient_callback
func go_objective_gradient_callback(
	dim_c C.int, point_c, gradient_c *C.double,
	callback_data_c unsafe.Pointer) (error_code_c C.int) {

	var point, gradient, grad_ret []float64

	//fmt.Printf("go_objective_gradient_callback(\n")
	//fmt.Printf("  dim: %v\n", dim_c)
	//fmt.Printf("  point[0]: %v; &point[0]: %p\n", *point_c, point_c)
	//fmt.Printf("  point: %v;\n", point)
	//fmt.Printf("  gradient[0]: %v; &gradient[0]: %p\n", *gradient_c, gradient_c)
	//fmt.Printf("  gradient: %v;\n", gradient)
	//fmt.Printf("  grad_ret: %v;\n", grad_ret)
	//fmt.Printf("  cbdata: %v; &cbdata: %p\n", (*CallbackData)(callback_data_c), callback_data_c)
	//fmt.Printf("  error: %v;\n", error_code_c)
	//fmt.Printf(")\n")

	// Convert inputs
	dim := int(dim_c)
	WrapCArrayAsGoSlice_Float64(point_c, dim, &point)
	cbdata := (*CallbackData)(callback_data_c)

	// Evaluate the gradient of the objective function
	grad_ret = cbdata.objective.EvaluateGradient(point)

	// Convert outputs
	WrapCArrayAsGoSlice_Float64(gradient_c, dim, &gradient)
	copy(gradient, grad_ret)

	//fmt.Printf("go_objective_gradient_callback(\n")
	//fmt.Printf("  dim: %v\n", dim_c)
	//fmt.Printf("  point[0]: %v; &point[0]: %p\n", *point_c, point_c)
	//fmt.Printf("  point: %v;\n", point)
	//fmt.Printf("  gradient[0]: %v; &gradient[0]: %p\n", *gradient_c, gradient_c)
	//fmt.Printf("  gradient: %v;\n", gradient)
	//fmt.Printf("  grad_ret: %v;\n", grad_ret)
	//fmt.Printf("  cbdata: %v; &cbdata: %p\n", (*CallbackData)(callback_data_c), callback_data_c)
	//fmt.Printf("  error: %v;\n", error_code_c)
	//fmt.Printf(")\n")

	return
}

// Allows a C array to be treated as a Go slice.  Based on
// https://code.google.com/p/go-wiki/wiki/cgo.  This only works if the
// Go and C types happen to be interoperable (binary compatible).
func WrapCArrayAsGoSlice_Float64(array *C.double, length int,
	slice *[]float64) {

	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(slice))
	sliceHeader.Cap = length
	sliceHeader.Len = length
	sliceHeader.Data = uintptr(unsafe.Pointer(array))
}

// Converts a C array to a Go slice by converting each array element
func ConvertCArrayToGoSlice_Float64(array *C.double, length int,
	slice []float64) ([]float64) {

	// Allocate a new slice if needed
	if slice == nil || len(slice) != length {
		slice = make([]float64, length)
	}
	// Copy the data, converting each element.  Do array indexing by
	// hand. :(
	var dummy C.double
	base := uintptr(unsafe.Pointer(array))
	stride := unsafe.Sizeof(dummy)
	var address uintptr
	for i := 0; i < length; i++ {
		address = base + uintptr(i) * stride
		slice[i] = float64(*(*C.double)(unsafe.Pointer(address)))
	}
	return slice
}
