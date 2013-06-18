! Copyright (c) 2013 Aubrey Barnard.  This is free software.  See
! LICENSE.txt for details.

! Module lbfgsb_c provides a simple, modern C interface to the L-BFGS-B
! FORTRAN 77 code.  This module compiles together with module lbfgsb and
! the L-BFGS-B FORTRAN 77 code to create a L-BFGS-B library with a C
! API.  The Go package then uses the C API.  My goal is to implement all
! of the necessary functionality of the library in Fortran so it can be
! used by other code and not just Go.
module lbfgsb_c
  use, intrinsic :: iso_c_binding
  use lbfgsb
  implicit none
  private

  ! Public procedures
  public lbfgsb_minimize

  ! Public constants and error codes
  enum, bind(c)
     enumerator :: &
          LBFGSB_OK = 0, &  ! No error
          LBFGSB_APPROXIMATE, &  ! Approximate result; could not satisfy tolerances
          LBFGSB_RUNTIME_ERROR, &  ! Error in L-BFGS-B or its invocation
          LBFGSB_FUNCTION_ERROR, &  ! Error in computing objective function
          LBFGSB_GRADIENT_ERROR, &  ! Error in computing objective gradient
          LBFGSB_TASK_ERROR  ! Unrecognized task
  end enum

  ! Signatures for C callbacks for computing the objective function
  ! value and the objective function gradient
  public objective_function_c, objective_gradient_c
  abstract interface

     ! Signature of objective function C callback
     !
     ! 'dim': Dimensionality of the optimization space; the size of the
     !    arrays used for points, gradients.
     !
     ! 'point': Point at which to evaluate the objective value.  Array
     !    of size 'dim'.
     !
     ! 'objective_function_value': Returns the value of the objective
     !    function.
     !
     ! 'callback_data': Arbitrary data to be used by the callback.
     !
     ! 'error_message_length': Usable length of 'error_message_c'
     !    buffer.  Recommend at least 100.
     !
     ! 'error_message': Returns a message explaining the exit status.
     !
     ! 'error_code': Returns the exit status code, one of the
     !    LBFGSB_STATUS_* constants defined in enumeration above.
     function objective_function_c(dim, point, objective_function_value, &
          callback_data, error_message_length, error_message) &
          result(error_code) bind(c)
       use, intrinsic :: iso_c_binding
       implicit none
       integer(c_int), intent(in), value :: dim
       real(c_double), intent(in) :: point(dim)
       real(c_double), intent(out) :: objective_function_value
       ! The pointer must be passed by value because we want the address
       ! of the data not the address of the pointer.
       type(c_ptr), intent(in), value :: callback_data
       integer(c_int), intent(in), value :: error_message_length
       character(c_char), intent(out) :: &
            error_message(error_message_length)
       integer(c_int) :: error_code
     end function objective_function_c

     ! Signature of objective gradient C callback
     !
     ! 'dim': Dimensionality of the optimization space; the size of the
     !    arrays used for points, gradients.
     !
     ! 'point': Point at which to evaluate the objective value.  Array
     !    of size 'dim'.
     !
     ! 'objective_function_gradient': Returns the value of the objective
     !    gradient.
     !
     ! 'callback_data': Arbitrary data to be used by the callback.
     !
     ! 'error_message_length': Usable length of 'error_message_c'
     !    buffer.  Recommend at least 100.
     !
     ! 'error_message': Returns a message explaining the exit status.
     !
     ! 'error_code': Returns the exit status code, one of the
     !    LBFGSB_STATUS_* constants defined in enumeration above.
     function objective_gradient_c(dim, point, objective_function_gradient, &
          callback_data, error_message_length, error_message) &
          result(error_code) bind(c)
       use, intrinsic :: iso_c_binding
       implicit none
       integer(c_int), intent(in), value :: dim
       real(c_double), intent(in) :: point(dim)
       real(c_double), intent(out) :: objective_function_gradient(dim)
       type(c_ptr), intent(in), value :: callback_data
       integer(c_int), intent(in), value :: error_message_length
       character(c_char), intent(out) :: &
            error_message(error_message_length)
       integer(c_int) :: error_code
     end function objective_gradient_c

  end interface

