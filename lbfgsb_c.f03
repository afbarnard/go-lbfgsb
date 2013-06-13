! Module to provide a simple, modern C interface to L-BFGS-B that is
! tailored to Go.
module lbfgsb_c
  use, intrinsic :: iso_c_binding
  implicit none
  private

  integer, parameter :: dp = kind(0d0)

  ! Signatures for C callbacks for computing the objective function
  ! value and the objective function gradient
  abstract interface

     ! Signature of objective function C callback
     !
     ! dim:
     !   Dimensionality of the optimization space, the size of the
     !   arrays used for points, gradients
     !
     ! point:
     !   Point at which to evaluate the objective value.  Array of size
     !   'dim'.
     !
     ! objective_function_value:
     !   Returned value of the objective function
     !
     ! callback_data:
     !   Arbitrary data to be used by the callback
     !
     ! error_code:
     !   Returned error code as defined in TODO
     ! TODO error codes? error struct with code and message? how wrap errors, if at all?
     function objective_function_c(dim, point, objective_function_value, callback_data) &
          result(error_code) bind(c)
       use, intrinsic :: iso_c_binding
       implicit none
       integer(c_int), intent(in), value :: dim
       real(c_double), intent(in) :: point(dim)
       real(c_double), intent(out) :: objective_function_value
       ! The pointer must be passed by value because we want the address
       ! of the data not the address of the pointer.
       type(c_ptr), intent(in), value :: callback_data
       integer(c_int) :: error_code
     end function objective_function_c

     ! Signature of objective gradient C callback
     !
     ! dim:
     !   Dimensionality of the optimization space, the size of the
     !   arrays used for points, gradients
     !
     ! point:
     !   Point at which to evaluate the objective gradient.  Array of
     !   size 'dim'.
     !
     ! objective_function_gradient:
     !   Returned value of the objective gradient
     !
     ! callback_data:
     !   Arbitrary data to be used by the callback
     !
     ! error_code:
     !   Returned error code as defined in TODO
     function objective_gradient_c(dim, point, objective_function_gradient, callback_data) &
          result(error_code) bind(c)
       use, intrinsic :: iso_c_binding
       implicit none
       integer(c_int), intent(in), value :: dim
       real(c_double), intent(in) :: point(dim)
       real(c_double), intent(out) :: objective_function_gradient(dim)
       type(c_ptr), intent(in), value :: callback_data
       integer(c_int) :: error_code
     end function objective_gradient_c

  end interface

contains

  ! L-BFGS-B
  function lbfgsb_minimize( &
       ! Callbacks
       func, grad, callback_data, &
       ! Dimensionality
       dim_c, &
       ! Bounds
       bounds_control_c, lower_bounds_c, upper_bounds_c, &
       ! Parameters
       approximation_size_c, f_tolerance_c, g_tolerance_c, &
       ! Input
       initial_point_c, &
       ! Result
       min_x_c, min_f_c, min_g_c, &
       ! Other
       debug_c) &
       result(error_code_c) bind(c)

    use, intrinsic :: iso_c_binding
    implicit none

    ! Signature
    type(c_funptr), intent(in), value :: func, grad
    type(c_ptr), intent(in), value :: callback_data
    integer(c_int), intent(in), value :: dim_c, approximation_size_c, debug_c
    real(c_double), intent(in), value :: f_tolerance_c, g_tolerance_c
    integer(c_int), intent(in) :: bounds_control_c(dim_c)
    real(c_double), intent(in) :: lower_bounds_c(dim_c), upper_bounds_c(dim_c), initial_point_c(dim_c)
    real(c_double), intent(out) :: min_x_c(dim_c), min_f_c, min_g_c(dim_c)
    integer(c_int) :: error_code_c

    ! Named constants specific to L-BFGS-B internals
    integer, parameter :: &
         task_size = 60, &
         char_state_size = task_size, &
         bool_state_size = 4, &
         int_state_size = 44, &
         real_state_size = 29

    ! Locals (scalars before arrays)
    ! Fortran versions of arguments
    procedure(objective_function_c), pointer :: func_p
    procedure(objective_gradient_c), pointer :: grad_p
    integer :: dim, approximation_size
    real(dp) :: f_tolerance, g_tolerance
    integer :: bounds_control(dim_c)
    real(dp) :: lower_bounds(dim_c), upper_bounds(dim_c), point(dim_c)
    ! Variables and memory for L-BFGS-B
    integer :: print_control
    real(dp) :: func_value, f_factor
    character(len=task_size) :: task
    character(len=char_state_size) :: char_state
    logical :: bool_state(bool_state_size)
    integer :: int_state(int_state_size), &
         working_int_memory(3 * dim_c)
    real(dp) :: grad_value(dim_c), real_state(real_state_size), &
         working_real_memory( &
         2 * approximation_size_c * dim_c + 5 * dim_c + &
         11 * approximation_size_c ** 2 + 8 * approximation_size_c &
         )

    ! Convert inputs from C types to Fortran types
    call c_f_procpointer(func, func_p)
    call c_f_procpointer(grad, grad_p)
    dim = dim_c
    approximation_size = approximation_size_c
    f_tolerance = f_tolerance_c
    g_tolerance = g_tolerance_c
    ! Array copies
    bounds_control = bounds_control_c
    lower_bounds = lower_bounds_c
    upper_bounds = upper_bounds_c
    point = initial_point_c

    ! Translate f_tolerance to f_factor
    f_factor = f_tolerance / 0d0  ! TODO

    ! Translate debug_c to print_control
    print_control = debug_c  / 0  ! TODO

    ! Initialize the task
    task = 'START'

    ! Loop to do tasks and coordinate the optimization
    do while ( &
         task(1:2) == 'FG' .or. &
         task(1:5) == 'NEW_X' .or. &
         task(1:5) == 'START')

       ! Call L-BFGS-B code
       call setulb(dim, approximation_size, point, &
            lower_bounds, upper_bounds, bounds_control, &
            func_value, grad_value, &
            f_factor, g_tolerance, &
            working_real_memory, working_int_memory, &
            task, print_control, &
            char_state, bool_state, int_state, real_state)

       ! Do the 'calculate function and gradient' task
       if (task(1:2) == 'FG') then
          ! Try to get away with not converting Fortran arrays to C

          ! Call objective function
          error_code_c = func_p(dim_c, point, func_value, callback_data)
          if (error_code_c /= 0) exit

          ! Call objective function gradient
          error_code_c = grad_p(dim_c, point, grad_value, callback_data)
          if (error_code_c /= 0) exit
       end if

       ! Nothing to do for other tasks (NEW_X just means a step was
       ! taken to the current x)
    end do

    ! Analyze task and error state to see how to return
    ! TODO

    ! Convert outputs from Fortran types to C types
    min_x_c = point
    min_f_c = func_value
    min_g_c = grad_value
    error_code_c = 0
  end function lbfgsb_minimize

end module lbfgsb_c
