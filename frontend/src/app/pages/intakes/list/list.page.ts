import { CommonModule } from '@angular/common';
import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';

import { CourseIntakesApiService } from '@core/api/course-intakes-api.service';
import type { CourseIntake } from '@core/api/contracts';
import { AuthStateService } from '@core/auth/auth-state.service';
import { PERMISSIONS } from '@core/auth/permissions';
import { courseIntakeStatusLabel, isIntakeManageRole } from '@core/domain/course-intakes.workflow';

@Component({
  selector: 'app-page-intakes-list',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    RouterLink,
    MatButtonModule,
    MatCardModule,
    MatFormFieldModule,
    MatIconModule,
    MatInputModule,
    MatSelectModule,
  ],
  templateUrl: './list.page.html',
  styleUrl: './list.page.scss',
})
export class IntakesListPageComponent implements OnInit {
  private readonly api = inject(CourseIntakesApiService);
  private readonly authState = inject(AuthStateService);
  private readonly fb = inject(FormBuilder);

  protected readonly loading = signal(true);
  protected readonly error = signal<string | null>(null);
  protected readonly intakes = signal<CourseIntake[]>([]);

  protected readonly canManage = computed(() => {
    const role = this.authState.currentUser()?.roles[0] ?? 'employee';
    return isIntakeManageRole(role) || this.authState.hasPermission(PERMISSIONS.intakesManage);
  });

  protected readonly statusOptions = [
    { value: '', label: 'Все статусы' },
    { value: 'open', label: 'Открыт' },
    { value: 'closed', label: 'Набор закрыт' },
    { value: 'canceled', label: 'Отменён' },
    { value: 'completed', label: 'Завершён' },
  ];

  protected readonly form = this.fb.group({
    status: [''],
    query: [''],
  });

  ngOnInit(): void {
    this.load();
  }

  protected applyFilters(): void {
    this.load();
  }

  protected clearFilters(): void {
    this.form.setValue({ status: '', query: '' });
    this.load();
  }

  protected statusLabel(status: string): string {
    return courseIntakeStatusLabel(status);
  }

  private load(): void {
    this.loading.set(true);
    this.error.set(null);

    const status = normalizeText(this.form.controls.status.value);
    const query = (this.form.controls.query.value ?? '').trim().toLowerCase();

    this.api.list(status).subscribe({
      next: (items) => {
        const filtered = !query
          ? items
          : items.filter((item) =>
              [item.title, item.description ?? '', item.status]
                .join(' ')
                .toLowerCase()
                .includes(query),
            );

        this.intakes.set(filtered);
        this.loading.set(false);
      },
      error: () => {
        this.error.set('Не удалось загрузить наборы.');
        this.loading.set(false);
      },
    });
  }
}

function normalizeText(value: string | null | undefined): string | undefined {
  const trimmed = value?.trim();
  return trimmed ? trimmed : undefined;
}
