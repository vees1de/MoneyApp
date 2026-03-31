import { HttpErrorResponse, HttpInterceptorFn } from '@angular/common/http';
import { inject } from '@angular/core';
import { Router } from '@angular/router';
import { catchError, throwError } from 'rxjs';

import { AuthSessionService } from './auth-session.service';
import { AuthStateService } from './auth-state.service';

export const authHttpInterceptor: HttpInterceptorFn = (req, next) => {
  const session = inject(AuthSessionService);
  const authState = inject(AuthStateService);
  const router = inject(Router);

  const token = session.getAccessToken();
  const withAuth = token
    ? req.clone({
        setHeaders: {
          Authorization: `Bearer ${token}`,
        },
      })
    : req;

  return next(withAuth).pipe(
    catchError((error: unknown) => {
      if (error instanceof HttpErrorResponse && error.status === 401) {
        session.clear();
        authState.setCurrentUser(null);
        router.navigateByUrl('/login');
      }
      return throwError(() => error);
    }),
  );
};
