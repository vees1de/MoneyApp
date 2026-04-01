import { CommonModule } from '@angular/common';
import { Component, computed, inject, signal } from '@angular/core';
import { MatBadgeModule } from '@angular/material/badge';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatToolbarModule } from '@angular/material/toolbar';
import { Router, RouterLink, RouterOutlet } from '@angular/router';

import { AuthStateService } from '@core/auth/auth-state.service';
import { PERMISSIONS } from '@core/auth/permissions';

interface HeaderNotification {
  id: string;
  text: string;
  route: string;
  read: boolean;
}

interface HeaderNavItem {
  label: string;
  route: string;
}

@Component({
  selector: 'app-shell',
  standalone: true,
  imports: [
    CommonModule,
    RouterOutlet,
    RouterLink,
    MatBadgeModule,
    MatButtonModule,
    MatIconModule,
    MatToolbarModule,
  ],
  templateUrl: './app-shell.component.html',
  styleUrl: './app-shell.component.scss',
})
export class AppShellComponent {
  private readonly router = inject(Router);

  protected readonly authState = inject(AuthStateService);
  protected readonly notificationsOpen = signal(false);
  protected readonly navItems = computed<HeaderNavItem[]>(() => {
    const items: Array<HeaderNavItem & { visible: boolean }> = [
      {
        label: 'Курсы',
        route: '/catalog',
        visible: this.authState.hasPermission(PERMISSIONS.coursesRead),
      },
      {
        label: 'Назначения',
        route: '/intakes',
        visible: true,
      },
      {
        label: 'Заявки',
        route: '/suggestions',
        visible: true,
      },
      {
        label: 'Мое обучение',
        route: '/my-learning',
        visible: this.authState.hasAnyPermission([
          PERMISSIONS.enrollmentsRead,
          PERMISSIONS.enrollmentsManage,
        ]),
      },
      {
        label: 'Календарь',
        route: '/calendar/overview',
        visible: this.authState.hasPermission(PERMISSIONS.enrollmentsRead),
      },
      {
        label: 'Аналитика',
        route: '/reports/overview',
        visible: this.authState.hasAnyPermission([
          PERMISSIONS.analyticsReadHr,
          PERMISSIONS.analyticsReadManager,
        ]),
      },
    ];

    return items.filter((item) => item.visible);
  });

  protected readonly notifications = signal<HeaderNotification[]>([
    {
      id: 'n-1',
      text: 'Новая заявка ждет согласования',
      route: '/approvals/inbox',
      read: false,
    },
    {
      id: 'n-2',
      text: 'Новый курс доступен в каталоге',
      route: '/catalog',
      read: false,
    },
  ]);

  protected readonly unreadCount = computed(
    () => this.notifications().filter((item) => !item.read).length,
  );
  protected readonly showTestAdminLink = computed(() => this.authState.hasAnyRole(['admin', 'hr']));

  protected toggleNotifications(): void {
    this.notificationsOpen.update((state) => !state);
  }

  protected isActive(route: string): boolean {
    if (route === '/applications/my') {
      return this.router.url.startsWith('/applications');
    }

    return this.router.url.startsWith(route);
  }

  protected markAsRead(id: string): void {
    this.notifications.update((state) =>
      state.map((item) => (item.id === id ? { ...item, read: true } : item)),
    );
  }
}
