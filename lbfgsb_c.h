// Copyright (c) 2013 Aubrey Barnard.  This is free software.  See
// LICENSE.txt for details.

// C declarations for Fortran types and procedures bound (exposed) to C.

#ifndef __LBFGSB_C_H__
#define __LBFGSB_C_H__

// TODO declare constants enumerated in Fortran code

// Signature of objective function callback.  Matches 'function
// objective_function_c', explained in Fortran module.
typedef int (*lbfgsb_objective_function_type)
(int dim,
 double *point,
 double *objective_function_value,
 void *callback_data,
 int error_message_length,
 char *error_message);

// Signature of objective gradient callback.  Matches 'function
// objective_gradient_c', explained in Fortran module.
typedef int (*lbfgsb_objective_gradient_type)
(int dim,
 double *point,
 double *objective_function_gradient,
 void *callback_data,
 int error_message_length,
 char *error_message);

// Signature of L-BFGS-B minimizer.  Matches 'function lbfgsb_minimize',
// explained in Fortran module.
int lbfgsb_minimize
(
 // Callbacks for objective function and gradient
 lbfgsb_objective_function_type objective_function,
 lbfgsb_objective_gradient_type objective_gradient,
 void *callback_data,

 // Dimensionality, number of variables
 int dim,

 // Bounds
 int *bounds_control,
 double *lower_bounds,
 double *upper_bounds,

 // Minimization parameters
 int approximation_size,
 double f_tolerance,
 double g_tolerance,

 // Input
 double *initial_point,

 // Result
 double *min_x,
 double *min_f,
 double *min_g,

 // Error, debug
 int error_message_length,
 char *error_message,
 int debug
);

#endif
