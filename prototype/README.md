Go-Fortran Optimization Interface Prototype
===========================================


This directory contains a collection of code for experimenting with
foreign function interfaces for interoperating between Fortran, C, and
Go.  The code outlines the features and implementation of a Go-Fortran
interface and is a prototype of a Go interface for Nocedal's L-BFGS-B
code which is in FORTRAN 77.

The process was to implement a very basic optimization solver (gradient
descent) in Fortran and then develop interfaces for it in both C and Go.
There are three main programs, one in each language, which call the
Fortran optimizer with the same parameters and objective function.  The
objective function and its gradient are implemented in each calling
language alongside the main program.  The Fortran main serves as an
operational baseline because it is a single-language program.


How It Works
------------

The design of the optimization procedure is a function that takes the
objective and gradient functions as input and returns the minimum (and
any error).  The objective function and its gradient are provided as
function arguments to the optimization procedure along with any
parameters controlling the process.  (In contrast, the L-BFGS-B
implementation passes state up and down the call stack to accomplish
what could be done with function callbacks in a modern API.)

The call structures with some explanation are below.  Indentations are
stack levels.  The right arrows mean "calls"; the left arrows mean
"returns".  Arguments to calls are figurative and are only included to
indicate information flow.  Square brackets "[]" identify locations.
The remaining figurative pseudocode hopefully will be clear.


Fortran
-------

This is the straightforward single-language setup passing the objective
function and gradient as procedure arguments.

    [Fortran (F): program mainf]
    ...
    -> [F: module (mod) gd] gradient_descent(obj_func, obj_grad, ...)

         -> [F: gd] obj_func(pt) == [F: mod functions] sphere(pt)  # objective function
           <-(function value, error) [exit F functions to F gd]

         -> [F: gd] obj_grad(pt) == [F: functions] sphere_grad(pt)  # objective gradient
           <-(gradient, error) [exit F functions to F gd]

         repeat until optimized or error
      <-(minimum, error) [exit F gd to F mainf]
    ...


C
-

This is the structure of C calling Fortran calling C.  Intermediary
callbacks and storage are necessary to adapt the Fortran procedure
signatures to the C callbacks.

    [C: mainc: main]
    ...
    -> [F: mod gd_c] gradient_descent_f(obj_func, obj_grad, data, ...)
         store callbacks and data for later

         -> [F: mod gd] gradient_descent(func, grad, ...)

              -> [F: gd] func() == [F: gd_c] func_wrapper()  # intermediary
                   access callback cb_func and data cb_data
                   -> [F: gd_c] cb_func(cb_data, ...) == [C: mainc] objfunc_sphere()  # objective function
                     <-(function value, error) [exit C mainc to F gd_c]
                <-(function value, error) [exit F gd_c to F gd]

              -> [F: gd] grad() == [F: gd_c] grad_wrapper()  # intermediary
                   access callback cb_grad and data cb_data
                   -> [F: gd_c] cb_grad(cb_data, ...) == [C: mainc] objgrad_sphere()  # objective gradient
                     <-(gradient, error) [exit C mainc to F gd_c]
                <-(gradient, error) [exit F gd_c to F gd]

              repeat until optimized or error
           <-(minimum, error) [exit F gd to F gd_c]
      <-(minimum, error) [exit F gd_c to C main]
    ...


Go
--

This is the structure of Go calling C calling Fortran calling Go that is
exported to C calling Go.  The C layer in the middle is necessary to use
Go functions (exported to C) in C function pointers.  Here there are
also intermediary callbacks to adapt signatures from Fortran to C to Go.
There is also an extra outer layer that models importing a package.

    [Go: maing: main]
    ...
    -> [Go: gd] GradientDescent(objective_object, ...)
         construct go_cb_data to contain objective_object
         -> [C: gd_go_inter] gradient_descent_c(go_cb_data, ...)

              -> [F: mod gd_c] gradient_descent_f(go_func_cb, go_grad_cb, go_cb_data, ...)
                   store callbacks and data for later

                   -> [F: mod gd] gradient_descent(func, grad, ...)

                        -> [F: gd] func() == [F: gd_c] func_wrapper()  # intermediary
                             access callback cb_func and data go_cb_data
                             -> [F: gd_c] cb_func(go_cb_data, ...) == [Go/C: gd] go_func_cb()  # intermediary
                                  get go_func from go_cb_data
                                  -> [Go: gd] go_func() == [Go: maing] Sphere.Evaluate()  # objective function
                                    <-(function value, error) [exit Go maing to Go gd]
                               <-(function value, error) [exit Go gd to F gd_c]
                          <-(function value, error) [exit F gd_c to F gd]

                        -> [F: gd] grad() == [F: gd_c] grad_wrapper()  # intermediary
                             access callback cb_grad and data go_cb_data
                             -> [F: gd_c] cb_grad(go_cb_data, ...) == [Go/C: gd] go_grad_cb()  # intermediary
                                  get go_grad from go_cb_data
                                  -> [Go: gd] go_grad() == [Go: maing] Sphere.EvaluateGradient()  # objective gradient
                                    <-(gradient, error) [exit Go maing to Go gd]
                               <-(gradient, error) [exit Go gd to F gd_c]
                          <-(gradient, error) [exit F gd_c to F gd]

                        repeat until optimized or error
                     <-(minimum, error) [exit F gd to F gd_c]
                <-(minimum, error) [exit F gd_c to C gd_go_inter]
           <-(minimum, error) [exit C gd_go_inter to Go gd]
      <-(minimum, error) [exit Go gd to Go maing]
    ...

The above structure of the prototype is more complicated than a direct
Go-Fortran interface because a direct Fortran API could be tailored to
accommodate Go/C.  The prototype Fortran API was designed as if it only
needed to consider Fortran.  Therefore, the API only accepts Fortran
procedures and has no facility for callback data.  These limitations
necessitate the intermediary adapters but such limitations would not
occur in a tailored API.  The complexity of a tailored and integrated
Go-Fortran API should be similar to the C-Fortran one above.


Copyright (c) 2013 Aubrey Barnard.  This is free software.  See
LICENSE.txt for details.
