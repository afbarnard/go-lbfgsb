Collected Output Statements of lbfgsb.f
=======================================

These are all the output statements in lbfgsb.f followed by their format
definitions.  The purpose is to analyze the output levels in order to
choose the appropriate output level for statements that don't have one.

My conclusion is that the output statements that are not guarded by an
output level are high priority and so should have an output level of 0.
This decision would also support the description in the lbfgsb API
documentation which claims that `iprint<0` produces no output and that
`iprint=0` only produces a final summary; the 0 level would be amended
with important messages.


Output Level None
-----------------

```fortran
         if (gd .ge. zero) then
c                               the directional derivative >=0.
c                               Line search is impossible.
            write(6,*)' ascent direction in projection gd = ', gd
            info = -4
            return
         endif

         write(6,*) ' Positive dir derivative in projection '
         write(6,*) ' Using the backtracking step '
```


Output Level Mixed
------------------

```fortran
      if (iprint .ge. 99) then
         write (6,*) 'LINE SEARCH',iback,' times; norm of step = ',xstep
         write (6,2001) iter,f,sbgnrm
         if (iprint .gt. 100) then
            write (6,1004) 'X =',(x(i), i = 1, n)
            write (6,1004) 'G =',(g(i), i = 1, n)
         endif
      else if (iprint .gt. 0) then
         imod = mod(iter,iprint)
         if (imod .eq. 0) write (6,2001) iter,f,sbgnrm
      endif
 1004 format (/,a4, 1p, 6(1x,d11.4),/,(4x,1p,6(1x,d11.4)))
 2001 format
     +  (/,'At iterate',i5,4x,'f= ',1p,d12.5,4x,'|proj g|= ',1p,d12.5)

      if (iprint .ge. 0) then
         write (6,3003)
         write (6,3004)
         write(6,3005) n,iter,nfgv,nintol,nskip,nact,sbgnrm,f
         if (iprint .ge. 100) then
            write (6,1004) 'X =',(x(i),i = 1,n)
         endif
         if (iprint .ge. 1) write (6,*) ' F =',f
      endif
 1004 format (/,a4, 1p, 6(1x,d11.4),/,(4x,1p,6(1x,d11.4)))
 3003 format (/,
     + '           * * *',/,/,
     + 'Tit   = total number of iterations',/,
     + 'Tnf   = total number of function evaluations',/,
     + 'Tnint = total number of segments explored during',
     +           ' Cauchy searches',/,
     + 'Skip  = number of BFGS updates skipped',/,
     + 'Nact  = number of active bounds at final generalized',
     +          ' Cauchy point',/,
     + 'Projg = norm of the final projected gradient',/,
     + 'F     = final function value',/,/,
     + '           * * *')
 3004 format (/,3x,'N',4x,'Tit',5x,'Tnf',2x,'Tnint',2x,
     +       'Skip',2x,'Nact',5x,'Projg',8x,'F')
 3005 format (i5,2(1x,i6),(1x,i6),(2x,i4),(1x,i5),1p,2(2x,d10.3))

      if (iprint .ge. 0) then
         write (6,3009) task
         if (info .ne. 0) then
            if (info .eq. -1) write (6,9011)
            if (info .eq. -2) write (6,9012)
            if (info .eq. -3) write (6,9013)
            if (info .eq. -4) write (6,9014)
            if (info .eq. -5) write (6,9015)
            if (info .eq. -6) write (6,*)' Input nbd(',k,') is invalid.'
            if (info .eq. -7)
     +      write (6,*)' l(',k,') > u(',k,').  No feasible solution.'
            if (info .eq. -8) write (6,9018)
            if (info .eq. -9) write (6,9019)
         endif
         if (iprint .ge. 1) write (6,3007) cachyt,sbtime,lnscht
         write (6,3008) time
         if (iprint .ge. 1) then
            if (info .eq. -4 .or. info .eq. -9) then
               write (itfile,3002)
     +             iter,nfgv,nseg,nact,word,iback,stp,xstep
            endif
            write (itfile,3009) task
            if (info .ne. 0) then
               if (info .eq. -1) write (itfile,9011)
               if (info .eq. -2) write (itfile,9012)
               if (info .eq. -3) write (itfile,9013)
               if (info .eq. -4) write (itfile,9014)
               if (info .eq. -5) write (itfile,9015)
               if (info .eq. -8) write (itfile,9018)
               if (info .eq. -9) write (itfile,9019)
            endif
            write (itfile,3008) time
         endif
      endif
 3002 format(2(1x,i4),2(1x,i5),2x,a3,1x,i4,1p,2(2x,d7.1),6x,'-',10x,'-')
 3007 format (/,' Cauchy                time',1p,e10.3,' seconds.',/
     +        ' Subspace minimization time',1p,e10.3,' seconds.',/
     +        ' Line search           time',1p,e10.3,' seconds.')
 3008 format (/,' Total User time',1p,e10.3,' seconds.',/)
 3009 format (/,a60)
 9011 format (/,
     +' Matrix in 1st Cholesky factorization in formk is not Pos. Def.')
 9012 format (/,
     +' Matrix in 2st Cholesky factorization in formk is not Pos. Def.')
 9013 format (/,
     +' Matrix in the Cholesky factorization in formt is not Pos. Def.')
 9014 format (/,
     +' Derivative >= 0, backtracking line search impossible.',/,
     +'   Previous x, f and g restored.',/,
     +' Possible causes: 1 error in function or gradient evaluation;',/,
     +'                  2 rounding errors dominate computation.')
 9015 format (/,
     +' Warning:  more than 10 function and gradient',/,
     +'   evaluations in the last line search.  Termination',/,
     +'   may possibly be caused by a bad search direction.')
 9018 format (/,' The triangular system is singular.')
 9019 format (/,
     +' Line search cannot locate an adequate point after 20 function',/
     +,'  and gradient evaluations.  Previous x, f and g restored.',/,
     +' Possible causes: 1 error in function or gradient evaluation;',/,
     +'                  2 rounding error dominate computation.')
```


