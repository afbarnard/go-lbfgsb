! Copyright (c) 2013 Aubrey Barnard.  This is free software.  See
! LICENSE.txt for details.

! Simple implementation of gradient descent for Go-Fortran-Go interface
! prototyping.

module gd
  implicit none
  private

  ! Public constants, variables
  public dp
  ! Public types
  public point_value_gradient
  ! Public procedures
  public gradient_descent
  ! Should the interfaces be public?  Doesn't seem to matter.

  ! Double precision kind for reals
  integer, parameter :: dp = kind(0d0)

  ! Information about a point important for optimization: (x,f,g)
  type :: point_value_gradient
     real(dp) :: f
     real(dp), allocatable :: x(:), g(:)
  end type point_value_gradient

  abstract interface
     ! What an objective function looks like (f: R**n -> R)
     function func_rn_r(x, f) result(error_code)
       import :: dp
       implicit none
       real(dp), intent(in) :: x(:)
       real(dp), intent(out) :: f
       integer :: error_code
     end function func_rn_r

     ! What the gradient of an objective function looks like (f: R**n -> R**n)
     function func_rn_rn(x, g) result(error_code)
       import :: dp
       implicit none
       real(dp), intent(in) :: x(:)
       real(dp), intent(out) :: g(size(x))
       integer :: error_code
     end function func_rn_rn
  end interface

contains

  ! Version of gradient descent that only runs for a set number of
  ! iterations and uses a fixed step size (instead of a line search or a
  ! schedule, etc.).  These simplifications make for a very simple
  ! implementation that supports the goals of interfacing with different
  ! languages and calling user-specified procedures to compute the
  ! function and gradient.
  function gradient_descent(func, grad, x0, stepsize, iters, min) result(error_code)
    implicit none
    ! Signature
    procedure(func_rn_r) :: func
    procedure(func_rn_rn) :: grad
    real(dp), intent(in) :: x0(:), stepsize
    integer, intent(in) :: iters
    type(point_value_gradient), intent(out) :: min
    integer :: error_code

    ! Locals
    real(dp) :: x(size(x0)), g(size(x0)), f
    integer :: iteration

    !print *,"gradient_descent(", x0, stepsize, iters, "x:", min%x, "f:", min%f, "g:", min%g, ")"

    ! Init
    x = x0

    ! Loop to do a fixed number of gradient descent steps
    do iteration = 1, iters
       error_code = func(x, f)
       if (error_code /= 0) return
       error_code = grad(x, g)
       if (error_code /= 0) return
       !print *, iteration, "x:", x, "f:", f, "g:", g
       x = x - stepsize * g
    end do

    ! Compute and store the "minimum"
    if (.not. allocated(min%x)) allocate(min%x(size(x0)))
    if (.not. allocated(min%g)) allocate(min%g(size(x0)))
    min%x = x
    error_code = func(x, min%f)
    if (error_code /= 0) return
    error_code = grad(x, min%g)
    if (error_code /= 0) return
    !print *, iteration, "x:", min%x, "f:", min%f, "g:", min%g

    ! Success!
    error_code = 0
  end function gradient_descent

end module gd
