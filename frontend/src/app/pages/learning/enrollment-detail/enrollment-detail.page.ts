import { CommonModule } from '@angular/common';
import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatIconModule } from '@angular/material/icon';
import { catchError, forkJoin, of } from 'rxjs';

import { CertificatesApiService } from '@core/api/certificates-api.service';
import { EnrollmentsApiService } from '@core/api/enrollments-api.service';
import { API_BASE_URL } from '@core/config/api.config';
import type { Certificate } from '@entities/certificate';
import type { Enrollment } from '@entities/enrollment';

@Component({
  selector: 'app-page-learning-enrollment-detail',
  standalone: true,
  imports: [CommonModule, RouterLink, MatButtonModule, MatCardModule, MatIconModule],
  templateUrl: './enrollment-detail.page.html',
  styleUrl: './enrollment-detail.page.scss',
})
export class LearningEnrollmentDetailPageComponent implements OnInit {
  private readonly enrollmentsApi = inject(EnrollmentsApiService);
  private readonly certificatesApi = inject(CertificatesApiService);
  private readonly route = inject(ActivatedRoute);

  protected readonly loading = signal(true);
  protected readonly acting = signal(false);
  protected readonly error = signal<string | null>(null);
  protected readonly enrollment = signal<Enrollment | null>(null);
  protected readonly certificates = signal<Certificate[]>([]);
  protected readonly selectedCertificateFile = signal<File | null>(null);

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
    if (!item || item.status === 'completed' || item.status === 'canceled') {
      return false;
    }

    return !certificate || certificate.status === 'rejected';
  });

  ngOnInit(): void {
    this.load();
  }

  protected shortId(value: string): string {
    if (!value || value.length < 14) {
      return value;
    }

    return `${value.slice(0, 8)}…${value.slice(-4)}`;
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

  protected sourceLabel(source: string): string {
    const labels: Record<string, string> = {
      intake: 'Intake',
      manager: 'Manager',
      hr: 'HR',
      self: 'Self',
      catalog: 'Catalog',
    };

    return labels[source] ?? source;
  }

  protected mandatoryLabel(item: Enrollment): string {
    return item.is_mandatory ? 'Обязательное' : 'По выбору';
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

  protected certificateFileUrl(certificate: Certificate): string | null {
    const storageKey = certificate.file_storage_key?.trim();
    if (!storageKey) {
      return null;
    }

    return `${API_BASE_URL}/uploads/${encodeURI(storageKey)}`;
  }

  protected certificateFlowMessage(): string {
    const item = this.enrollment();
    if (!item) {
      return '';
    }

    const certificate = this.latestCertificate();
    if (!certificate) {
      return item.status === 'completed'
        ? 'Обучение уже завершено.'
        : 'Загрузите сертификат. После апрува HR обучение завершится автоматически.';
    }

    switch (certificate.status) {
      case 'uploaded':
        return 'Сертификат отправлен на проверку HR. После апрува обучение завершится автоматически.';
      case 'verified':
        return 'Сертификат подтверждён HR. Обучение завершено автоматически.';
      case 'rejected':
        return 'HR отклонил сертификат. Загрузите новый файл, чтобы отправить его повторно.';
      default:
        return item.status === 'completed'
          ? 'Обучение завершено автоматически.'
          : 'Загрузите сертификат, чтобы отправить его на проверку.';
    }
  }

  protected onCertificateFileSelected(event: Event): void {
    const input = event.target as HTMLInputElement | null;
    const file = input?.files?.[0] ?? null;
    this.selectedCertificateFile.set(file);
    if (input) {
      input.value = '';
    }
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
    const storageKey = `certificates/${item.id}/${timestamp}-${safeName}`;
    const formData = new FormData();
    formData.append('enrollment_id', item.id);
    formData.append('course_id', item.course_id);
    formData.append('storage_provider', 'local');
    formData.append('storage_key', storageKey);
    formData.append('original_name', file.name);
    formData.append('mime_type', file.type || 'application/octet-stream');
    formData.append('size_bytes', String(Math.max(file.size, 1)));
    formData.append('file', file, file.name);

    this.certificatesApi
      .upload(formData)
      .subscribe({
        next: (certificate) => {
          this.certificates.update((items) => [certificate, ...items]);
          this.selectedCertificateFile.set(null);
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

    forkJoin({
      enrollment: this.enrollmentsApi.getById(id),
      certificates: this.certificatesApi.listMy().pipe(catchError(() => of([]))),
    }).subscribe({
      next: ({ enrollment, certificates }) => {
        this.enrollment.set(enrollment);
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
