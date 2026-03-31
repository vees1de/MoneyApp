import { CommonModule } from '@angular/common';
import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { MatListModule } from '@angular/material/list';

import { DashboardApiService } from '@core/api/dashboard-api.service';
import { AuthStateService } from '@core/auth/auth-state.service';
import { WidgetShellComponent } from '@app/widgets/widget-shell/widget-shell.component';

interface TeamItem {
  id: string;
  name: string;
  position?: string | null;
}

@Component({
  selector: 'app-team-overview-widget',
  standalone: true,
  imports: [CommonModule, MatListModule, WidgetShellComponent],
  templateUrl: './team-overview-widget.component.html',
  styleUrl: './team-overview-widget.component.scss',
})
export class TeamOverviewWidgetComponent implements OnInit {
  private readonly dashboardApi = inject(DashboardApiService);
  private readonly authState = inject(AuthStateService);

  protected readonly loading = signal(true);
  protected readonly error = signal<string | null>(null);
  protected readonly teamSize = signal<number | null>(null);
  protected readonly teamMembers = signal<TeamItem[]>([]);

  protected readonly displayName = computed(() => this.authState.displayName());

  ngOnInit(): void {
    if (!this.authState.hasRole('manager')) {
      this.loading.set(false);
      return;
    }

    this.dashboardApi.getManagerDashboard().subscribe({
      next: (payload) => {
        this.teamSize.set(payload.stats?.team_size ?? null);
        this.teamMembers.set(
          (payload.team_preview ?? []).slice(0, 4).map((item) => ({
            id: item.user_id,
            name: `${item.first_name} ${item.last_name}`.trim(),
            position: item.position_title,
          })),
        );
        this.loading.set(false);
      },
      error: () => {
        this.error.set('failed');
        this.loading.set(false);
      },
    });
  }
}
