Contents
========


Directories
-----------

* `lbfgsb`: The FORTRAN 77 code from Nocedal (et al) that implements
  L-BFGS-B, in its original form.  I have included only the files from
  their distribution that are necessary to build and run L-BFGS-B for
  this project (although I have retained their README).  For their full
  distribution, which includes papers on the algorithms and code, sample
  programs, sample outputs, etc., visit
  http://users.eecs.northwestern.edu/~nocedal/lbfgsb.html.

  Note on L-BFGS-B License

  The L-BFGS-B distribution comes with a blank BSD 3-Clause License
  template.  I filled in the template based on the stated authors and
  year on the web page above.  Hopefully it will match their intentions
  but the license is not original to their code.

* `prototype`: Code for experimenting with foreign function interfaces
  between Fortran, C, and Go.  Use it as a seed for your own
  experiments!


Files
-----

Files are either part of the L-BFGS-B library or the Go package.

* `lbfgsb.f03`: Fortran 2003 interface to L-BFGS-B FORTRAN 77 code.
  This and the code in `lbfgsb/` are the core of the library.

* `lbfgsb_c.*`: C declarations and Fortran 2003 code for library C API.

* `*.go`: L-BFGS-B Go package.

* `lbfgsb_go_interface.*`: C files for Go package.


Copyright (c) 2013 Aubrey Barnard.  This is free software.  See
LICENSE.txt for details.
