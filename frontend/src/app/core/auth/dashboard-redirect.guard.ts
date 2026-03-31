import { inject } from '@angular/core';
import { CanActivateFn, Router } from '@angular/router';

import { AuthStateService } from './auth-state.service';
import type { RoleCode } from './auth.types';

const ROLE_DASHBOARD_PRIORITY: { role: RoleCode; route: string }[] = [
  { role: 'admin', route: '/dashboard/admin' },
  { role: 'hr', route: '/dashboard/hr' },
  { role: 'manager', route: '/dashboard/manager' },
  { role: 'trainer', route: '/dashboard/trainer' },
  { role: 'employee', route: '/dashboard/employee' },
];

export const dashboardRedirectGuard: CanActivateFn = () => {
  const authState = inject(AuthStateService);
  const router = inject(Router);
  const user = authState.currentUser();

  if (!user) {
    return router.parseUrl('/login');
  }

  const primaryRole = user.roles[0];
  const match = ROLE_DASHBOARD_PRIORITY.find(({ role }) => role === primaryRole);
  if (!match) {
    return router.parseUrl('/forbidden');
  }

  return router.parseUrl(match.route);
};
