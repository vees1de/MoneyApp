import { CommonModule } from '@angular/common';
import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { catchError, forkJoin, of } from 'rxjs';

import { CourseApplicationsApiService } from '@core/api/course-applications-api.service';
import { CourseIntakesApiService } from '@core/api/course-intakes-api.service';
import { CourseSuggestionsApiService } from '@core/api/course-suggestions-api.service';
import { LearningPlanApiService } from '@core/api/learning-plan-api.service';
import { AuthStateService } from '@core/auth/auth-state.service';
import { PERMISSIONS } from '@core/auth/permissions';
import { COURSE_INTAKE_MANAGER_APPROVAL_ENABLED } from '@core/config/feature-flags';
import { isIntakeManageRole } from '@core/domain/course-intakes.workflow';
import { WidgetShellComponent } from '@app/widgets/widget-shell/widget-shell.component';

@Component({
  selector: 'app-quick-actions-widget',
  standalone: true,
  imports: [CommonModule, RouterLink, MatButtonModule, MatIconModule, WidgetShellComponent],
  templateUrl: './quick-actions-widget.component.html',
  styleUrl: './quick-actions-widget.component.scss',
})
export class QuickActionsWidgetComponent implements OnInit {
  private readonly learningPlanApi = inject(LearningPlanApiService);
  private readonly applicationsApi = inject(CourseApplicationsApiService);
  private readonly intakesApi = inject(CourseIntakesApiService);
  private readonly suggestionsApi = inject(CourseSuggestionsApiService);
  private readonly authState = inject(AuthStateService);

  protected readonly loading = signal(true);
  protected readonly error = signal<string | null>(null);
  protected readonly activeLearning = signal(0);
  protected readonly myApplications = signal(0);
  protected readonly openIntakes = signal(0);
  protected readonly mySuggestions = signal(0);

  protected readonly canManageIntakes = computed(() => {
    const role = this.authState.currentUser()?.roles[0] ?? 'employee';
    return isIntakeManageRole(role) || this.authState.hasPermission(PERMISSIONS.intakesManage);
  });

  protected readonly canReviewManager = computed(
    () => COURSE_INTAKE_MANAGER_APPROVAL_ENABLED && this.authState.hasAnyRole(['manager', 'admin']),
  );

  ngOnInit(): void {
    forkJoin({
      plan: this.learningPlanApi.getMyPlan().pipe(catchError(() => of(null))),
      applications: this.applicationsApi.listMy().pipe(catchError(() => of([]))),
      intakes: this.intakesApi.list('open').pipe(catchError(() => of([]))),
      suggestions: this.suggestionsApi.listMy().pipe(catchError(() => of([]))),
    }).subscribe({
      next: ({ plan, applications, intakes, suggestions }) => {
        this.activeLearning.set(plan?.summary?.in_progress ?? 0);
        this.myApplications.set((applications ?? []).length);
        this.openIntakes.set((intakes ?? []).length);
        this.mySuggestions.set((suggestions ?? []).length);
        this.loading.set(false);
      },
      error: () => {
        this.error.set('failed');
        this.loading.set(false);
      },
    });
  }
}
