import { CommonModule } from '@angular/common';
import { Component, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';

import { CoursesApiService } from '@core/api/courses-api.service';
import type { Course } from '@entities/course';

@Component({
  selector: 'app-page-catalog-list',
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
export class CatalogListPageComponent {
  private readonly api = inject(CoursesApiService);
  private readonly fb = inject(FormBuilder);

  protected readonly loading = signal(true);
  protected readonly error = signal<string | null>(null);
  protected readonly courses = signal<Course[]>([]);

  protected readonly form = this.fb.group({
    status: [''],
    source_type: [''],
    level: [''],
    limit: [50],
    offset: [0],
  });

  protected readonly statusOptions = [
    { value: '', label: 'Все статусы' },
    { value: 'draft', label: 'Draft' },
    { value: 'published', label: 'Published' },
    { value: 'archived', label: 'Archived' },
  ];

  protected readonly sourceTypeOptions = [
    { value: '', label: 'Все источники' },
    { value: 'internal', label: 'Internal' },
    { value: 'external', label: 'External' },
  ];

  constructor() {
    this.loadCourses();
  }

  protected submitFilters(): void {
    this.loadCourses();
  }

  protected resetFilters(): void {
    this.form.setValue({
      status: '',
      source_type: '',
      level: '',
      limit: 50,
      offset: 0,
    });
    this.loadCourses();
  }

  protected trackByCourse(_: number, course: Course): string {
    return course.id;
  }

  private loadCourses(): void {
    this.loading.set(true);
    this.error.set(null);

    const v = this.form.getRawValue();
    const query: Record<string, string | number> = {
      limit: v.limit ?? 50,
      offset: v.offset ?? 0,
    };

    if (v.status) query['status'] = v.status;
    if (v.source_type) query['source_type'] = v.source_type;
    if (v.level) query['level'] = v.level;

    this.api.list(query).subscribe({
      next: (courses) => {
        this.courses.set(courses ?? []);
        this.loading.set(false);
      },
      error: () => {
        this.error.set('Не удалось загрузить курсы');
        this.loading.set(false);
      },
    });
  }
}
