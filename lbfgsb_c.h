// Copyright (c) 2013 Aubrey Barnard.  This is free software.  See
// LICENSE.txt for details.

// C declarations for Fortran types and procedures bound (exposed) to C.

#ifndef __LBFGSB_C_H__
#define __LBFGSB_C_H__

// TODO declare constants enumerated in Fortran code

// Signature of objective function callback.  Matches 'function
// objective_function_c', explained in Fortran module.
typedef int (*lbfgsb_objective_function_type)
(
 int dim,
 double *point,
 double *objective_function_value,
 void *callback_data,
 char *status_message,
 int status_message_length
 );

// Signature of objective gradient callback.  Matches 'function
// objective_gradient_c', explained in Fortran module.
typedef int (*lbfgsb_objective_gradient_type)
(
 int dim,
 double *point,
 double *objective_function_gradient,
 void *callback_data,
 char *status_message,
 int status_message_length
 );

// Signature of logging function callback.  Matches 'function
// log_function_c', explained in Fortran module.
typedef int (*lbfgsb_log_function_type)
(
 void *callback_data,
 int iteration,
 int fg_evals,
 int fg_evals_total,
 double step_length,
 int dim,
 double *x,
 double f,
 double *g,
 double f_delta,
 double f_delta_bound,
 double g_norm,
 double g_norm_bound
 );

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

 // Parameters
 int approximation_size,
 double f_tolerance,
 double g_tolerance,

 // Input
 double *initial_point,

 // Result
 double *min_x,
 double *min_f,
 double *min_g,
 int *iters,
 int *evals,

 // Printing, logging
 int fortran_print_control,
 lbfgsb_log_function_type log_function,
 void *log_function_callback_data,

 // Exit status
 char *status_message,
 int status_message_length
 );

#endif
