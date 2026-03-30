import { CommonModule, Location } from '@angular/common';
import { Component, computed, inject, signal } from '@angular/core';
import { Router, RouterLink, RouterOutlet } from '@angular/router';
import { MatBadgeModule } from '@angular/material/badge';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatToolbarModule } from '@angular/material/toolbar';

import { AuthStateService } from '@core/auth/auth-state.service';

interface HeaderNotification {
  id: string;
  text: string;
  route: string;
  read: boolean;
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
  private readonly location = inject(Location);
  private readonly router = inject(Router);
  protected readonly authState = inject(AuthStateService);
  protected readonly notificationsOpen = signal(false);
  protected readonly notifications = signal<HeaderNotification[]>([
    {
      id: 'n-1',
      text: 'Новая заявка на согласование',
      route: '/approvals/inbox',
      read: false,
    },
    {
      id: 'n-2',
      text: 'Дедлайн курса через 3 дня',
      route: '/my-learning',
      read: false,
    },
  ]);
  protected readonly unreadCount = computed(
    () => this.notifications().filter((item) => !item.read).length,
  );

  protected get showBackButton(): boolean {
    return !this.router.url.startsWith('/dashboard/test-role');
  }

  protected toggleNotifications(): void {
    this.notificationsOpen.update((state) => !state);
  }

  protected markAsRead(id: string): void {
    this.notifications.update((state) =>
      state.map((item) => (item.id === id ? { ...item, read: true } : item)),
    );
  }

  protected goBack(): void {
    this.location.back();
  }
}
