# Copyright (c) 2013 Aubrey Barnard.  This is free software.  See
# LICENSE.txt for details.

# Makes the Fortran and C parts of this experimental package.  Can also
# build the Go bits for convenience.

.PHONY: all clean

all: mainf mainc maing libgd.a

# Fortran compilation
%.o: %.f03
	gcc -Wall -std=f2003 -fimplicit-none -c $<

# C compilation
%.o: %.c
	gcc -Wall -std=c99 -c $<

# Fortran module static library
libgd.a: gd.o gd_c.o
	ar cr $@ $^

# Fortran main
mainf: gd.o mainf.o
	gfortran -Wall -o $@ $^

# C main
mainc: gd.o gd_c.o mainc.o
	gcc -Wall -o $@ $^ -lgfortran

# Go main
maing: libgd.a gd.go maing.go
	go build -o $@

# Additional dependencies
mainf.o: gd.o
mainc.o: gd_c.h
gd_c.o: gd.o

# Delete derived
clean:
	@rm -f mainf mainc maing lib*.a *.o *.mod *~
