// Copyright (c) 2013 Aubrey Barnard.  This is free software.  See
// LICENSE.txt for details.

// C declarations for wrapper of Fortran gradient descent implementation
// bound (exposed) to C

#ifndef __GD_C_H__
#define __GD_C_H__

typedef int (*objective_function_type)(int, double[], double*, void*);
typedef int (*objective_gradient_type)(int, double[], double[], void*);

int gradient_descent_f(objective_function_type, objective_gradient_type,
                       void*, double, int, int, double[],
                       double[], double*, double[]);

#endif
