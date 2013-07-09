Go L-BFGS-B
===========


Description
-----------

Go L-BFGS-B is software for solving numerical optimization problems
using the limited-memory (L) Broyden-Fletcher-Goldfarb-Shanno (BFGS)
algorithm with bounds constraints (B).  L-BFGS-B is a deterministic,
gradient-based algorithm for finding local minima of smooth, continuous
objective functions subject to bounds on each variable.  The bounds are
optional, so this software also solves unconstrained problems.  L-BFGS-B
is accurate and efficient for problems of 1000s of variables.  There is
more information about [L-BFGS-B on
Wikipedia](http://en.wikipedia.org/wiki/L-BFGS) or in your favorite
optimization textbook.

This software provides modern, intuitive interfaces to the [L-BFGS-B
Fortran
software](http://users.eecs.northwestern.edu/~nocedal/software.html) by
the authors of the algorithm, Jorge Nocedal et al.  Interfaces are
provided for Go (Golang), C, and Fortran 2003.


License
-------

Go L-BFGS-B is free, open source software.  You are permitted to copy
and use it however you wish.  It is released under the BSD 2-Clause
License.  See the file `LICENSE.txt` in your distribution (or [on
GitHub](https://github.com/afbarnard/go-lbfgsb/blob/master/LICENSE.txt))
for details.

The original Fortran code is released under the BSD 3-Clause License
(see `lbfgsb/License.txt` in your distribution or [on
GitHub](https://github.com/afbarnard/go-lbfgsb/blob/master/lbfgsb/License.txt)).
Since I have not modified the Fortran code, I have decided just to
incorporate it rather than re-license it.


Features
--------

Go L-BFGS-B is version 0.1.0.

This software is young, but has already been useful and reliable in a
machine learning application.

* Simple API allowing Go programs to easily do efficient, accurate
  optimization.  You only need to provide functions (or an object) that
  evaluate the objective function value and gradient.

* Bounds constraints allow typical, simple inequality (box) constraints
  on individual variables without needing a more specialized algorithm.

* Incorporates recent improvements and corrections to L-BFGS-B
  algorithm.  (Uses L-BFGS-B version 3.0 from March, 2011.)

* Customizable logging.


Future Features
---------------

* Modern (Fortran 2003) API for original FORTRAN 77 optimizer.

* C API for original FORTRAN 77 optimizer.

* Customizable termination conditions.


Go-Fortran Interface
--------------------

The architecture of this software has two main pieces: a library with a
C API (but written in modern Fortran), and a Go package providing the
functionality of the library as a Go API.  The library is designed to be
used from both Fortran and C, and so could work with other languages
besides Go.

I have endeavored to make this a good Go-Fortran interface example as
well as a useful package.  Accordingly, I have included my experimental
interface prototype code to be used for learning and experimenting with
Cgo and interfaces between Go and C (or Fortran).  However, let me know
about (or contribute!) possible improvements.


Requirements
------------

Building this software requires:

* Fortran 2003 compiler with support for procedure pointers, such as GCC
  4.4.6 or later (gfortran)

* Go 1.0 or later

* Standard development tools: make, ar, ld.


Download, Build, Install
------------------------

Conveniently, you can use the Go tools to download, build, and install
this package, but it is not quite fully automatic: there is an
intervening step to compile the Fortran code.  This is needed because
the Go compilers do not know Fortran.

1. Download.  Using `go get` requires `git`.  If you want to download
   the latest code after having downloaded this package previously, add
   the update flag (`-u`) to the command.  The `-d` flag tells Go to
   only download the code and not build or install the package.

   ```shell
   [go-wrkspc]$ go get -d github.com/afbarnard/go-lbfgsb
   ```

2. Build Fortran.  Change to the directory containing the downloaded
   package and run make.  The workspace that Go uses to download and
   install packages is the first workspace in your GOPATH, or the Go
   installation directory if GOPATH is empty.

   ```shell
   [go-wrkspc]$ echo $GOPATH
   /home/go-pkgs:/home/go-wrkspc
   [go-wrkspc]$ cd ~/go-pkgs/src/github.com/afbarnard/go-lbfgsb
   [go-lbfgsb]$ make
   [go-lbfgsb]$ cd ~/go-wrkspc
   ```

3. Build, install Go.  Run `go get` again to complete the Go
   installation.  This builds and installs the package.

   ```shell
   [go-wrkspc]$ go get github.com/afbarnard/go-lbfgsb
   ```

4. Use.  Import the `go-lbfgsb` package into your Go program.  Build
   your program normally.

   ```go
   import lbfgsb "github.com/afbarnard/go-lbfgsb"
   ```

   ```shell
   [main-prog]$ go build
   ```


Alternative Download, Install, Build
------------------------------------

If you do not want to use `go get`, you can download this package
manually and make it part of your Go workspace.

1. Download.  Get a project release archive from GitHub.

2. Install.  Extract the release archive into your Go workspace.

   ```shell
   [go-wrkspc]$ tar -zxf ~/downloads/go-lbfsgb-1.2.3.tar.gz --directory optim
   ```

3. Build.  Run make in the package directory.

   ```shell
   [go-wrkspc]$ cd optim/go-lbfgsb-1.2.3
   [go-lbfgsb-1.2.3]$ make
   [go-lbfgsb-1.2.3]$ cd ~/go-wrkspc
   ```

4. Use.  Import the package as a local package and build normally.

   ```go
   import lbfgsb "optim/go-lbfgsb-1.2.3"
   ```

   ```shell
   [main-prog]$ go build
   ```


Contact, Contribute
-------------------

* [Aubrey Barnard](https://github.com/afbarnard)

Contributions of all sorts, patches, bugs, issues, questions, etc. are
welcome, but are most welcome after due diligence.  Contact the author
by creating a [new
issue](https://github.com/afbarnard/go-lbfgsb/issues/new).  Contribute
to the project by forking it, hacking, and issuing pull requests.


Copyright (c) 2013 Aubrey Barnard.  This is free software.  See
LICENSE.txt for details.
