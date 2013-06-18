! Copyright (c) 2013 Aubrey Barnard.  This is free software.  See
! LICENSE.txt for details.

! Program to exercise the gradient descent module and provide a baseline
! for its operation (to compare to interfaces to other languages).

! Module containing objective function and gradient
module functions
  use gd

contains

  ! Parabola (sphere) function
  function sphere(x, f) result(error_code)
    real(dp), intent(in) :: x(:)
    real(dp), intent(out) :: f
    integer :: error_code
    f = sum(x ** 2)
    !print *, "sphere(", x, ") =", f
    error_code = 0
  end function sphere

  ! Parabola (sphere) gradient
  function sphere_grad(x, g) result(error_code)
    real(dp), intent(in) :: x(:)
    real(dp), intent(out) :: g(size(x))
    integer :: error_code
    g = 2.0_dp * x
    !print *, "sphere_grad(", x, ") =", g
    error_code = 0
  end function sphere_grad

end module functions

! Program to run gradient descent on the objective and display results
program mainf
  use gd
  use functions

  ! Locals
  integer, parameter :: dim = 3
  type(point_value_gradient) :: min
  real(dp) :: x0(dim)
  integer :: error_code

  print *, "Fortran optimization interface prototype"
  print *

  ! Init
  x0 = [7_dp, -8_dp, 9_dp]

  ! Do minimization for 100 iterations
  error_code = gradient_descent(sphere, sphere_grad, x0, 1d-1, 100, min)

  ! Output result
  print *, "mainf("
  print *, "     x0:", x0
  print *, "  min%x:", min%x
  print *, "  min%g:", min%g
  print *, "  min%f:", min%f
  print *, "  error:", error_code
  print *, ")"

end program mainf
