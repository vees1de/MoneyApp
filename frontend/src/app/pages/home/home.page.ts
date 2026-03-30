import { CommonModule } from '@angular/common';
import { Component, computed, inject } from '@angular/core';
import { RouterLink } from '@angular/router';

import { AuthStateService } from '@core/auth/auth-state.service';
import { AppFeature, UserRole } from '@core/auth/auth.types';

interface HomeLink {
  label: string;
  route: string;
  feature: AppFeature;
}

@Component({
  selector: 'app-page-home',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './home.page.html',
  styleUrl: './home.page.scss',
})
export class HomePageComponent {
  protected readonly authState = inject(AuthStateService);
  protected readonly user = this.authState.currentUser;
  protected readonly recentActions = this.authState.recentActions;

  private readonly roleDashboards: Record<UserRole, string> = {
    [UserRole.Administrator]: '/dashboard/admin',
    [UserRole.HrLnd]: '/dashboard/hr',
    [UserRole.Manager]: '/dashboard/manager',
    [UserRole.Employee]: '/dashboard/employee',
    [UserRole.Trainer]: '/dashboard/trainer',
  };

  private readonly allLinks: HomeLink[] = [
    { label: 'Внешние заявки', route: '/external-requests', feature: AppFeature.ExternalRequests },
    { label: 'Согласования', route: '/approvals/inbox', feature: AppFeature.Approvals },
    { label: 'Календарь', route: '/calendar', feature: AppFeature.Calendar },
    { label: 'Сертификаты', route: '/learning/certificates', feature: AppFeature.Certificates },
    { label: 'Отчеты', route: '/reports/overview', feature: AppFeature.Reports },
    { label: 'Университет', route: '/university/directions', feature: AppFeature.University },
    { label: 'Уведомления', route: '/notifications', feature: AppFeature.Notifications },
    { label: 'Профиль', route: '/profile', feature: AppFeature.Profile },
    { label: 'Админ: Users', route: '/admin/users', feature: AppFeature.AdminUsers },
    { label: 'Админ: Roles', route: '/admin/roles', feature: AppFeature.AdminRoles },
  ];

  protected readonly dashboardLinks = computed(() => {
    const user = this.user();
    if (!user) return [];

    return user.roles.map((role) => ({
      label: `Рабочий стол: ${role}`,
      route: this.roleDashboards[role],
    }));
  });

  protected readonly featureLinks = computed(() =>
    this.allLinks.filter((item) => this.authState.hasFeature(item.feature)),
  );
}
