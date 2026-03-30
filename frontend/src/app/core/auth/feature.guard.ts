import { inject } from '@angular/core';
import { CanActivateFn, Router } from '@angular/router';

import { AuthStateService } from './auth-state.service';
import { AppFeature } from './auth.types';

export const featureGuard = (feature: AppFeature): CanActivateFn => {
  return () => {
    const authState = inject(AuthStateService);
    const router = inject(Router);

    if (authState.hasFeature(feature)) {
      return true;
    }

    return router.parseUrl('/forbidden');
  };
};