Output Level 0
--------------

```fortran
      if (iprint .ge. 0) then
         if (prjctd) write (6,*)
     +   'The initial X is infeasible.  Restart with its projection.'
         if (.not. cnstnd)
     +      write (6,*) 'This problem is unconstrained.'
      endif

      if (iprint .gt. 0) write (6,1001) nbdd
 1001 format (/,'At X0 ',i9,' variables are exactly at the bounds')

         if (iprint .ge. 0) write (6,*) 'Subgnorm = 0.  GCP = X.'

      if (iprint .ge. 0) then
         write (6,7001) epsmch
         write (6,*) 'N = ',n,'    M = ',m
         if (iprint .ge. 1) then
            write (itfile,2001) epsmch
            write (itfile,*)'N = ',n,'    M = ',m
            write (itfile,9001)
            if (iprint .gt. 100) then
               write (6,1004) 'L =',(l(i),i = 1,n)
               write (6,1004) 'X0 =',(x(i),i = 1,n)
               write (6,1004) 'U =',(u(i),i = 1,n)
            endif
         endif
      endif
 1004 format (/,a4, 1p, 6(1x,d11.4),/,(4x,1p,6(1x,d11.4)))
 2001 format ('RUNNING THE L-BFGS-B CODE',/,/,
     + 'it    = iteration number',/,
     + 'nf    = number of function evaluations',/,
     + 'nseg  = number of segments explored during the Cauchy search',/,
     + 'nact  = number of active bounds at the generalized Cauchy point'
     + ,/,
     + 'sub   = manner in which the subspace minimization terminated:'
     + ,/,'        con = converged, bnd = a bound was reached',/,
     + 'itls  = number of iterations performed in the line search',/,
     + 'stepl = step length used',/,
     + 'tstep = norm of the displacement (total step)',/,
     + 'projg = norm of the projected gradient',/,
     + 'f     = function value',/,/,
     + '           * * *',/,/,
     + 'Machine precision =',1p,d10.3)
 7001 format ('RUNNING THE L-BFGS-B CODE',/,/,
     + '           * * *',/,/,
     + 'Machine precision =',1p,d10.3)
 9001 format (/,3x,'it',3x,'nf',2x,'nseg',2x,'nact',2x,'sub',2x,'itls',
     +        2x,'stepl',4x,'tstep',5x,'projg',8x,'f')
```


Output Level 1
--------------

