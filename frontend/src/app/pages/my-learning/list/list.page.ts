import { CommonModule } from '@angular/common';
import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { catchError, forkJoin, of } from 'rxjs';

import { CertificatesApiService } from '@core/api/certificates-api.service';
import { CoursesApiService } from '@core/api/courses-api.service';
import { EnrollmentsApiService } from '@core/api/enrollments-api.service';
import type { Certificate } from '@entities/certificate';
import type { Course } from '@entities/course';
import type { Enrollment } from '@entities/enrollment';

type Tone = 'info' | 'success' | 'warning' | 'muted' | 'danger';

interface LearningCardView {
  id: string;
  title: string;
  description: string;
  status: string;
  statusLabel: string;
  statusTone: Tone;
  deadlineText: string;
  activityText: string;
  enrolledText: string;
  certificateLabel: string;
  certificateTone: Tone;
  certificateReady: boolean;
  overdue: boolean;
  needsCertificate: boolean;
  deadlineSort: number;
}

interface SummaryCard {
  label: string;
  value: string;
  hint: string;
  tone: Tone;
}

@Component({
  selector: 'app-page-my-learning-list',
  standalone: true,
  imports: [CommonModule, RouterLink, MatButtonModule, MatIconModule],
  templateUrl: './list.page.html',
  styleUrl: './list.page.scss',
})
export class MyLearningListPageComponent implements OnInit {
  private readonly coursesApi = inject(CoursesApiService);
  private readonly certificatesApi = inject(CertificatesApiService);
  private readonly api = inject(EnrollmentsApiService);

  protected readonly loading = signal(true);
  protected readonly error = signal<string | null>(null);
  protected readonly items = signal<LearningCardView[]>([]);
  protected readonly summaryCards = computed<SummaryCard[]>(() => {
    const items = this.items();

    return [
      {
        label: 'Активные',
        value: String(
          items.filter((item) => item.status === 'in_progress' || item.status === 'enrolled').length,
        ),
        hint: 'Курсы в работе',
        tone: 'info',
      },
      {
        label: 'С сертификатом',
        value: String(items.filter((item) => item.certificateReady).length),
        hint: 'Можно завершить',
        tone: 'success',
      },
      {
        label: 'Нужен файл',
        value: String(items.filter((item) => item.needsCertificate).length),
        hint: 'Сначала загрузите сертификат',
        tone: 'warning',
      },
      {
        label: 'Просроченные',
        value: String(items.filter((item) => item.overdue).length),
        hint: 'Требуют внимания',
        tone: 'danger',
      },
    ];
  });
  protected readonly focusItem = computed<LearningCardView | null>(() => {
    const items = this.items();
    return items.find((item) => item.needsCertificate) ?? items.find((item) => item.overdue) ?? items[0] ?? null;
  });

  ngOnInit(): void {
    forkJoin({
      enrollments: this.api.listMy(),
      courses: this.coursesApi.list({ limit: 300, offset: 0 }).pipe(catchError(() => of([] as Course[]))),
      certificates: this.certificatesApi.listMy().pipe(catchError(() => of([] as Certificate[]))),
    }).subscribe({
      next: ({ enrollments, courses, certificates }) => {
        this.items.set(this.mapItems(enrollments ?? [], courses ?? [], certificates ?? []));
        this.loading.set(false);
      },
      error: () => {
        this.error.set('Не удалось загрузить обучение');
        this.loading.set(false);
      },
    });
  }

  protected trackById(_: number, item: LearningCardView): string {
    return item.id;
  }

