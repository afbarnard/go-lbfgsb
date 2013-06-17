// C declarations for code used by Go to interface with the L-BFGS-B
// Fortran library.

#ifndef __LBFGSB_GO_INTERFACE_H__
#define __LBFGSB_GO_INTERFACE_H__

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
 int error_message_length,
 char *error_message,
 int debug
 );

#endif
