! Copyright (c) 2013 Aubrey Barnard.  This is free software.  See
! LICENSE.txt for details.

! C interoperability for Fortran gd module.  Needs Fortran 2003 for
! ISO_C_BINDING, procedure pointers, and procedure components.
! (Procedure components require GCC 4.6 or later; other stuff works with
! GCC 4.4 but I did encounter some internal compiler errors with 4.4.)

! Module of definitions and bindings to allow the gd module to be
! (properly) used from C
module gd_c
  use, intrinsic :: iso_c_binding
  use gd
  implicit none
  private

  ! Public procedures
  public gradient_descent_cwrap

  ! Fortran passes arguments by reference, so make sure to use the
  ! 'value' attribute for C arguments passed by value
  abstract interface
     ! Signature of objective function C callback
     function objective_function(dim, point, function_value, callback_data) &
          result(error_code) bind(c)
       use, intrinsic :: iso_c_binding
       implicit none
       integer(c_int), intent(in), value :: dim
       real(c_double), intent(in) :: point(dim)
       real(c_double), intent(out) :: function_value
       type(c_ptr), intent(in), value :: callback_data
       integer(c_int) :: error_code
     end function objective_function

     ! Signature of objective gradient C callback
     function objective_gradient(dim, point, gradient, callback_data) &
          result(error_code) bind(c)
       use, intrinsic :: iso_c_binding
       implicit none
       integer(c_int), intent(in), value :: dim
       real(c_double), intent(in) :: point(dim)
       real(c_double), intent(out) :: gradient(dim)
       type(c_ptr), intent(in), value :: callback_data
       integer(c_int) :: error_code
     end function objective_gradient
  end interface

  ! Container for objective function C callback and data for
  ! communicating between gradient_descent_cwrap and func_wrapper
  type func_callback_data
     procedure(objective_function), pointer, nopass :: func
     type(c_ptr) :: data
  end type func_callback_data

  ! Container for objective gradient C callback and data for
  ! communicating between gradient_descent_cwrap and grad_wrapper
  type grad_callback_data
     procedure(objective_gradient), pointer, nopass :: grad
     type(c_ptr) :: data
  end type grad_callback_data

  ! Module variables for communicating between wrappers
  type(func_callback_data) :: f_cbdata
  type(grad_callback_data) :: g_cbdata

contains

  ! Wrapper to expose gd.gradient_descent to C as gradient_descent_f
  function gradient_descent_cwrap(&
       func, grad, callback_data, &
       stepsize_c, iters_c, &
       dim_c, x0_c, &
       min_x_c, min_f_c, min_g_c) &
       result(error_code_c) bind(c, name="gradient_descent_f")
    use, intrinsic :: iso_c_binding
    implicit none
    ! Signature
    type(c_funptr), intent(in), value :: func, grad
    type(c_ptr), intent(in), value :: callback_data
    real(c_double), intent(in), value :: stepsize_c
    integer(c_int), intent(in), value :: iters_c, dim_c
    real(c_double), intent(in) :: x0_c(dim_c)
    real(c_double), intent(out) :: min_x_c(dim_c), min_g_c(dim_c)
    real(c_double), intent(out) :: min_f_c
    integer(c_int) :: error_code_c

    ! Locals
    real(dp) :: x0(dim_c), stepsize
    integer :: iters, error_code
    type(point_value_gradient) :: min
    procedure(objective_function), pointer :: func_p
    procedure(objective_gradient), pointer :: grad_p

    ! Convert input values
    x0 = x0_c
    stepsize = stepsize_c
    iters = iters_c

    !print *, "gradient_descent_cwrap("
    !print *, "  stepsize_c:", stepsize_c
    !print *, "  stepsize:", stepsize
    !print *, "  iters_c:", iters_c
    !print *, "  iters:", iters
    !print *, "  dim_c:", dim_c
    !print *, "  x0_c:", x0_c
    !print *, "  x0:", x0
    !print *, "  min_x_c:", min_x_c
    !print *, "  min%x:", min%x
    !print *, "  min_f_c:", min_f_c
    !print *, "  min%f:", min%f
    !print *, "  min_g_c:", min_g_c
    !print *, "  min%g:", min%g
    !print *, "  error_code_c:", error_code_c
    !print *, "  error_code:", error_code
    !print *, ")"

    ! Prepare and store data for the actual callbacks (which will be
    ! called by the callback wrappers)
    call c_f_procpointer(func, func_p)
    f_cbdata%func => func_p
    f_cbdata%data = callback_data
    call c_f_procpointer(grad, grad_p)
    g_cbdata%grad => grad_p
    g_cbdata%data = callback_data

    ! Call gradient descent
    error_code = gradient_descent(func_wrapper, grad_wrapper, x0, &
         stepsize, iters, min)

    ! Convert output values
    if (error_code == 0) then
       min_x_c = min%x
       min_f_c = min%f
       min_g_c = min%g
    end if
    error_code_c = error_code

    !print *, "gradient_descent_cwrap("
    !print *, "  stepsize_c:", stepsize_c
    !print *, "  stepsize:", stepsize
    !print *, "  iters_c:", iters_c
    !print *, "  iters:", iters
    !print *, "  dim_c:", dim_c
    !print *, "  x0_c:", x0_c
    !print *, "  x0:", x0
    !print *, "  min_x_c:", min_x_c
    !print *, "  min%x:", min%x
    !print *, "  min_f_c:", min_f_c
    !print *, "  min%f:", min%f
    !print *, "  min_g_c:", min_g_c
    !print *, "  min%g:", min%g
    !print *, "  error_code_c:", error_code_c
    !print *, "  error_code:", error_code
    !print *, ")"
  end function gradient_descent_cwrap

  ! Callback Wrappers
  ! -----------------
  !
  ! The C callbacks do not have the right signature because (1) the
  ! types are C types, (2) they allow for arbitrary callback data, and
  ! (3) they need to know the array size.  Therefore the following
  ! wrappers are necessary.  They have the correct Fortran signature to
  ! be passed to gradient_descent, but they then must use module
  ! variables to communicate the extra information: the actual callback
  ! and its data.

  ! Adapter between Fortran callback and C callback (for objective
  ! function)
  function func_wrapper(point, fvalue) result(error_code)
    implicit none
    ! Signature
    real(dp), intent(in) :: point(:)
    real(dp), intent(out) :: fvalue
    integer :: error_code
    ! Locals
    integer(c_int) :: dim_c, error_code_c
    real(c_double) :: point_c(size(point)), fvalue_c

    ! Convert inputs
    dim_c = size(point)
    point_c = point

    ! Call real callback with given args and data
    error_code_c = f_cbdata%func(dim_c, point_c, fvalue_c, f_cbdata%data)

    ! Convert outputs
    fvalue = fvalue_c
    error_code = error_code_c
  end function func_wrapper

  ! Adapter between Fortran callback and C callback (for objective
  ! gradient)
  function grad_wrapper(point, gradient) result(error_code)
    implicit none
    ! Signature
    real(dp), intent(in) :: point(:)
    real(dp), intent(out) :: gradient(size(point))
    integer :: error_code
    ! Locals
    integer(c_int) :: dim_c, error_code_c
    real(c_double) :: point_c(size(point)), gradient_c(size(point))

    ! Convert inputs
    dim_c = size(point)
    point_c = point

    ! Call real callback with given args and data
    error_code_c = g_cbdata%grad(dim_c, point_c, gradient_c, g_cbdata%data)

    ! Convert outputs
    gradient = gradient_c
    error_code = error_code_c
  end function grad_wrapper

end module gd_c
