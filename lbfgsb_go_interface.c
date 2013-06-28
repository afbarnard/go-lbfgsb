// Copyright (c) 2013 Aubrey Barnard.  This is free software.  See
// LICENSE.txt for details.

// C wrapper for Fortran L-BFGS-B implementation.  This is
// needed solely to pass the addresses of the exported Go callbacks to C
// because Go and C function pointers are not interoperable.  (Otherwise
// this wouldn't be needed as Go can directly call Fortran code that
// has been exposed to C with bind(c).)

#include <stddef.h>

#include "lbfgsb_go_interface.h"
#include "lbfgsb_c.h"
#include "_cgo_export.h"

int lbfgsb_minimize_c
(
 void *callback_data,
 int dim,
 int *bounds_control,
 double *lower_bounds,
 double *upper_bounds,
 int approximation_size,
 double f_tolerance,
 double g_tolerance,
 double *initial_point,
 double *min_x,
 double *min_f,
 double *min_g,
 int *iters,
 int *evals,
 int fortran_print_control,
 int do_logging,
 void *log_function_callback_data,
 char *status_message,
 int status_message_length
 )
{
  // Only pass the logging function if asked
  lbfgsb_log_function_type log_function_pointer = NULL;
  if (do_logging) {
    log_function_pointer = go_log_function_callback;
  }

  // Call the Fortran code to do the minimization
  return lbfgsb_minimize
    (
     go_objective_function_callback,
     go_objective_gradient_callback,
     callback_data,
     dim,
     bounds_control,
     lower_bounds,
     upper_bounds,
     approximation_size,
     f_tolerance,
     g_tolerance,
     initial_point,
     min_x,
     min_f,
     min_g,
     iters,
     evals,
     fortran_print_control,
     log_function_pointer,
     log_function_callback_data,
     status_message,
     status_message_length
     );
}
