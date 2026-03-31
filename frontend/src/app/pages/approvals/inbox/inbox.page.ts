import { CommonModule } from '@angular/common';
import { Component, OnInit, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { MatCardModule } from '@angular/material/card';

import { ExternalRequestsApiService } from '@core/api/external-requests-api.service';
import type { PendingApprovalItem } from '@core/api/contracts';
import { externalRequestStatusLabel } from '@core/domain/external-request.workflow';

@Component({
  selector: 'app-page-approvals-inbox',
  standalone: true,
  imports: [CommonModule, RouterLink, MatCardModule],
  templateUrl: './inbox.page.html',
  styleUrl: './inbox.page.scss',
})
export class ApprovalsInboxPageComponent implements OnInit {
  private readonly api = inject(ExternalRequestsApiService);

  protected readonly loading = signal(true);
  protected readonly error = signal<string | null>(null);
  protected readonly items = signal<PendingApprovalItem[]>([]);

  ngOnInit(): void {
    this.api.listPendingApprovals().subscribe({
      next: (items) => {
        this.items.set(items ?? []);
        this.loading.set(false);
      },
      error: () => {
        this.error.set('Не удалось загрузить входящие согласования');
        this.loading.set(false);
      },
    });
  }

  protected statusLabel(status: string): string {
    return externalRequestStatusLabel(status);
  }
}
