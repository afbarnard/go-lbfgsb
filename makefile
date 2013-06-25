# Copyright (c) 2013 Aubrey Barnard.  This is free software.  See
# LICENSE.txt for details.

# Builds the Fortran L-BFGS-B code into a library for use by Go

# Setup for make
.PHONY: all clean

# Configuration and options

# Compiler (I'm only worrying about GCC for now since it supports Fortran, C, and Go.)
compiler := gcc
# General compiler options, e.g. -O -g
compile_options := -g
# General compiler warnings
compile_warnings := -Wall -Wextra
# Fortran compiler options.  Disallow implicit variables.
compile_options_fortran := -fimplicit-none

# Compilation

# Build the library by default
all: liblbfgsb.a

# FORTRAN 77 compilation.  The original L-BFGS-B makefile adds bounds
# checking code with '-fbounds-check', so do that here in case it's
# important.  I have checked the code for all the warnings about float
# (in)equality and unused parameters.  They seem to be fine, so I have
# disabled those warnings (which were only enabled by -Wextra anyway).
%.o: %.f
	$(compiler) $(compile_options) $(compile_warnings) $(compile_options_fortran) -fbounds-check -std=legacy -Wno-compare-reals -Wno-unused-parameter -c $< -o $@

# Fortran 2003 compilation
%.o: %.f03
	$(compiler) $(compile_options) $(compile_warnings) $(compile_options_fortran) -fbounds-check -std=f2003 -c $< -o $@

# Dependencies

# Original L-BFGS-B
lbfgsb/lbfgsb.o: lbfgsb/blas.o lbfgsb/linpack.o lbfgsb/timer.o
# The Linpack parts do not appear to depend on each other

# Go-Fortran L-BFGS-B interface
lbfgsb.o: lbfgsb/lbfgsb.o
lbfgsb_c.o: lbfgsb.o

# Libraries and executables

# Library
liblbfgsb.a: lbfgsb_c.o lbfgsb.o lbfgsb/lbfgsb.o lbfgsb/blas.o lbfgsb/linpack.o lbfgsb/timer.o
	ar cr $@ $^

# Commands

# Delete derived
clean:
	@rm -f *.o *.mod *~ lbfgsb/*.o liblbfgsb.a
