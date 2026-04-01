import { CommonModule } from '@angular/common';
import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { catchError, forkJoin, of, switchMap } from 'rxjs';

import { CertificatesApiService } from '@core/api/certificates-api.service';
import { CoursesApiService } from '@core/api/courses-api.service';
import { EnrollmentsApiService } from '@core/api/enrollments-api.service';
import { normalizeText } from '@core/domain/course-intake-form.util';
import type { Certificate } from '@entities/certificate';
import type { Course } from '@entities/course';
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
    MatIconModule,
    MatInputModule,
  ],
  templateUrl: './enrollment-detail.page.html',
  styleUrl: './enrollment-detail.page.scss',
})
export class LearningEnrollmentDetailPageComponent implements OnInit {
  private readonly enrollmentsApi = inject(EnrollmentsApiService);
  private readonly certificatesApi = inject(CertificatesApiService);
  private readonly coursesApi = inject(CoursesApiService);
  private readonly route = inject(ActivatedRoute);
  private readonly fb = inject(FormBuilder);

  protected readonly loading = signal(true);
  protected readonly acting = signal(false);
  protected readonly error = signal<string | null>(null);
  protected readonly enrollment = signal<Enrollment | null>(null);
  protected readonly course = signal<Course | null>(null);
  protected readonly certificates = signal<Certificate[]>([]);
  protected readonly selectedCertificateFile = signal<File | null>(null);

  protected readonly progressForm = this.fb.group({
    course_module_id: ['', [Validators.required]],
    progress_percent: [50, [Validators.required]],
  });

  protected readonly completionForm = this.fb.group({
    notes: [''],
  });

  protected readonly certificateForm = this.fb.group({
    certificate_no: [''],
    issued_by: [''],
    notes: [''],
  });

  protected readonly latestCertificate = computed<Certificate | null>(() => {
    const item = this.enrollment();
    if (!item) {
      return null;
    }

    const certificates = this.certificates()
      .filter((certificate) => certificate.enrollment_id === item.id)
      .sort((left, right) => {
        const leftTime = new Date(left.uploaded_at).getTime();
        const rightTime = new Date(right.uploaded_at).getTime();
        return rightTime - leftTime;
      });

    return certificates[0] ?? null;
  });

  protected readonly canUploadCertificate = computed(() => {
    const item = this.enrollment();
    const certificate = this.latestCertificate();
    if (!item) {
      return false;
    }

    if (item.status === 'completed') {
      return !certificate || certificate.status === 'rejected';
    }

    return item.status === 'in_progress';
  });

  protected readonly canWorkOnCourse = computed(() => this.enrollment()?.status === 'in_progress');
  protected readonly certificateReady = computed(() => {
    const certificate = this.latestCertificate();
    return !!certificate && certificate.status !== 'rejected';
  });
  protected readonly canCompleteCourse = computed(
    () => this.canWorkOnCourse() && this.certificateReady(),
  );
  protected readonly courseTitle = computed(() => {
    const course = this.course();
    return course?.title?.trim() || this.enrollment()?.course_id || 'Карточка обучения';
  });
  protected readonly courseDescription = computed(() => {
    const course = this.course();
    return (
      course?.short_description?.trim() ||
      'Сначала загрузите сертификат, затем завершите курс. После этого HR проверит документ.'
    );
  });

  ngOnInit(): void {
    this.load();
  }

  protected enrollmentStatusLabel(status: string): string {
    const labels: Record<string, string> = {
      enrolled: 'Назначено',
      in_progress: 'В процессе',
      completed: 'Завершено',
      canceled: 'Отменено',
    };

    return labels[status] ?? status;
  }

  protected certificateStatusLabel(status?: string | null): string {
    const labels: Record<string, string> = {
      uploaded: 'На проверке HR',
      verified: 'Проверен HR',
      rejected: 'Отклонён HR',
    };

    if (!status) {
      return 'Не загружен';
    }

    return labels[status] ?? status;
  }

  protected enrollmentStatusTone(status: string): 'info' | 'success' | 'warning' | 'muted' {
    switch (status) {
      case 'completed':
        return 'success';
      case 'in_progress':
        return 'info';
      case 'enrolled':
        return 'warning';
      default:
        return 'muted';
    }
  }

