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

  ! Public status codes describing the exit or error status of L-BFGS-B.
  ! Multiple statuses are necessary because success in optimization is
  ! more complex than a single flag or error.  There are four exit
  ! statuses and two runtime error statuses.
  !
  ! 1. Success.  Normal termination having converged.
  !
  ! 2. Approximate success.  Normal operation resulting in a more
  ! approximate answer.  For example, unable to meet termination
  ! tolerances.
  !
  ! 3. Warning.  The result could be OK, but there were some issues that
  ! may have reduced the quality of the result and require examination.
  ! For example, slight numerical problems, exceeding iteration or time
  ! bounds.
  !
  ! 4. Failure of optimization.  For example, a necessary condition of
  ! the algorithm was not met, severe numerical problems.  (This status
  ! includes failures to evaluate the objective function or objective
  ! gradient.)
  !
  ! 5. Usage error.  For example, invalid constraints, bad parameters.
  ! Responsibility is on the caller.
  !
  ! 6. Internal error.  Other runtime or programming/logic error which
  ! may be a bug.  Responsibility is on this package.
  enum, bind(c)
     ! Constants automatically increment
     enumerator :: &
          LBFGSB_STATUS_SUCCESS = 0, &
          LBFGSB_STATUS_APPROXIMATE, &
          LBFGSB_STATUS_WARNING, &
          LBFGSB_STATUS_FAILURE, &
          LBFGSB_STATUS_USAGE_ERROR, &
          LBFGSB_STATUS_INTERNAL_ERROR
  end enum

  ! Signatures for C callbacks for computing the objective function
  ! value and the objective function gradient
  public objective_function_c, objective_gradient_c
  abstract interface

     ! Signature of objective function C callback.
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
     ! 'status_message': Returns a message (null-terminated C string)
     !    explaining the exit status.
     !
     ! 'status_message_length': Usable length of 'status_message_c'
     !    buffer.  Recommend at least 100.
     !
     ! 'status': Returns the exit status code, one of the
     !    LBFGSB_STATUS_* constants defined in enumeration above.
     function objective_function_c(dim, point, objective_function_value, &
          callback_data, status_message, status_message_length) &
          result(status) bind(c)
       use, intrinsic :: iso_c_binding
       implicit none
       integer(c_int), intent(in), value :: dim
       real(c_double), intent(in) :: point(dim)
       real(c_double), intent(out) :: objective_function_value
       ! The pointer must be passed by value because we want the address
       ! of the data not the address of the pointer.
       type(c_ptr), intent(in), value :: callback_data
       integer(c_int), intent(in), value :: status_message_length
       character(c_char), intent(out) :: &
            status_message(status_message_length)
       integer(c_int) :: status
     end function objective_function_c

     ! Signature of objective gradient C callback.
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
     ! 'status_message': Returns a message (null-terminated C string)
     !    explaining the exit status.
     !
     ! 'status_message_length': Usable length of 'status_message_c'
     !    buffer.  Recommend at least 100.
     !
     ! 'status': Returns the exit status code, one of the
     !    LBFGSB_STATUS_* constants defined in enumeration above.
     function objective_gradient_c(dim, point, objective_function_gradient, &
          callback_data, status_message, status_message_length) &
          result(status) bind(c)
       use, intrinsic :: iso_c_binding
       implicit none
       integer(c_int), intent(in), value :: dim
       real(c_double), intent(in) :: point(dim)
       real(c_double), intent(out) :: objective_function_gradient(dim)
       type(c_ptr), intent(in), value :: callback_data
       integer(c_int), intent(in), value :: status_message_length
       character(c_char), intent(out) :: &
            status_message(status_message_length)
       integer(c_int) :: status
     end function objective_gradient_c

  end interface

  ! Private constants
  integer, parameter :: state_size = 14

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
  ! 'iters_c': Returns the number of iterations performed.
  !
  ! 'evals_c': Returns the number of evaluations performed.  Each
  !    evaluation consists of both a function and a gradient call (so
  !    the total number of callbacks is double the number of
  !    evaluations).
  !
  ! 'status_message_c': Returns a message (null-terminated C string)
  !    explaining the exit status.
  !
  ! 'status_message_length_c': Usable length of 'status_message_c'
  !    buffer.  Recommend at least 100.
  !
  ! 'print_control_c': Fortran output verbosity level.  If set to
  !    generate output, a summary file 'iterate.dat' is also generated.
  !
  !    print_control_c =
  !       * 0: no output
  !       * 1: print one summary at the end
  !       * 2-99: print F and G every so many iterations
  !       * 100: print details of every iteration but not X and G
  !       * 101: also print changes of the active set and the final X
  !       * 102: print details of every iteration including X and G
  !
  ! 'status_c': Returns the exit status code, one of the LBFGSB_STATUS_*
  !    constants defined in enumeration above.
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
       min_x_c, min_f_c, min_g_c, iters_c, evals_c, &
       ! Exit status, printing
       status_message_c, status_message_length_c, print_control_c) &
       result(status_c) bind(c)

    implicit none

    ! Signature
    type(c_funptr), intent(in), value :: func, grad
    type(c_ptr), intent(in), value :: callback_data
    integer(c_int), intent(in), value :: dim_c, approximation_size_c, &
         status_message_length_c, print_control_c
    real(c_double), intent(in), value :: f_tolerance_c, g_tolerance_c
    integer(c_int), intent(in) :: bounds_control_c(dim_c)
    real(c_double), intent(in) :: lower_bounds_c(dim_c), &
         upper_bounds_c(dim_c), initial_point_c(dim_c)
    character(c_char), intent(out) :: &
         status_message_c(status_message_length_c)
    integer(c_int), intent(out) :: iters_c, evals_c
    real(c_double), intent(out) :: min_x_c(dim_c), min_f_c, &
         min_g_c(dim_c)
    integer(c_int) :: status_c

    ! Locals (scalars before arrays)
    ! Fortran versions of arguments
    procedure(objective_function_c), pointer :: func_pointer
    procedure(objective_gradient_c), pointer :: grad_pointer
    real(dp) :: point(dim_c)
    ! Variables and memory for L-BFGS-B
    integer :: print_control
    real(dp) :: func_value, f_factor
    character(len=task_size) :: task
    character(len=char_state_size) :: char_state
    character(len=state_size) :: state
    character(len=2*task_size) :: message
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
    ! Copy initial_point_c to point because point is written to
    point = initial_point_c
    ! Other arrays do not need to be copied because there binary
    ! representations are compatible (but I'm not sure if this will
    ! always be the case).  Plus the other arrays are only read, not
    ! written.

    ! Start with an empty status message (fill entire string with nulls)
    status_message_c = c_null_char

    ! Translate f_tolerance to f_factor.  The convergence tolerance for
    ! the objective function is computed by the L-BFGS-B code as
    ! f_factor * epsilon(1d0) but I want to express the tolerance in
    ! terms of digits of precision, analogous to g_tolerance.
    f_factor = f_tolerance_c / epsilon(1d0)

    ! Translate print_control_c which is a zero-based version of
    ! print_control (which is possibly negative)
    print_control = print_control_c - 1
    ! TODO consider debug vs. print_control and what about logging? (could log via a callback)

    ! Initialize the state and task
    state = 'START'
    task = state

    ! Loop to do tasks and coordinate the optimization
    do while ( &
         state == 'EVAL_FG' .or. &
         state == 'NEW_X' .or. &
         state == 'WARNING' .or. &
         state == 'START')

       ! Call L-BFGS-B code
       call setulb(dim_c, approximation_size_c, point, &
            lower_bounds_c, upper_bounds_c, bounds_control_c, &
            func_value, grad_value, &
            f_factor, g_tolerance_c, &
            working_real_memory, working_int_memory, &
            task, print_control, &
            char_state, bool_state, int_state, real_state)

       ! Interpret the returned task
       call interpret_task(task, state, message)

       ! Act on the current state
       select case (state)
       case ('EVAL_FG')
          ! Calculate function and gradient.  Try to get away with not
          ! converting Fortran arrays to C.

          ! Call objective function
          status_c = func_pointer(dim_c, point, func_value, &
               callback_data, status_message_c, status_message_length_c)
          if (status_c /= LBFGSB_STATUS_SUCCESS) exit

          ! Call objective function gradient
          status_c = grad_pointer(dim_c, point, grad_value, &
               callback_data, status_message_c, status_message_length_c)
          if (status_c /= LBFGSB_STATUS_SUCCESS) exit
       case ('WARNING')
          ! TODO handle warnings
       case ('NEW_X')
          ! TODO handle logging
       end select
    end do
    ! End optimization

    ! Return statistics
    iters_c = int_state(30)  ! Current iteration
    evals_c = int_state(34)  ! Total evaluations (each eval = [F(),G()])

    ! Analyze status and state to see how to return
    if (status_c == LBFGSB_STATUS_SUCCESS) then
       ! Objective and gradient evaluations were OK but L-BFGS-B may not
       ! be.  Regardless, take what we can from the outputs.
       min_x_c = point
       min_f_c = func_value
       min_g_c = grad_value

       ! Check for normal or problematic termination
       select case (state)
       case ('CONVERGENCE')
          ! Converged.  Normal termination.  Leave error (task) message
          ! as it may be informative.
          status_c = LBFGSB_STATUS_SUCCESS
       case ('ABNORMAL')
          ! Could not satisfy termination conditions.  Result is best
          ! approximation.
          status_c = LBFGSB_STATUS_APPROXIMATE
       case ('WARNING')
          status_c = LBFGSB_STATUS_WARNING
       case ('ERROR_USAGE')
          ! User error
          status_c = LBFGSB_STATUS_USAGE_ERROR
       case ('ERROR_INTERNAL')
          ! Runtime or internal error
          status_c = LBFGSB_STATUS_INTERNAL_ERROR
       case default
          ! Unrecognized state
          status_c = LBFGSB_STATUS_INTERNAL_ERROR
          message = 'Error: Unrecognized state: '//task
       end select

       ! Copy task message into status message
       call convert_f_c_string(message, status_message_c, status_message_length_c)
    else
       ! There was a problem computing the objective or the gradient.
       ! The error message and code were already properly set by the
       ! call to function/gradient.  Fill C types with zeros so callers
       ! do not act unwittingly on garbage.
       min_x_c = 0d0
       min_f_c = 0d0
       min_g_c = 0d0
    end if

    ! TODO flush Fortran output
  end function lbfgsb_minimize

  ! Interprets the various task strings coming out of L-BFGS-B.  Maps
  ! them to concrete, disjoint states which are easier and less
  ! ambiguous to handle.  Also extracts the message, if any.
  !
  ! The concrete, disjoint states are START, EVAL_FG, NEW_X,
  ! CONVERGENCE, ABNORMAL, WARNING, ERROR_USAGE, ERROR_INTERNAL
  ! padded to 'state_size' characters.
  subroutine interpret_task(task, state, message)
    character(len=*), intent(in) :: task
    character(len=state_size), intent(out) :: state
    character(len=*), intent(out) :: message
    integer :: cut_index

    ! Clear the message to ensure sensible return value
    message = ' '

    ! Extract the first word of the task which is delimited by a colon
    ! (if any) or is the whole word.  (Intrinsic index() returns 0 if
    ! substring not found.)
    cut_index = index(task, ':') - 1
    if (cut_index == -1) cut_index = len_trim(task)

    ! Discriminate the tasks based on the initial characters
    select case (task(1:cut_index))
    case ('START')
       state = 'START'
    case ('FG', 'FG_LNSRCH', 'FG_START')
       state = 'EVAL_FG'
    case ('NEW_X')
       state = 'NEW_X'
    case ('CONVERGENCE')
       state = 'CONVERGENCE'
       message = task(14:)
    case ('ABNORMAL_TERMINATION_IN_LNSRCH')
       state = 'ABNORMAL'
       message = task
    case ('WARNING')
       ! All the warnings appear to relate only to the line search code
       ! and so may not get back to here
       state = 'WARNING'
       message = task(10:)
    case ('ERROR')
       ! It appears all the reported errors are usage errors, rather
       ! than, say division by zero
       state = 'ERROR_USAGE'
       message = task(8:)
    case default
       ! Unrecognized task
       state = 'ERROR_INTERNAL'
       message = 'Unrecognized task: '//task
    end select

    ! Assigned task values
    ! 'FG_START'
    ! 'CONVERGENCE: NORM_OF_PROJECTED_GRADIENT_<=_PGTOL'
    ! 'ABNORMAL_TERMINATION_IN_LNSRCH'
    ! 'RESTART_FROM_LNSRCH'
    ! 'CONVERGENCE: NORM_OF_PROJECTED_GRADIENT_<=_PGTOL'
    ! 'CONVERGENCE: REL_REDUCTION_OF_F_<=_FACTR*EPSMCH'
    ! 'ERROR: N .LE. 0'
    ! 'ERROR: M .LE. 0'
    ! 'ERROR: FACTR .LT. 0'
    ! 'ERROR: INVALID NBD'
    ! 'ERROR: NO FEASIBLE SOLUTION'
    ! 'FG_LNSRCH'
    ! 'NEW_X'
    ! Line search
    ! 'ERROR: STP .LT. STPMIN'
    ! 'ERROR: STP .GT. STPMAX'
    ! 'ERROR: INITIAL G .GE. ZERO'
    ! 'ERROR: FTOL .LT. ZERO'
    ! 'ERROR: GTOL .LT. ZERO'
    ! 'ERROR: XTOL .LT. ZERO'
    ! 'ERROR: STPMIN .LT. ZERO'
    ! 'ERROR: STPMAX .LT. STPMIN'
    ! 'FG'
    ! 'WARNING: ROUNDING ERRORS PREVENT PROGRESS'
    ! 'WARNING: XTOL TEST SATISFIED'
    ! 'WARNING: STP = STPMAX'
    ! 'WARNING: STP = STPMIN'
    ! 'CONVERGENCE'
    !
    ! Compared task values
    ! 'START'
    ! 'ERROR'
    ! 'FG_LN'
    ! 'NEW_X'
    ! 'FG_ST'
    ! 'STOP'
    ! 'CPU'
  end subroutine interpret_task

  ! Converts Fortran strings to C strings ensuring length bounds, null
  ! chars, and all that.
  subroutine convert_f_c_string(string_f, string_c, string_c_length)
    implicit none
    ! Signature
    character(len=*), intent(in) :: string_f
    integer(c_int), intent(in) :: string_c_length
    character(c_char), intent(out) :: string_c(string_c_length)
    ! Locals
    integer :: length, i

    ! Find the length of the shorter string.  Leave room for a
    ! terminating null character.
    length = min(len_trim(string_f), string_c_length - 1)
    ! Copy 'length' characters from the Fortran string to the C
    ! character array.  A string must be converted explicitly to an
    ! array in Fortran as array assignment broadcasts the string
    ! (technically a scalar) to each element of the array.
    forall(i = 1:length) string_c(i) = string_f(i:i)
    ! Fill rest of string (at least one character) with nulls
    string_c(length+1:string_c_length) = c_null_char
  end subroutine convert_f_c_string

end module lbfgsb_c