contains

  ! lbfgsb_minimize optimizes the given objective within the given
  ! bounds using the L-BFGS-B optimization algorithm.  The objective is
  ! specified via its value and gradient functions.  Returns an exit
  ! status code.  Throughout, 'x' refers to points, 'f' refers to the
  ! objective function, 'g' refers to the gradient of the objective
  ! function.
  !
  ! 'func': Pointer to objective value function whose signature is given
  !    by objective_function_c or lbfgsb_objective_function_type.
  !
  ! 'grad': Pointer to objective gradient function whose signature is
  !    given by objective_gradient_c or lbfgsb_objective_gradient_type.
  !
  ! 'callback_data': Pointer to user-specified data passed to 'func' and
  !    'grad' when they are called.  May be null (or anything) because
  !    this function does not process it, only passes it along.
  !
  ! 'dim_c': Dimensionality of the optimization space; length of the
  !    arrays 'initial_point_c', 'min_x_c', 'min_g_c',
  !    'bounds_control_c', 'lower_bounds_c', 'upper_bounds_c'.
  !
  ! 'bounds_control_c': Array specifying the type of bounds for each
  !    dimension.
  !
  !    bounds_control_c[i] =
  !       * 0 means x[i] is unbounded,
  !       * 1 means x[i] has a lower bound in lower_bounds_c[i],
  !       * 2 means x[i] has both lower and upper bounds,
  !       * 3 means x[i] has an upper bound in upper_bounds_c[i].
  !
  ! 'lower_bounds_c': Array of lower bounds.  lower_bounds_c[i] is
  !    accessed only when indicated in bounds_control_c[i].
  !
  ! 'upper_bounds_c': Array of upper bounds.  upper_bounds_c[i] is
  !    accessed only when indicated in bounds_control_c[i].
  !
  ! 'approximation_size_c': The amount of history (points and gradients)
  !    to store and use to approximate the inverse Hessian matrix.  More
  !    history allows a better approximation and uses more memory.  The
  !    recommended range is [3,20].
  !
  ! 'f_tolerance_c': Precision of objective function required for
  !    convergence.  That is, specifying 1e-5 achieves about 5 digits of
  !    precision of the objective function value.  Convergence at the
  !    i-th iteration requires:
  !
  !    (f(x[i-1]) - f(x[i])) / max(|f(x[i-1])|,|f(x[i])|,1) <= f_tolerance.
  !
  ! 'g_tolerance_c': Maximum magnitude of objective gradient allowed for
  !    convergence.  That is, a value of 1e-5 specifies the gradient
  !    must equal zero to at least 5 digits.  Convergence at the i-th
  !    iteration requires:
  !
  !    |P(g(x[i]))|inf <= g_tolerance
  !
  !    where P(g(x)) is the projected gradient at x.
  !
  ! 'initial_point_c': Point from which minimization starts, x[0].
  !
  ! 'min_x_c': Returns the location of the minimum, an array.
  !
  ! 'min_f_c': Returns the objective function value at the minimum.
  !
  ! 'min_g_c': Returns the gradient at the minimum, an array.
  !
  ! 'error_message_length_c': Usable length of 'error_message_c' buffer.
  !    Recommend at least 100.
  !
  ! 'error_message_c': Message explaining the exit status.
  !
  ! 'debug_c': Output verbosity level where 0 means no output and larger
  !    values mean increasing output.
  !
  ! 'error_code_c': Returns the exit status code, one of the
  !    LBFGSB_STATUS_* constants defined in enumeration above.
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
       ! Error, debug
       error_message_length_c, error_message_c, debug_c) &
       result(error_code_c) bind(c)

    implicit none

    ! Signature
    type(c_funptr), intent(in), value :: func, grad
    type(c_ptr), intent(in), value :: callback_data
    integer(c_int), intent(in), value :: dim_c, approximation_size_c, &
         error_message_length_c, debug_c
    real(c_double), intent(in), value :: f_tolerance_c, g_tolerance_c
    integer(c_int), intent(in) :: bounds_control_c(dim_c)
    real(c_double), intent(in) :: lower_bounds_c(dim_c), &
         upper_bounds_c(dim_c), initial_point_c(dim_c)
    character(c_char), intent(out) :: &
         error_message_c(error_message_length_c)
    real(c_double), intent(out) :: min_x_c(dim_c), min_f_c, &
         min_g_c(dim_c)
    integer(c_int) :: error_code_c

    ! Locals (scalars before arrays)
    ! Fortran versions of arguments
    procedure(objective_function_c), pointer :: func_pointer
    procedure(objective_gradient_c), pointer :: grad_pointer
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
    call c_f_procpointer(func, func_pointer)
    call c_f_procpointer(grad, grad_pointer)
    ! Array copies (TODO are these necessary? i.e. how compatible are the storage and types?)
    bounds_control = bounds_control_c
    lower_bounds = lower_bounds_c
    upper_bounds = upper_bounds_c
    point = initial_point_c

    ! Start with an empty error message (fill entire string with nulls)
    error_message_c = c_null_char

    ! Translate f_tolerance to f_factor.  The convergence tolerance for
    ! the objective function is computed by the L-BFGS-B code as
    ! f_factor * epsilon(1d0) but I want to express the tolerance in
    ! terms of digits of precision, analogous to g_tolerance.
    f_factor = f_tolerance_c / epsilon(1d0)

    ! Translate debug_c (zero-based verbosity indicator) approximately
    ! to print_control (general signed integer with meaning attached to
    ! certain values)
    print_control = debug_c - 1

    ! Initialize the task
    task = 'START'

    ! Loop to do tasks and coordinate the optimization
    do while ( &
         task(1:2) == 'FG' .or. &
         task(1:5) == 'NEW_X' .or. &
         task(1:5) == 'START')

       ! Call L-BFGS-B code
       call setulb(dim_c, approximation_size_c, point, &
            lower_bounds, upper_bounds, bounds_control, &
            func_value, grad_value, &
            f_factor, g_tolerance_c, &
            working_real_memory, working_int_memory, &
            task, print_control, &
            char_state, bool_state, int_state, real_state)

       ! Do the 'calculate function and gradient' task
       if (task(1:2) == 'FG') then
          ! Try to get away with not converting Fortran arrays to C

          ! Call objective function
          error_code_c = func_pointer(dim_c, point, func_value, &
               callback_data, error_message_length_c, error_message_c)
          if (error_code_c /= LBFGSB_OK) exit

          ! Call objective function gradient
          error_code_c = grad_pointer(dim_c, point, grad_value, &
               callback_data, error_message_length_c, error_message_c)
          if (error_code_c /= LBFGSB_OK) exit
       end if

       ! Nothing to do for other tasks (NEW_X just means a step was
       ! taken to the current x)
    end do

    ! Analyze task and error state to see how to return
    if (error_code_c == LBFGSB_OK) then
       ! Objective and gradient evaluations were OK but L-BFGS-B may not
       ! be.  Regardless, take what we can from the outputs.
       min_x_c = point
       min_f_c = func_value
       min_g_c = grad_value

       ! Copy task into error message
       call convert_f_c_string(task, error_message_length_c, error_message_c)

       ! Check for normal or problematic termination
       if (task(1:4) == 'CONV') then
          ! Converged.  Normal termination.  Leave error (task) message
          ! as it may be informative.
          error_code_c = LBFGSB_OK
       else if (task(1:4) == 'ABNO') then
          ! Could not satisfy termination conditions.  Result is best
          ! approximation.
          error_code_c = LBFGSB_APPROXIMATE
       else if (task(1:5) == 'ERROR') then
          ! Runtime or user error
          error_code_c = LBFGSB_RUNTIME_ERROR
       else
          ! Unrecognized task
          error_code_c = LBFGSB_TASK_ERROR
       end if
    else
       ! There was a problem computing the objective or the gradient.
       ! The error message and code were already properly set by the
       ! call to function/gradient.  Fill C types with zeros so callers
       ! do not act unwittingly on garbage.
       min_x_c = 0d0
       min_f_c = 0d0
       min_g_c = 0d0
    end if
  end function lbfgsb_minimize

  ! Converts Fortran strings to C strings ensuring length bounds, null
  ! chars, and all that.
  subroutine convert_f_c_string(string_f, string_c_length, string_c)
    implicit none
    ! Signature
    character(len=*), intent(in) :: string_f
    integer(c_int), intent(in) :: string_c_length
    character(c_char), intent(out) :: string_c(string_c_length)
    ! Locals
    integer :: length

    ! Find the length of the shorter string
    length = min(len(string_f), string_c_length)
    ! Copy no more than 'length' characters
    string_c = string_f(1:length)
    ! Fill rest of string (if any) with nulls
    string_c(length+1:string_c_length) = c_null_char
  end subroutine convert_f_c_string

end module lbfgsb_c
