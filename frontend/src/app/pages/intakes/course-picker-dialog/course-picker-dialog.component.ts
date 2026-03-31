import { CommonModule } from '@angular/common';
import { Component, OnInit, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MAT_DIALOG_DATA, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';

import { CoursesApiService } from '@core/api/courses-api.service';
import type { Course } from '@entities/course';

interface CoursePickerDialogData {
  selectedCourseId?: string | null;
}

@Component({
  selector: 'app-course-picker-dialog',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    MatButtonModule,
    MatDialogModule,
    MatFormFieldModule,
    MatIconModule,
    MatInputModule,
    MatSelectModule,
  ],
  templateUrl: './course-picker-dialog.component.html',
  styleUrl: './course-picker-dialog.component.scss',
})
export class CoursePickerDialogComponent implements OnInit {
  private readonly coursesApi = inject(CoursesApiService);
  private readonly fb = inject(FormBuilder);
  private readonly dialogRef = inject(MatDialogRef<CoursePickerDialogComponent, Course | null>);
  protected readonly data =
    inject<CoursePickerDialogData>(MAT_DIALOG_DATA, { optional: true }) ?? {};

  protected readonly loading = signal(true);
  protected readonly error = signal<string | null>(null);
  protected readonly courses = signal<Course[]>([]);

  protected readonly form = this.fb.group({
    q: [''],
    source_type: [''],
    level: [''],
  });

  protected readonly sourceTypeOptions = [
    { value: '', label: 'Все источники' },
    { value: 'catalog', label: 'Каталог' },
    { value: 'requested', label: 'Запрошен ранее' },
    { value: 'imported', label: 'Импортирован' },
  ];

  protected readonly levelOptions = [
    { value: '', label: 'Любой уровень' },
    { value: 'junior', label: 'Junior' },
    { value: 'middle', label: 'Middle' },
    { value: 'senior', label: 'Senior' },
    { value: 'lead', label: 'Lead' },
  ];

  ngOnInit(): void {
    this.loadCourses();
  }

  protected applyFilters(): void {
    this.loadCourses();
  }

  protected resetFilters(): void {
    this.form.reset({
      q: '',
      source_type: '',
      level: '',
    });
    this.loadCourses();
  }

  protected selectCourse(course: Course): void {
    this.dialogRef.close(course);
  }

  protected close(): void {
    this.dialogRef.close(null);
  }

  protected isSelected(course: Course): boolean {
    return course.id === this.data.selectedCourseId;
  }

  private loadCourses(): void {
    this.loading.set(true);
    this.error.set(null);

    const values = this.form.getRawValue();
    this.coursesApi
      .list({
        limit: 30,
        sort: values.q?.trim() ? 'relevance' : 'newest',
        status: 'published',
        q: values.q?.trim() || undefined,
        source_type: values.source_type || undefined,
        level: values.level || undefined,
      })
      .subscribe({
        next: (courses) => {
          this.courses.set(courses ?? []);
          this.loading.set(false);
        },
        error: () => {
          this.error.set('Не удалось загрузить каталог курсов.');
          this.loading.set(false);
        },
      });
  }
}
