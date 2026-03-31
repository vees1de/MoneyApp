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

    try {
      const me = await firstValueFrom(this.authApi.me());
      this.authState.setCurrentUserFromMe(me);
    } catch {
      this.authSession.clear();
      this.authState.setCurrentUser(null);
    }
  }
}
