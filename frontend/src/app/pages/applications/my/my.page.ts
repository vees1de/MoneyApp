import { CommonModule } from '@angular/common';
import { Component, OnInit, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatIconModule } from '@angular/material/icon';
import { Observable, catchError, forkJoin, of } from 'rxjs';

import { CourseApplicationsApiService } from '@core/api/course-applications-api.service';
import { CourseIntakesApiService } from '@core/api/course-intakes-api.service';
import type { CourseApplication, CourseIntake } from '@core/api/contracts';
import {
  canWithdrawApplication,
  courseApplicationPaymentStatusLabel,
  courseApplicationStatusLabel,
  courseIntakeStatusLabel,
} from '@core/domain/course-intakes.workflow';

@Component({
  selector: 'app-page-applications-my',
  standalone: true,
  imports: [CommonModule, RouterLink, MatButtonModule, MatCardModule, MatIconModule],
  templateUrl: './my.page.html',
  styleUrl: './my.page.scss',
})
export class ApplicationsMyPageComponent implements OnInit {
  private readonly applicationsApi = inject(CourseApplicationsApiService);
  private readonly intakesApi = inject(CourseIntakesApiService);

  protected readonly loading = signal(true);
  protected readonly acting = signal(false);
  protected readonly error = signal<string | null>(null);
  protected readonly applications = signal<CourseApplication[]>([]);
  protected readonly intakesById = signal<Record<string, CourseIntake>>({});

  ngOnInit(): void {
    this.load();
  }

  protected applicationStatusLabel(status: string): string {
    return courseApplicationStatusLabel(status);
  }

  protected paymentStatusLabel(status: string): string {
    return courseApplicationPaymentStatusLabel(status);
  }

  protected intakeStatusLabel(status: string | null | undefined): string {
    return status ? courseIntakeStatusLabel(status) : '—';
  }

  protected canWithdraw(application: CourseApplication): boolean {
    return canWithdrawApplication(application.status);
  }

  protected intakeTitle(intakeId: string): string {
    return this.intakesById()[intakeId]?.title ?? `Набор ${intakeId.slice(0, 8)}`;
  }

  protected intakeStatus(intakeId: string): string | null {
    return this.intakesById()[intakeId]?.status ?? null;
  }

  protected withdraw(application: CourseApplication): void {
    if (this.acting()) {
      return;
    }

    this.acting.set(true);
    this.error.set(null);

    this.applicationsApi.withdraw(application.id).subscribe({
      next: (updated) => {
        this.applications.update((items) =>
          items.map((item) => (item.id === updated.id ? updated : item)),
        );
        this.acting.set(false);
      },
      error: () => {
        this.error.set('Не удалось отозвать заявку.');
        this.acting.set(false);
      },
    });
  }

  private load(): void {
    this.loading.set(true);
    this.error.set(null);

    this.applicationsApi.listMy().subscribe({
      next: (applications) => {
        this.applications.set(applications ?? []);

        const intakeIds = [...new Set((applications ?? []).map((item) => item.intake_id))];
        if (intakeIds.length === 0) {
          this.intakesById.set({});
          this.loading.set(false);
          return;
        }

        const requests: Record<string, Observable<CourseIntake | null>> = {};
        intakeIds.forEach((id) => {
          requests[id] = this.intakesApi.getById(id).pipe(catchError(() => of(null)));
        });

        forkJoin(requests).subscribe({
          next: (intakes) => {
            const mapped = Object.entries(intakes).reduce<Record<string, CourseIntake>>(
              (accumulator, [id, intake]) => {
                if (intake) {
                  accumulator[id] = intake;
                }
                return accumulator;
              },
              {},
            );
            this.intakesById.set(mapped);
            this.loading.set(false);
          },
          error: () => {
            this.error.set('Не удалось загрузить связанные наборы.');
            this.loading.set(false);
          },
        });
      },
      error: () => {
        this.error.set('Не удалось загрузить мои заявки.');
        this.loading.set(false);
      },
    });
  }
}
