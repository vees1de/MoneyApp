import { CommonModule } from '@angular/common';
import { Component, OnInit, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { MatCardModule } from '@angular/material/card';
import { MatProgressBarModule } from '@angular/material/progress-bar';

import { EnrollmentsApiService } from '@core/api/enrollments-api.service';
import type { Enrollment } from '@entities/enrollment';

@Component({
  selector: 'app-page-my-learning-list',
  standalone: true,
  imports: [CommonModule, RouterLink, MatCardModule, MatProgressBarModule],
  templateUrl: './list.page.html',
  styleUrl: './list.page.scss',
})
export class MyLearningListPageComponent implements OnInit {
  private readonly api = inject(EnrollmentsApiService);

  protected readonly loading = signal(true);
  protected readonly error = signal<string | null>(null);
  protected readonly enrollments = signal<Enrollment[]>([]);

  ngOnInit(): void {
    this.api.listMy().subscribe({
      next: (enrollments) => {
        this.enrollments.set(enrollments ?? []);
        this.loading.set(false);
      },
      error: () => {
        this.error.set('Не удалось загрузить обучение');
        this.loading.set(false);
      },
    });
  }

  protected progressValue(item: Enrollment): number {
    const value = Number(item.completion_percent);
    return Number.isNaN(value) ? 0 : Math.max(0, Math.min(100, value));
  }
}
