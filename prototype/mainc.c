// Copyright (c) 2013 Aubrey Barnard.  This is free software.  See
// LICENSE.txt for details.

// C version of Fortran main.  Runs the Fortran gradient descent from C
// with C callbacks for evaluation of objective function and gradient.

// Do not include in Go build
// +build ignore

#include <stdio.h>
#include <stdlib.h>

// Declarations for Fortran code exposed to C
#include "gd_c.h"

#define DIM 3

// Holds some dummy callback data
struct cbdata {
  char *name;
  int data;
};

// Computes the sphere (multi-dimensional parabola) function
int objfunc_sphere(int dim, double point[], double *value, void *callback_data) {
  // Calculate the function value
  double work[dim];
  for (int i = 0; i < dim; i++) {
    work[i] = point[i] * point[i];
  }
  *value = 0.0;
  for (int i = 0; i < dim; i++) {
    *value += work[i];
  }

  // Do something with the callback data
  struct cbdata *cbdata = (struct cbdata *) callback_data;
  if (cbdata->data != 987)
    return cbdata->data;

  // Return success
  return 0;
}

// Computes the gradient of the sphere function
int objgrad_sphere(int dim, double point[], double grad[], void *callback_data) {
  // Calculate the gradient value
  for (int i = 0; i < dim; i++) {
    grad[i] = 2.0 * point[i];
  }

  // Do something with the callback data
  struct cbdata *cbdata = (struct cbdata *) callback_data;
  if (cbdata->data != 987)
    return cbdata->data;

  // Return success
  return 0;
}

// Runs gradient descent on the objective and displays results
int main() {
  printf("C-Fortran optimization interface prototype\n\n");

  struct cbdata cbdata = {"callback data", 987};
  double *x0;
  double *min_x;
  double *min_g;
  double min_f = -1.0;
  int error_code = -1;

  // Allocate, populate arrays.  Works just as well with static arrays
  // but wanted to make sure it works with dynamic.
  x0 = (double *) malloc(DIM * sizeof(double));
  min_x = (double *) malloc(DIM * sizeof(double));
  min_g = (double *) malloc(DIM * sizeof(double));
  x0[0] = 7.0;
  x0[1] = -8.0;
  x0[2] = 9.0;

  //printf("mainc(\n");
  //printf("     x0: %p; [%g, %g, %g]\n", x0, x0[0], x0[1], x0[2]);
  //printf("  min_x: %p; [%g, %g, %g]\n", min_x, min_x[0], min_x[1], min_x[2]);
  //printf("  min_g: %p; [%g, %g, %g]\n", min_g, min_g[0], min_g[1], min_g[2]);
  //printf("  min_f: %g\n", min_f);
  //printf("  error: %d\n", error_code);
  //printf(")\n");

  //printf("*** Call: gradient_descent_f\n");
  error_code = gradient_descent_f(&objfunc_sphere, &objgrad_sphere,
                                  (void*) &cbdata, 1e-1, 100,
                                  DIM, x0, min_x, &min_f, min_g);
  //printf("*** Rtrn: gradient_descent_f\n");

  printf("mainc(\n");
  printf("     x0: [%g, %g, %g]\n", x0[0], x0[1], x0[2]);
  printf("  min_x: [%g, %g, %g]\n", min_x[0], min_x[1], min_x[2]);
  printf("  min_g: [%g, %g, %g]\n", min_g[0], min_g[1], min_g[2]);
  printf("  min_f: %g\n", min_f);
  printf("  error: %d\n", error_code);
  printf(")\n");

  return 0;
}