  private mapItems(
    enrollments: Enrollment[],
    courses: Course[],
    certificates: Certificate[],
  ): LearningCardView[] {
    const courseMap = new Map(courses.map((course) => [course.id, course]));
    const latestCertificates = this.latestCertificatesByEnrollment(certificates);

    return enrollments
      .map((item) => {
        const course = courseMap.get(item.course_id);
        const certificate = latestCertificates.get(item.id) ?? null;
        const deadline = item.deadline_at ? new Date(item.deadline_at) : null;
        const activity = item.last_activity_at ? new Date(item.last_activity_at) : null;
        const enrolled = new Date(item.enrolled_at);
        const title = course?.title?.trim() || `Курс ${item.course_id.slice(0, 8)}`;
        const description =
          course?.short_description?.trim() || this.fallbackDescription(item.status, certificate?.status);
        const statusTone = this.statusTone(item.status);
        const certificateTone = this.certificateTone(certificate?.status);
        const certificateLabel = certificate
          ? this.certificateLabel(certificate.status)
          : 'Сертификат не загружен';
        const certificateReady = !!certificate && certificate.status !== 'rejected';
        const needsCertificate = item.status === 'in_progress' && !certificateReady;
        const overdue =
          item.status !== 'completed' &&
          !!deadline &&
          !Number.isNaN(deadline.getTime()) &&
          deadline.getTime() < Date.now();

        return {
          id: item.id,
          title,
          description,
          status: item.status,
          statusLabel: this.statusLabel(item.status),
          statusTone,
          deadlineText: this.formatDate(deadline),
          activityText: activity ? this.formatDateTime(activity) : 'Активность не зафиксирована',
          enrolledText: this.formatDate(enrolled),
          certificateLabel,
          certificateTone,
          certificateReady,
          overdue,
          needsCertificate,
          deadlineSort:
            deadline && !Number.isNaN(deadline.getTime()) ? deadline.getTime() : Number.POSITIVE_INFINITY,
        };
      })
      .sort((left, right) => {
        if (left.overdue !== right.overdue) {
          return left.overdue ? -1 : 1;
        }

        const leftRank = this.statusRank(left.status);
        const rightRank = this.statusRank(right.status);
        if (leftRank !== rightRank) {
          return leftRank - rightRank;
        }

        return left.deadlineSort - right.deadlineSort;
      });
  }

  private latestCertificatesByEnrollment(certificates: Certificate[]): Map<string, Certificate> {
    const map = new Map<string, Certificate>();

    for (const certificate of certificates) {
      if (!certificate.enrollment_id) {
        continue;
      }

      const current = map.get(certificate.enrollment_id);
      if (!current) {
        map.set(certificate.enrollment_id, certificate);
        continue;
      }

      const currentTime = new Date(current.uploaded_at).getTime();
      const nextTime = new Date(certificate.uploaded_at).getTime();
      if (Number.isNaN(currentTime) || nextTime > currentTime) {
        map.set(certificate.enrollment_id, certificate);
      }
    }

    return map;
  }

  private fallbackDescription(status: string, certificateStatus?: string | null): string {
    if (status === 'in_progress') {
      if (certificateStatus === 'rejected') {
        return 'Предыдущий сертификат отклонён. Загрузите новый файл.';
      }

      if (certificateStatus) {
        return 'Сертификат уже прикреплён. Остался последний шаг завершения.';
      }

      return 'Сначала загрузите сертификат, затем завершите курс.';
    }

    if (status === 'completed') {
      return 'Курс завершён и доступен для проверки HR.';
    }

    if (status === 'enrolled') {
      return 'Курс назначен, но ещё не начат.';
    }

    return 'Курс находится в архивном или отменённом состоянии.';
  }

  private statusLabel(status: string): string {
    const labels: Record<string, string> = {
      enrolled: 'Назначено',
      in_progress: 'В процессе',
      completed: 'Завершено',
      canceled: 'Отменено',
    };

    return labels[status] ?? status;
  }

  private certificateLabel(status?: string | null): string {
    const labels: Record<string, string> = {
      uploaded: 'Загружен и ждёт проверки',
      verified: 'Проверен HR',
      rejected: 'Отклонён HR',
    };

    if (!status) {
      return 'Сертификат не загружен';
    }

    return labels[status] ?? status;
  }

  private statusTone(status: string): Tone {
    switch (status) {
      case 'completed':
        return 'success';
      case 'in_progress':
        return 'info';
      case 'enrolled':
        return 'warning';
      case 'canceled':
        return 'muted';
      default:
        return 'muted';
    }
  }

  private statusRank(status: string): number {
    switch (status) {
      case 'in_progress':
        return 0;
      case 'enrolled':
        return 1;
      case 'completed':
        return 2;
      case 'canceled':
        return 3;
      default:
        return 4;
    }
  }

  private certificateTone(status?: string | null): Tone {
    switch (status) {
      case 'verified':
        return 'success';
      case 'uploaded':
        return 'warning';
      case 'rejected':
        return 'danger';
      default:
        return 'muted';
    }
  }

  private formatDate(value: Date | string | null | undefined): string {
    if (!value) {
      return '—';
    }

    const parsed = value instanceof Date ? value : new Date(value);
    if (Number.isNaN(parsed.getTime())) {
      return '—';
    }

    return parsed.toLocaleDateString('ru-RU', {
      day: '2-digit',
      month: 'short',
      year: 'numeric',
    });
  }

  private formatDateTime(value: Date | string | null | undefined): string {
    if (!value) {
      return '—';
    }

    const parsed = value instanceof Date ? value : new Date(value);
    if (Number.isNaN(parsed.getTime())) {
      return '—';
    }

    return parsed.toLocaleString('ru-RU', {
      day: '2-digit',
      month: 'short',
      hour: '2-digit',
      minute: '2-digit',
    });
  }
}
