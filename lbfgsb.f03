module lbfgsb
  implicit none
  private

  ! Named constants
  public dp
  integer, parameter :: dp = kind(0d0)

  ! Named constants specific to L-BFGS-B internals
  public task_size, char_state_size, bool_state_size, int_state_size, &
       real_state_size
  integer, parameter :: &
       task_size = 60, &
       char_state_size = task_size, &
       bool_state_size = 4, &
       int_state_size = 44, &
       real_state_size = 29

  ! Explicit interface for L-BFGS-B entrance point.  See lbfgsb/lbfgsb.f
  ! for explanation.
  public setulb
  interface
     subroutine setulb(n, m, x, l, u, nbd, f, g, factr, pgtol, &
          wa, iwa, task, iprint, csave, lsave, isave, dsave)
       import dp, task_size, char_state_size, bool_state_size, &
            int_state_size, real_state_size
       implicit none
       ! Inputs
       integer, intent(in) :: n, m, nbd(n), iprint
       real(dp), intent(in) :: l(n), u(n), factr, pgtol
       ! Inputs/Outputs
       character(len=task_size), intent(inout) :: task
       character(len=char_state_size), intent(inout) :: csave
       logical, intent(inout) :: lsave(bool_state_size)
       integer, intent(inout) :: iwa(3*n), isave(int_state_size)
       real(dp), intent(inout) :: x(n), f, g(n), &
            wa(2*m*n+5*n+11*m*m+8*m), dsave(real_state_size)
     end subroutine setulb
  end interface

end module lbfgsb
