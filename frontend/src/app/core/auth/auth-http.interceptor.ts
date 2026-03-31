import { HttpErrorResponse, HttpInterceptorFn } from '@angular/common/http';
import { inject } from '@angular/core';
import { Router } from '@angular/router';
import {
  Observable,
  catchError,
  finalize,
  of,
  shareReplay,
  switchMap,
  throwError,
} from 'rxjs';

import { AuthApiService } from './auth-api.service';
import { AuthSessionService } from './auth-session.service';
import { AuthStateService } from './auth-state.service';

let refreshInFlight$: Observable<string | null> | null = null;

export const authHttpInterceptor: HttpInterceptorFn = (req, next) => {
  const authApi = inject(AuthApiService);
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
      if (!(error instanceof HttpErrorResponse) || error.status !== 401) {
        return throwError(() => error);
      }

      if (req.url.includes('/auth/login') || req.url.includes('/auth/refresh')) {
        session.clear();
        authState.setCurrentUser(null);
        router.navigateByUrl('/login');
        return throwError(() => error);
      }

      const refreshToken = session.getRefreshToken();
      if (!refreshToken) {
        session.clear();
        authState.setCurrentUser(null);
        router.navigateByUrl('/login');
        return throwError(() => error);
      }

      if (!refreshInFlight$) {
        refreshInFlight$ = authApi.refresh({ refresh_token: refreshToken }).pipe(
          switchMap((response) => {
            session.setTokens(response.tokens);
            authState.setCurrentUser(response.user);
            session.setUserSnapshot(response.user);
            return of(response.tokens.access_token);
          }),
          catchError(() => {
            session.clear();
            authState.setCurrentUser(null);
            router.navigateByUrl('/login');
            return of(null);
          }),
          finalize(() => {
            refreshInFlight$ = null;
          }),
          shareReplay(1),
        );
      }

      return refreshInFlight$.pipe(
        switchMap((newToken) => {
          if (!newToken) {
            return throwError(() => error);
          }

          const retriedRequest = req.clone({
            setHeaders: {
              Authorization: `Bearer ${newToken}`,
            },
          });
          return next(retriedRequest);
        }),
      );
    }),
  );
};
