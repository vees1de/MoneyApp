import { CommonModule } from '@angular/common';
import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { RouterLink } from '@angular/router';
import { catchError, forkJoin, of, type Observable } from 'rxjs';

import { WidgetShellComponent } from '@app/widgets/widget-shell/widget-shell.component';
import { AuthStateService } from '@core/auth/auth-state.service';
import type { IdentityUserView } from '@core/auth/auth.types';
import { PERMISSIONS } from '@core/auth/permissions';
import { CourseApplicationsApiService } from '@core/api/course-applications-api.service';
import { CourseIntakesApiService } from '@core/api/course-intakes-api.service';
import type { CourseApplication, CourseIntake } from '@core/api/contracts';
import { UsersApiService } from '@core/api/users-api.service';
import { resolveApiUrl } from '@core/api/url.util';
import { canHrReviewApplication, courseApplicationStatusLabel } from '@core/domain/course-intakes.workflow';
import { identityUserDisplayName } from '@core/domain/identity.util';

@Component({
  selector: 'app-intake-hr-approvals-widget',
  standalone: true,
  imports: [
    CommonModule,
    RouterLink,
    MatButtonModule,
    MatFormFieldModule,
    MatIconModule,
    MatInputModule,
    WidgetShellComponent,
  ],
  templateUrl: './intake-hr-approvals-widget.component.html',
  styleUrl: './intake-hr-approvals-widget.component.scss',
})
export class IntakeHrApprovalsWidgetComponent implements OnInit {
  private readonly applicationsApi = inject(CourseApplicationsApiService);
  private readonly intakesApi = inject(CourseIntakesApiService);
  private readonly usersApi = inject(UsersApiService);
  private readonly authState = inject(AuthStateService);

  protected readonly loading = signal(true);
  protected readonly acting = signal(false);
  protected readonly error = signal<string | null>(null);
  protected readonly directoryUnavailable = signal(false);
  protected readonly applications = signal<CourseApplication[]>([]);
  protected readonly intakesById = signal<Record<string, CourseIntake>>({});
  protected readonly usersById = signal<Record<string, IdentityUserView>>({});
  protected readonly comments = signal<Record<string, string>>({});

  protected readonly hasDirectoryAccess = computed(
    () =>
      this.authState.hasPermission(PERMISSIONS.usersRead) ||
      this.authState.hasAnyRole(['hr', 'admin']),
  );

  protected readonly pendingApplications = computed(() =>
    [...this.applications()]
      .filter((application) => canHrReviewApplication(application.status))
      .sort(
        (left, right) =>
          new Date(right.created_at).getTime() - new Date(left.created_at).getTime(),
      ),
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

  protected userLabel(applicantId: string): string {
    const user = this.lookupUser(applicantId);
    return user ? identityUserDisplayName(user) : `Сотрудник ${applicantId.slice(0, 8)}`;
  }

  protected userPosition(applicantId: string): string | null {
    return normalizeText(this.lookupUser(applicantId)?.employee_profile?.position_title) ?? null;
  }

  protected userEmail(applicantId: string): string | null {
    return this.lookupUser(applicantId)?.email ?? null;
  }

  protected userAvatarUrl(applicantId: string): string | null {
    return resolveApiUrl(this.lookupUser(applicantId)?.avatar_url);
  }

  protected userAvatarInitials(applicantId: string): string {
    const user = this.lookupUser(applicantId);
    if (!user) {
      return 'U';
    }

    const profile = user.employee_profile;
    if (profile) {
      const initials = `${profile.last_name?.[0] ?? ''}${profile.first_name?.[0] ?? ''}`
        .trim()
        .toUpperCase();
      if (initials) {
        return initials;
      }
    }

    const emailInitial = user.email?.trim()?.[0]?.toUpperCase();
    return emailInitial || 'U';
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
    if (this.acting()) {
      return;
    }

    this.acting.set(true);
    this.error.set(null);

    this.applicationsApi
      .approveHr(application.id, {
        comment: normalizeText(this.comment(application.id)),
      })
      .subscribe({
        next: (approved) => {
          this.removeApplication(approved.id);

          this.applicationsApi.enroll(approved.id).subscribe({
            next: () => {
              this.acting.set(false);
            },
            error: () => {
              this.error.set(
                'Заявка согласована HR, но сотрудника не удалось добавить в набор.',
              );
              this.acting.set(false);
            },
          });
        },
        error: () => {
          this.error.set('Не удалось принять сотрудника в набор.');
          this.acting.set(false);
        },
      });
  }

  protected reject(application: CourseApplication): void {
    if (this.acting()) {
      return;
    }

    this.acting.set(true);
    this.error.set(null);

    this.applicationsApi
      .rejectHr(application.id, {
        comment: normalizeText(this.comment(application.id)),
      })
      .subscribe({
        next: (rejected) => {
          this.removeApplication(rejected.id);
          this.acting.set(false);
        },
        error: () => {
          this.error.set('Не удалось отклонить заявку HR.');
          this.acting.set(false);
        },
      });
  }

  private load(): void {
    this.loading.set(true);
    this.error.set(null);
    this.directoryUnavailable.set(false);

    const users$ = this.hasDirectoryAccess()
      ? this.usersApi.listAdminUsers().pipe(
          catchError(() => {
            this.directoryUnavailable.set(true);
            return of<IdentityUserView[]>([]);
          }),
        )
      : of<IdentityUserView[]>([]);

    forkJoin({
      intakes: this.intakesApi.list('open'),
      users: users$,
    }).subscribe({
      next: ({ intakes, users }) => {
        const openIntakes = intakes ?? [];
        this.intakesById.set(
          openIntakes.reduce<Record<string, CourseIntake>>((accumulator, intake) => {
            accumulator[intake.id] = intake;
            return accumulator;
          }, {}),
        );

        this.usersById.set(
          (users ?? []).reduce<Record<string, IdentityUserView>>((accumulator, user) => {
            accumulator[user.id] = user;
            return accumulator;
          }, {}),
        );

        if (openIntakes.length === 0) {
          this.applications.set([]);
          this.loading.set(false);
          return;
        }

        const requests: Record<string, Observable<CourseApplication[]>> = {};
        for (const intake of openIntakes) {
          requests[intake.id] = this.intakesApi
            .listApplications(intake.id)
            .pipe(catchError(() => of([])));
        }

        forkJoin(requests).subscribe({
          next: (applicationsByIntake) => {
            const allApplications = Object.values(applicationsByIntake).flatMap(
              (items) => items ?? [],
            );
            this.applications.set(allApplications);
            this.loading.set(false);
          },
          error: () => {
            this.error.set('Не удалось загрузить заявки в открытые наборы.');
            this.loading.set(false);
          },
        });
      },
      error: () => {
        this.error.set('Не удалось загрузить открытые наборы.');
        this.loading.set(false);
      },
    });
  }

  private lookupUser(applicantId: string): IdentityUserView | null {
    const currentUser = this.authState.currentUser();
    if (currentUser?.id === applicantId) {
      return currentUser;
    }

    return this.usersById()[applicantId] ?? null;
  }

  private removeApplication(applicationId: string): void {
    this.applications.update((items) => items.filter((item) => item.id !== applicationId));
  }
}

function normalizeText(value: string | null | undefined): string | undefined {
  const trimmed = value?.trim();
  return trimmed ? trimmed : undefined;
}
