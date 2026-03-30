import { inject } from '@angular/core';
import { CanActivateFn, Router } from '@angular/router';

import { AuthStateService } from './auth-state.service';
import type { PermissionCode } from './permissions';

export const permissionGuard = (permissions: PermissionCode[]): CanActivateFn => {
  return () => {
    const authState = inject(AuthStateService);
    const router = inject(Router);

    if (authState.hasAnyPermission(permissions)) {
      return true;
    }

    return router.parseUrl('/forbidden');
  };
};

// Backward-compatible alias for existing route config naming.
export const featureGuard = (permission: PermissionCode): CanActivateFn => {
  return permissionGuard([permission]);
};
