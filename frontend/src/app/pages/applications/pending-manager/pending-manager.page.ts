import { CommonModule } from '@angular/common';
import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { Observable, catchError, forkJoin, of } from 'rxjs';

import { CourseApplicationsApiService } from '@core/api/course-applications-api.service';
import { CourseIntakesApiService } from '@core/api/course-intakes-api.service';
import { UsersApiService } from '@core/api/users-api.service';
import type { CourseApplication, CourseIntake } from '@core/api/contracts';
import { AuthStateService } from '@core/auth/auth-state.service';
import { PERMISSIONS } from '@core/auth/permissions';
import { canManagerReviewApplication, courseApplicationStatusLabel } from '@core/domain/course-intakes.workflow';
import { identityUserDisplayName } from '@core/domain/identity.util';
import type { IdentityUserView } from '@core/auth/auth.types';

@Component({
  selector: 'app-page-applications-pending-manager',
  standalone: true,
  imports: [
    CommonModule,
    RouterLink,
    MatButtonModule,
    MatCardModule,
    MatFormFieldModule,
    MatIconModule,
    MatInputModule,
  ],
  templateUrl: './pending-manager.page.html',
  styleUrl: './pending-manager.page.scss',
})
export class ApplicationsPendingManagerPageComponent implements OnInit {
  private readonly applicationsApi = inject(CourseApplicationsApiService);
  private readonly intakesApi = inject(CourseIntakesApiService);
  private readonly usersApi = inject(UsersApiService);
  private readonly authState = inject(AuthStateService);

  protected readonly loading = signal(true);
  protected readonly acting = signal(false);
  protected readonly error = signal<string | null>(null);
  protected readonly applications = signal<CourseApplication[]>([]);
  protected readonly intakesById = signal<Record<string, CourseIntake>>({});
  protected readonly usersById = signal<Record<string, IdentityUserView>>({});
  protected readonly comments = signal<Record<string, string>>({});

  protected readonly hasDirectoryAccess = computed(
    () =>
      this.authState.hasPermission(PERMISSIONS.usersRead) || this.authState.hasAnyRole(['hr', 'admin']),
  );

  ngOnInit(): void {
    this.load();
  }

  protected statusLabel(status: string): string {
    return courseApplicationStatusLabel(status);
  }

  protected intakeTitle(intakeId: string): string {
    return this.intakesById()[intakeId]?.title ?? `Набор ${intakeId.slice(0, 8)}`;
  }

  protected applicantLabel(applicantId: string): string {
    const user = this.usersById()[applicantId];
    return user ? identityUserDisplayName(user) : `Сотрудник ${applicantId.slice(0, 8)}`;
  }

  protected canReview(application: CourseApplication): boolean {
    return canManagerReviewApplication(application.status);
  }

  protected comment(id: string): string {
    return this.comments()[id] ?? '';
  }

  protected setComment(id: string, value: string): void {
    this.comments.update((state) => ({
      ...state,
      [id]: value,
    }));
  }

  protected approve(application: CourseApplication): void {
    this.makeDecision(application.id, 'approve');
  }

  protected reject(application: CourseApplication): void {
    this.makeDecision(application.id, 'reject');
  }

  private load(): void {
    this.loading.set(true);
    this.error.set(null);

    this.applicationsApi.listPendingManager().subscribe({
      next: (applications) => {
        this.applications.set(applications ?? []);

        const intakeIds = [...new Set((applications ?? []).map((item) => item.intake_id))];
        const applicantIds = [...new Set((applications ?? []).map((item) => item.applicant_id))];

        const intakeRequests: Record<string, Observable<CourseIntake | null>> = {};
        intakeIds.forEach((id) => {
          intakeRequests[id] = this.intakesApi.getById(id).pipe(catchError(() => of(null)));
        });

        forkJoin({
          intakes:
            intakeIds.length > 0 ? forkJoin(intakeRequests) : of<Record<string, CourseIntake | null>>({}),
          users: this.hasDirectoryAccess()
            ? this.usersApi.listAdminUsers().pipe(catchError(() => of([])))
            : of([]),
        }).subscribe({
          next: ({ intakes, users }) => {
            const mappedIntakes = Object.entries(intakes).reduce<Record<string, CourseIntake>>(
              (accumulator, [id, intake]) => {
                if (intake) {
                  accumulator[id] = intake;
                }
                return accumulator;
              },
              {},
            );

            const mappedUsers = (users ?? [])
              .filter((user) => applicantIds.includes(user.id))
              .reduce<Record<string, IdentityUserView>>((accumulator, user) => {
                accumulator[user.id] = user;
                return accumulator;
              }, {});

            this.intakesById.set(mappedIntakes);
            this.usersById.set(mappedUsers);
            this.loading.set(false);
          },
          error: () => {
            this.error.set('Не удалось загрузить связанные данные для согласования.');
            this.loading.set(false);
          },
        });
      },
      error: () => {
        this.error.set('Не удалось загрузить заявки на ваше согласование.');
        this.loading.set(false);
      },
    });
  }

  private makeDecision(applicationId: string, action: 'approve' | 'reject'): void {
    if (this.acting()) {
      return;
    }

    this.acting.set(true);
    this.error.set(null);

    const payload = { comment: normalizeText(this.comment(applicationId)) };
    const request$ =
      action === 'approve'
        ? this.applicationsApi.approveManager(applicationId, payload)
        : this.applicationsApi.rejectManager(applicationId, payload);

    request$.subscribe({
      next: (updated) => {
        this.applications.update((items) =>
          items.map((item) => (item.id === updated.id ? updated : item)),
        );
        this.acting.set(false);
      },
      error: () => {
        this.error.set(
          action === 'approve'
            ? 'Не удалось согласовать заявку руководителем.'
            : 'Не удалось отклонить заявку руководителем.',
        );
        this.acting.set(false);
      },
    });
  }
}

function normalizeText(value: string | null | undefined): string | undefined {
  const trimmed = value?.trim();
  return trimmed ? trimmed : undefined;
}
