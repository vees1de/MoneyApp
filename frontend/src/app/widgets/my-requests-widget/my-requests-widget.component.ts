import { CommonModule } from '@angular/common';
import { Component, OnInit, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { MatListModule } from '@angular/material/list';

import { ExternalRequestsApiService } from '@core/api/external-requests-api.service';
import type { ExternalRequest } from '@core/api/contracts';
import { WidgetShellComponent } from '@app/widgets/widget-shell/widget-shell.component';

@Component({
  selector: 'app-my-requests-widget',
  standalone: true,
  imports: [CommonModule, RouterLink, MatListModule, WidgetShellComponent],
  templateUrl: './my-requests-widget.component.html',
  styleUrl: './my-requests-widget.component.scss',
})
export class MyRequestsWidgetComponent implements OnInit {
  private readonly api = inject(ExternalRequestsApiService);

  protected readonly loading = signal(true);
  protected readonly error = signal<string | null>(null);
  protected readonly requests = signal<ExternalRequest[]>([]);

  ngOnInit(): void {
    this.api.listMy().subscribe({
      next: (requests) => {
        this.requests.set((requests ?? []).slice(0, 3));
        this.loading.set(false);
      },
      error: () => {
        this.error.set('failed');
        this.loading.set(false);
      },
    });
  }
}
