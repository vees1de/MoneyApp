import { CommonModule } from '@angular/common';
import { Component, OnInit, inject, signal } from '@angular/core';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatIconModule } from '@angular/material/icon';

import { CoursesApiService } from '@core/api/courses-api.service';
import type { Course } from '@entities/course';

@Component({
  selector: 'app-page-catalog-detail',
  standalone: true,
  imports: [CommonModule, RouterLink, MatButtonModule, MatCardModule, MatIconModule],
  templateUrl: './detail.page.html',
  styleUrl: './detail.page.scss',
})
export class CatalogDetailPageComponent implements OnInit {
  private readonly coursesApi = inject(CoursesApiService);
  private readonly route = inject(ActivatedRoute);

  protected readonly loading = signal(true);
  protected readonly error = signal<string | null>(null);
  protected readonly course = signal<Course | null>(null);

  ngOnInit(): void {
    const id = this.route.snapshot.paramMap.get('courseId');
    if (!id) {
      this.error.set('Не найден идентификатор курса.');
      this.loading.set(false);
      return;
    }

    this.coursesApi.getById(id).subscribe({
      next: (course) => {
        this.course.set(course);
        this.loading.set(false);
      },
      error: () => {
        this.error.set('Не удалось загрузить курс.');
        this.loading.set(false);
      },
    });
  }
}
