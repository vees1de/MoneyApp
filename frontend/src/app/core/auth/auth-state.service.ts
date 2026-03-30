import { Injectable, computed, signal } from '@angular/core';

import { AppFeature, AuthUser, RecentAction, UserRole } from './auth.types';

@Injectable({ providedIn: 'root' })
export class AuthStateService {
  private readonly currentUserSignal = signal<AuthUser | null>({
    id: 'u-001',
    fullName: 'Demo User',
    roles: [
      UserRole.Employee,
      UserRole.Manager,
      UserRole.HrLnd,
      UserRole.Trainer,
      UserRole.Administrator,
    ],
    features: [
      AppFeature.Dashboard,
      AppFeature.ExternalRequests,
      AppFeature.Approvals,
      AppFeature.Calendar,
      AppFeature.Certificates,
      AppFeature.Reports,
      AppFeature.University,
      AppFeature.Notifications,
      AppFeature.Profile,
      AppFeature.Catalog,
      AppFeature.AdminUsers,
      AppFeature.AdminRoles,
    ],
  });

  private readonly recentActionsSignal = signal<RecentAction[]>([
    {
      id: 'ra-1',
      label: 'Открыт реестр заявок на внешние курсы',
      route: '/external-requests',
      at: 'just now',
    },
    {
      id: 'ra-2',
      label: 'Просмотрен дашборд подразделения',
      route: '/reports/overview',
      at: '5m ago',
    },
  ]);

  readonly currentUser = computed(() => this.currentUserSignal());
  readonly isAuthenticated = computed(() => this.currentUserSignal() !== null);
  readonly recentActions = computed(() => this.recentActionsSignal());

  setCurrentUser(user: AuthUser | null): void {
    this.currentUserSignal.set(user);
  }

  addRecentAction(action: RecentAction): void {
    this.recentActionsSignal.update((state) => [action, ...state].slice(0, 12));
  }

  hasRole(role: UserRole): boolean {
    const user = this.currentUserSignal();
    return !!user && user.roles.includes(role);
  }

  hasAnyRole(roles: UserRole[]): boolean {
    const user = this.currentUserSignal();
    return !!user && roles.some((role) => user.roles.includes(role));
  }

  hasFeature(feature: AppFeature): boolean {
    const user = this.currentUserSignal();
    return !!user && user.features.includes(feature);
  }
}
