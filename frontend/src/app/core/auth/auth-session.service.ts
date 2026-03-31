import { Injectable } from '@angular/core';

import type { AuthTokens } from './auth.types';

@Injectable({ providedIn: 'root' })
export class AuthSessionService {
  private readonly accessTokenKey = 'auth.access_token';
  private readonly refreshTokenKey = 'auth.refresh_token';
  private readonly expiresInKey = 'auth.expires_in';

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

  clear(): void {
    localStorage.removeItem(this.accessTokenKey);
    localStorage.removeItem(this.refreshTokenKey);
    localStorage.removeItem(this.expiresInKey);
  }
}
