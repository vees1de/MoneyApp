import { CommonModule } from '@angular/common';
import { Component, OnInit, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatProgressBarModule } from '@angular/material/progress-bar';

import { EnrollmentsApiService } from '@core/api/enrollments-api.service';
import type { Enrollment } from '@entities/enrollment';

@Component({
  selector: 'app-page-learning-enrollment-detail',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    RouterLink,
    MatButtonModule,
    MatCardModule,
    MatFormFieldModule,
    MatInputModule,
    MatProgressBarModule,
  ],
  templateUrl: './enrollment-detail.page.html',
  styleUrl: './enrollment-detail.page.scss',
})
export class LearningEnrollmentDetailPageComponent implements OnInit {
  private readonly api = inject(EnrollmentsApiService);
  private readonly route = inject(ActivatedRoute);
  private readonly fb = inject(FormBuilder);

  protected readonly loading = signal(true);
  protected readonly acting = signal(false);
  protected readonly error = signal<string | null>(null);
  protected readonly enrollment = signal<Enrollment | null>(null);

  protected readonly progressForm = this.fb.group({
    course_module_id: ['', [Validators.required]],
    progress_percent: [50, [Validators.required]],
  });

  ngOnInit(): void {
    this.load();
  }

  protected progressValue(item: Enrollment): number {
    const value = Number(item.completion_percent);
    return Number.isNaN(value) ? 0 : Math.max(0, Math.min(100, value));
  }

  protected start(): void {
    const item = this.enrollment();
    if (!item || this.acting()) return;

    this.acting.set(true);
    this.api.start(item.id).subscribe({
      next: (updated) => {
        this.enrollment.set(updated);
        this.acting.set(false);
      },
      error: () => {
        this.error.set('Не удалось стартовать обучение');
        this.acting.set(false);
      },
    });
  }

  protected updateProgress(): void {
    const item = this.enrollment();
    if (!item || this.acting() || this.progressForm.invalid) return;

    this.acting.set(true);
    const v = this.progressForm.getRawValue();

    this.api
      .updateProgress(item.id, {
        course_module_id: v.course_module_id,
        status: 'in_progress',
        progress_percent: String(v.progress_percent ?? 0),
      })
      .subscribe({
        next: (updated) => {
          this.enrollment.set(updated);
          this.acting.set(false);
        },
        error: () => {
          this.error.set('Не удалось обновить прогресс');
          this.acting.set(false);
        },
      });
  }

  protected complete(): void {
    const item = this.enrollment();
    if (!item || this.acting()) return;

    this.acting.set(true);
    this.api.complete(item.id, { completion_type: 'manual', notes: 'MVP complete' }).subscribe({
      next: (updated) => {
        this.enrollment.set(updated);
        this.acting.set(false);
      },
      error: () => {
        this.error.set('Не удалось завершить обучение');
        this.acting.set(false);
      },
    });
  }

  private load(): void {
    const id = this.route.snapshot.paramMap.get('enrollmentId');
    if (!id) {
      this.error.set('Не найден enrollmentId');
      this.loading.set(false);
      return;
    }

    this.api.getById(id).subscribe({
      next: (enrollment) => {
        this.enrollment.set(enrollment);
        this.loading.set(false);
      },
      error: () => {
        this.error.set('Не удалось загрузить обучение');
        this.loading.set(false);
      },
    });
  }
}
