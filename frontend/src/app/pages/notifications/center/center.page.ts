import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { RouterLink } from '@angular/router';
import { NotificationsFacade } from '@features/notifications';
import type { Notification } from '@entities/notification';

@Component({
  selector: 'app-page-notifications-center',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './center.page.html',
  styleUrl: './center.page.scss',
})
export class NotificationsCenterPageComponent {
  private readonly facade = inject(NotificationsFacade);
  protected readonly routePath = '/notifications/center';
  protected readonly entitySample: Notification[] = [];

  protected loadPage(): void {
    this.facade.load();
  }
}
