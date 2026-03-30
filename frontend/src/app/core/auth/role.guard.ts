import { inject } from '@angular/core';
import { CanActivateFn, Router } from '@angular/router';

import { AuthStateService } from './auth-state.service';
import type { RoleCode } from './auth.types';

export const roleGuard = (allowedRoles: RoleCode[]): CanActivateFn => {
  return () => {
    const authState = inject(AuthStateService);
    const router = inject(Router);

    if (authState.hasAnyRole(allowedRoles)) {
      return true;
    }

    return router.parseUrl('/forbidden');
  };
};
