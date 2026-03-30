import { CommonModule } from '@angular/common';
import { Component, computed, inject } from '@angular/core';
import { RouterLink, RouterLinkActive } from '@angular/router';

import { AuthStateService } from '@core/auth/auth-state.service';
import { AppFeature } from '@core/auth/auth.types';

interface NavItem {
  label: string;
  route: string;
  feature: AppFeature;
}

@Component({
  selector: 'app-sidebar',
  standalone: true,
  imports: [CommonModule, RouterLink, RouterLinkActive],
  templateUrl: './sidebar.component.html',
  styleUrl: './sidebar.component.scss',
})
export class SidebarComponent {
  private readonly authState = inject(AuthStateService);

  private readonly allItems: NavItem[] = [
    { label: 'Home', route: '/home', feature: AppFeature.Dashboard },
    { label: 'Dashboard', route: '/dashboard', feature: AppFeature.Dashboard },
    {
      label: 'Внешние заявки',
      route: '/external-requests',
      feature: AppFeature.ExternalRequests,
    },
    { label: 'Согласования', route: '/approvals/inbox', feature: AppFeature.Approvals },
    { label: 'Календарь', route: '/calendar', feature: AppFeature.Calendar },
    { label: 'Каталог', route: '/catalog', feature: AppFeature.Catalog },
    { label: 'Сертификаты', route: '/learning/certificates', feature: AppFeature.Certificates },
    { label: 'Отчеты', route: '/reports/overview', feature: AppFeature.Reports },
    { label: 'Университет', route: '/university/directions', feature: AppFeature.University },
    {
      label: 'Уведомления',
      route: '/notifications',
      feature: AppFeature.Notifications,
    },
    { label: 'Профиль', route: '/profile', feature: AppFeature.Profile },
    { label: 'Админ: Users', route: '/admin/users', feature: AppFeature.AdminUsers },
    { label: 'Админ: Roles', route: '/admin/roles', feature: AppFeature.AdminRoles },
  ];

  protected readonly navItems = computed(() =>
    this.allItems.filter((item) => this.authState.hasFeature(item.feature)),
  );
}
