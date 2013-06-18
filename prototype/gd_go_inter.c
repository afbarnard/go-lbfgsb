// Copyright (c) 2013 Aubrey Barnard.  This is free software.  See
// LICENSE.txt for details.

// C wrapper for Fortran gradient descent implementation.  This is
// needed solely to pass the addresses of the exported Go callbacks to C
// because Go and C function pointers are not interoperable.  (Otherwise
// this wouldn't be needed as Go could directly call Fortran code that
// has been exposed to C with bind(c).)

#include "gd_go_inter.h"
#include "gd_c.h"
#include "_cgo_export.h"

int gradient_descent_c(void * callback_data, double stepsize, int iters,
                       int dim, double *x0,
                       double *min_x, double *min_f, double *min_g) {
  return gradient_descent_f(&go_objective_function_callback,
                            &go_objective_gradient_callback,
                            callback_data, stepsize, iters,
                            dim, x0, min_x, min_f, min_g);
}
