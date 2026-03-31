import { Injectable, computed, signal } from '@angular/core';

import type { IdentityUserView, MeResponse, RecentAction, RoleCode } from './auth.types';
import type { PermissionCode } from './permissions';

@Injectable({ providedIn: 'root' })
export class AuthStateService {
  private readonly currentUserSignal = signal<IdentityUserView | null>(null);

  private readonly recentActionsSignal = signal<RecentAction[]>([
    {
      id: 'ra-1',
      label: 'Opened external requests list',
      route: '/external-requests',
      at: 'just now',
    },
    {
      id: 'ra-2',
      label: 'Opened analytics dashboard',
      route: '/reports/overview',
      at: '5m ago',
    },
  ]);

  readonly currentUser = computed(() => this.currentUserSignal());
  readonly isAuthenticated = computed(() => this.currentUserSignal() !== null);
  readonly recentActions = computed(() => this.recentActionsSignal());
  readonly displayName = computed(() => {
    const profile = this.currentUserSignal()?.employee_profile;
    if (!profile) return 'User';
    return `${profile.first_name} ${profile.last_name}`.trim();
  });

  setCurrentUser(user: IdentityUserView | null): void {
    this.currentUserSignal.set(user);
  }

  setCurrentUserFromMe(payload: MeResponse): void {
    this.currentUserSignal.set(payload.user);
  }

  addRecentAction(action: RecentAction): void {
    this.recentActionsSignal.update((state) => [action, ...state].slice(0, 12));
  }

  hasRole(role: RoleCode): boolean {
    const user = this.currentUserSignal();
    return !!user && user.roles.includes(role);
  }

  hasAnyRole(roles: RoleCode[]): boolean {
    const user = this.currentUserSignal();
    return !!user && roles.some((role) => user.roles.includes(role));
  }

  hasPermission(permission: PermissionCode): boolean {
    const user = this.currentUserSignal();
    return !!user && user.permissions.includes(permission);
  }

  hasAnyPermission(permissions: PermissionCode[]): boolean {
    const user = this.currentUserSignal();
    return !!user && permissions.some((permission) => user.permissions.includes(permission));
  }

  roleLabel(role: RoleCode): string {
    const roleMap: Record<string, string> = {
      admin: 'Administrator',
      hr: 'HR / L&D',
      manager: 'Manager',
      employee: 'Employee',
      trainer: 'Trainer',
    };
    return roleMap[role] ?? role;
  }
}
