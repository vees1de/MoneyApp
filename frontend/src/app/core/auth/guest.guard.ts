import { inject } from '@angular/core';
import { CanActivateFn, Router } from '@angular/router';

import { AuthSessionService } from './auth-session.service';
import { AuthStateService } from './auth-state.service';

export const guestGuard: CanActivateFn = () => {
  const authState = inject(AuthStateService);
  const authSession = inject(AuthSessionService);
  const router = inject(Router);

  if (authState.isAuthenticated() || authSession.hasAccessToken()) {
    return router.parseUrl('/dashboard');
  }

  return true;
};
