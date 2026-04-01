import { CommonModule } from '@angular/common';
import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';

import { AnalyticsApiService } from '@core/api/analytics-api.service';
import type { HrRiskOverdueItem, HrRisksResponse } from '@core/api/contracts';
import { WidgetShellComponent } from '@app/widgets/widget-shell/widget-shell.component';

interface RiskSummary {
  key: string;
  label: string;
  count: number;
  icon: string;
  severity: 'critical' | 'warning' | 'info';
}

@Component({
  selector: 'app-risks-widget',
  standalone: true,
  imports: [CommonModule, RouterLink, MatButtonModule, MatIconModule, WidgetShellComponent],
  templateUrl: './risks-widget.component.html',
  styleUrl: './risks-widget.component.scss',
})
export class RisksWidgetComponent implements OnInit {
  private readonly analyticsApi = inject(AnalyticsApiService);

  protected readonly loading = signal(true);
  protected readonly error = signal<string | null>(null);
  protected readonly risks = signal<RiskSummary[]>([]);
  protected readonly overdueItems = signal<HrRiskOverdueItem[]>([]);
  protected readonly selectedTab = signal<string>('summary');

  protected readonly totalRisks = computed(() =>
    this.risks().reduce((sum, r) => sum + r.count, 0),
  );

  protected readonly tabs = [
    { key: 'summary', label: 'Обзор' },
    { key: 'overdue', label: 'Просроченные' },
  ];

  ngOnInit(): void {
    this.analyticsApi.getRisks().subscribe({
      next: (data) => {
        this.risks.set(this.mapRisks(data));
        this.overdueItems.set(data.overdue_items ?? []);
        this.loading.set(false);
      },
      error: () => {
        this.error.set('failed');
        this.loading.set(false);
      },
    });
  }

  protected setTab(tab: string): void {
    this.selectedTab.set(tab);
  }

  protected daysOverdue(deadlineAt: string): number {
    const diff = Date.now() - new Date(deadlineAt).getTime();
    return Math.max(1, Math.floor(diff / (1000 * 60 * 60 * 24)));
  }

  private mapRisks(data: HrRisksResponse): RiskSummary[] {
    return [
      {
        key: 'overdue',
        label: 'Просроченные курсы',
        count: data.overdue_enrollments,
        icon: 'error_outline',
        severity: 'critical',
      },
      {
        key: 'deadline_soon',
        label: 'Дедлайн в течение 7 дней',
        count: data.deadline_soon,
        icon: 'schedule',
        severity: 'warning',
      },
      {
        key: 'inactive',
        label: 'Неактивные (30+ дней)',
        count: data.inactive_learners,
        icon: 'person_off',
        severity: 'warning',
      },
      {
        key: 'low_completion',
        label: 'Низкий прогресс к дедлайну',
        count: data.low_completion,
        icon: 'trending_down',
        severity: 'info',
      },
    ];
  }
}
