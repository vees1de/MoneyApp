import { CommonModule } from '@angular/common';
import { Component, OnInit, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { MatListModule } from '@angular/material/list';

import { NotificationsApiService } from '@core/api/notifications-api.service';
import type { AppNotification } from '@core/api/contracts';
import { WidgetShellComponent } from '@app/widgets/widget-shell/widget-shell.component';

@Component({
  selector: 'app-work-activity-widget',
  standalone: true,
  imports: [CommonModule, RouterLink, MatListModule, WidgetShellComponent],
  templateUrl: './work-activity-widget.component.html',
  styleUrl: './work-activity-widget.component.scss',
})
export class WorkActivityWidgetComponent implements OnInit {
  private readonly api = inject(NotificationsApiService);

  protected readonly loading = signal(true);
  protected readonly error = signal<string | null>(null);
  protected readonly notifications = signal<AppNotification[]>([]);

  ngOnInit(): void {
    this.api.list().subscribe({
      next: (notifications) => {
        this.notifications.set((notifications ?? []).slice(0, 3));
        this.loading.set(false);
      },
      error: () => {
        this.error.set('failed');
        this.loading.set(false);
      },
    });
  }

  protected routeFor(item: AppNotification): string {
    if (item.related_entity_type === 'external_request' && item.related_entity_id) {
      return `/external-requests/${item.related_entity_id}`;
    }
    return '/notifications';
  }
}