```fortran
      if (iprint .ge. 1) then
         write (6,1002) iter,f,sbgnrm
         write (itfile,1003) iter,nfgv,sbgnrm,f
      endif
 1002 format
     +  (/,'At iterate',i5,4x,'f= ',1p,d12.5,4x,'|proj g|= ',1p,d12.5)
 1003 format (2(1x,i4),5x,'-',5x,'-',3x,'-',5x,'-',5x,'-',8x,'-',3x,
     +        1p,2(1x,d10.3))

         if(iprint .ge. 1) write (6, 1005)
 1005 format (/,
     +' Singular triangular system detected;',/,
     +'   refresh the lbfgs memory and restart the iteration.')

         if(iprint .ge. 1) write (6, 1006)
 1006 format (/,
     +' Nonpositive definiteness in Cholesky factorization in formk;',/,
     +'   refresh the lbfgs memory and restart the iteration.')

         if(iprint .ge. 1) write (6, 1005)

            if(iprint .ge. 1) write (6, 1008)
 1008 format (/,
     +' Bad direction in the line search;',/,
     +'   refresh the lbfgs memory and restart the iteration.')

         if (iprint .ge. 1) write (6,1004) dr, ddum
 1004 format ('  ys=',1p,e10.3,'  -gs=',1p,e10.3,' BFGS update SKIPPED')

         if(iprint .ge. 1) write (6, 1007)
 1007 format (/,
     +' Nonpositive definiteness in Cholesky factorization in formt;',/,
     +'   refresh the lbfgs memory and restart the iteration.')

      if (iprint .ge. 1) write (itfile,3001)
     +          iter,nfgv,nseg,nact,word,iback,stp,xstep,sbgnrm,f
 3001 format(2(1x,i4),2(1x,i5),2x,a3,1x,i4,1p,2(2x,d7.1),1p,2(1x,d10.3))
```


Output Level 99
---------------

```fortran
      if (iprint .ge. 99) write (6,1001) iter + 1
 1001 format (//,'ITERATION ',i5)

      if (iprint .ge. 99) write (6,3010)
 3010 format (/,'---------------- CAUCHY entered-------------------')

      if (iprint .ge. 99)
     +   write (6,*) 'There are ',nbreak,'  breakpoints '

      if (iprint .ge. 99) then
         write (6,*)
         write (6,*) 'GCP found in this segment'
         write (6,4010) nseg,f1,f2
         write (6,6010) dtm
      endif
 4010 format ('Piece    ',i3,' --f1, f2 at start point ',1p,2(1x,d11.4))
 6010 format ('Distance to the stationary point =  ',1p,d11.4)

      if (iprint .ge. 99) write (6,2010)
 2010 format (/,'---------------- exit CAUCHY----------------------',/)

         if (iprint .ge. 99) write (6,*)
     +       n+1-ileave,' variables leave; ',nenter,' variables enter'

      if (iprint .ge. 99) write (6,*)
     +      nfree,' variables are free at GCP ',iter + 1

      if (iprint .ge. 99) write (6,1001)
 1001 format (/,'----------------SUBSM entered-----------------',/)

      if (iprint .ge. 99) write (6,1004)
 1004 format (/,'----------------exit SUBSM --------------------',/)
```


Output Level 100
----------------

```fortran
         if (iprint .gt. 100) write (6,1010) (xcp(i), i = 1, n)
 1010 format ('Cauchy X =  ',/,(4x,1p,6(1x,d11.4)))

      if (dt .ne. zero .and. iprint .ge. 100) then
         write (6,4011) nseg,f1,f2
         write (6,5010) dt
         write (6,6010) dtm
      endif
 4011 format (/,'Piece    ',i3,' --f1, f2 at start point ',
     +        1p,2(1x,d11.4))
 5010 format ('Distance to the next break point =  ',1p,d11.4)
 6010 format ('Distance to the stationary point =  ',1p,d11.4)

      if (iprint .ge. 100) write (6,*) 'Variable  ',ibp,'  is fixed.'

      if (iprint .gt. 100) write (6,1010) (xcp(i),i = 1,n)
 1010 format ('Cauchy X =  ',/,(4x,1p,6(1x,d11.4)))

               if (iprint .ge. 100) write (6,*)
     +             'Variable ',k,' leaves the set of free variables'

               if (iprint .ge. 100) write (6,*)
     +             'Variable ',k,' enters the set of free variables'
```