  protected selectedCertificateFileName(): string {
    return this.selectedCertificateFile()?.name ?? 'Файл не выбран';
  }

  protected updateProgress(): void {
    const item = this.enrollment();
    if (!item || this.acting() || this.progressForm.invalid || item.status !== 'in_progress') {
      return;
    }

    this.acting.set(true);
    this.error.set(null);
    const value = this.progressForm.getRawValue();

    this.enrollmentsApi
      .updateProgress(item.id, {
        course_module_id: value.course_module_id,
        status: 'in_progress',
        progress_percent: String(value.progress_percent ?? 0),
      })
      .subscribe({
        next: (updated) => {
          this.enrollment.set(updated);
          this.acting.set(false);
        },
        error: () => {
          this.error.set('Не удалось обновить прогресс.');
          this.acting.set(false);
        },
      });
  }

  protected complete(): void {
    const item = this.enrollment();
    if (!item || this.acting()) {
      return;
    }

    if (!this.canCompleteCourse()) {
      this.error.set('Сначала загрузите сертификат.');
      return;
    }

    this.acting.set(true);
    this.error.set(null);

    this.enrollmentsApi
      .complete(item.id, {
        completion_type: 'manual',
        notes:
          normalizeText(this.completionForm.controls.notes.value) ??
          'Завершено сотрудником вручную',
      })
      .subscribe({
        next: (updated) => {
          this.enrollment.set(updated);
          this.acting.set(false);
        },
        error: () => {
          this.error.set('Не удалось завершить обучение.');
          this.acting.set(false);
        },
      });
  }

  protected onCertificateFileSelected(event: Event): void {
    const input = event.target as HTMLInputElement | null;
    const file = input?.files?.[0] ?? null;
    this.selectedCertificateFile.set(file);
  }

  protected uploadCertificate(): void {
    const item = this.enrollment();
    const file = this.selectedCertificateFile();
    if (!item || !file || this.acting() || !this.canUploadCertificate()) {
      return;
    }

    this.acting.set(true);
    this.error.set(null);

    const safeName = file.name.replace(/[^a-zA-Z0-9._-]+/g, '-');
    const timestamp = Date.now();

    this.certificatesApi
      .upload({
        enrollment_id: item.id,
        course_id: item.course_id,
        certificate_no: normalizeText(this.certificateForm.controls.certificate_no.value),
        issued_by: normalizeText(this.certificateForm.controls.issued_by.value),
        notes: normalizeText(this.certificateForm.controls.notes.value),
        storage_provider: 'local',
        storage_key: `certificates/${item.id}/${timestamp}-${safeName}`,
        original_name: file.name,
        mime_type: file.type || 'application/octet-stream',
        size_bytes: Math.max(file.size, 1),
      })
      .subscribe({
        next: (certificate) => {
          this.certificates.update((items) => [certificate, ...items]);
          this.selectedCertificateFile.set(null);
          this.certificateForm.reset({
            certificate_no: '',
            issued_by: '',
            notes: '',
          });
          this.acting.set(false);
        },
        error: () => {
          this.error.set('Не удалось загрузить сертификат.');
          this.acting.set(false);
        },
      });
  }

  private load(): void {
    const id = this.route.snapshot.paramMap.get('enrollmentId');
    if (!id) {
      this.error.set('Не найден enrollmentId.');
      this.loading.set(false);
      return;
    }

    this.loading.set(true);
    this.error.set(null);

    this.enrollmentsApi
      .getById(id)
      .pipe(
        switchMap((enrollment) =>
          forkJoin({
            enrollment: of(enrollment),
            course: this.coursesApi.getById(enrollment.course_id).pipe(catchError(() => of(null))),
            certificates: this.certificatesApi.listMy().pipe(catchError(() => of([]))),
          }),
        ),
      )
      .subscribe({
        next: ({ enrollment, course, certificates }) => {
          this.enrollment.set(enrollment);
          this.course.set(course);
          this.certificates.set(certificates ?? []);
          this.loading.set(false);
        },
        error: () => {
          this.error.set('Не удалось загрузить карточку обучения.');
          this.loading.set(false);
        },
      });
  }
}
