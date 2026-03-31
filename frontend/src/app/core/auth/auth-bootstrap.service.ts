import { HttpErrorResponse } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { firstValueFrom } from 'rxjs';

import { AuthApiService } from './auth-api.service';
import { AuthSessionService } from './auth-session.service';
import { AuthStateService } from './auth-state.service';

@Injectable({ providedIn: 'root' })
export class AuthBootstrapService {
  private readonly authApi = inject(AuthApiService);
  private readonly authSession = inject(AuthSessionService);
  private readonly authState = inject(AuthStateService);

  async bootstrap(): Promise<void> {
    const accessToken = this.authSession.getAccessToken();
    if (!accessToken) {
      this.authState.setCurrentUser(null);
      return;
    }

    const snapshot = this.authSession.getUserSnapshot();
    if (snapshot) {
      this.authState.setCurrentUser(snapshot);
    }

    try {
      const me = await firstValueFrom(this.authApi.me());
      this.authState.setCurrentUserFromMe(me);
      this.authSession.setUserSnapshot(me.user);
    } catch (error: unknown) {
      if (error instanceof HttpErrorResponse && error.status === 401) {
        this.authSession.clear();
        this.authState.setCurrentUser(null);
        return;
      }

      // Keep last known session snapshot for transient network/backend errors.
      if (!snapshot) {
        this.authState.setCurrentUser(null);
      }
    }
  }
}
