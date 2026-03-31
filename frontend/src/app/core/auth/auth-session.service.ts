import { Injectable } from '@angular/core';

import type { AuthTokens, IdentityUserView } from './auth.types';

@Injectable({ providedIn: 'root' })
export class AuthSessionService {
  private readonly accessTokenKey = 'auth.access_token';
  private readonly refreshTokenKey = 'auth.refresh_token';
  private readonly expiresInKey = 'auth.expires_in';
  private readonly userSnapshotKey = 'auth.user_snapshot';

  setTokens(tokens: AuthTokens): void {
    localStorage.setItem(this.accessTokenKey, tokens.access_token);
    localStorage.setItem(this.refreshTokenKey, tokens.refresh_token);
    localStorage.setItem(this.expiresInKey, String(tokens.expires_in));
  }

  getAccessToken(): string | null {
    return localStorage.getItem(this.accessTokenKey);
  }

  getRefreshToken(): string | null {
    return localStorage.getItem(this.refreshTokenKey);
  }

  hasAccessToken(): boolean {
    return !!this.getAccessToken();
  }

  setUserSnapshot(user: IdentityUserView): void {
    localStorage.setItem(this.userSnapshotKey, JSON.stringify(user));
  }

  getUserSnapshot(): IdentityUserView | null {
    const raw = localStorage.getItem(this.userSnapshotKey);
    if (!raw) {
      return null;
    }

    try {
      return JSON.parse(raw) as IdentityUserView;
    } catch {
      this.removeUserSnapshot();
      return null;
    }
  }

  removeUserSnapshot(): void {
    localStorage.removeItem(this.userSnapshotKey);
  }

  clear(): void {
    localStorage.removeItem(this.accessTokenKey);
    localStorage.removeItem(this.refreshTokenKey);
    localStorage.removeItem(this.expiresInKey);
    localStorage.removeItem(this.userSnapshotKey);
  }
}
