import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { RouterLink } from '@angular/router';
import { EnrollmentsProgressFacade } from '@features/enrollments-progress';
import type { Certificate } from '@entities/certificate';

@Component({
  selector: 'app-page-learning-enrollment-detail',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './enrollment-detail.page.html',
  styleUrl: './enrollment-detail.page.scss',
})
export class LearningEnrollmentDetailPageComponent {
  private readonly facade = inject(EnrollmentsProgressFacade);
  protected readonly routePath = '/learning/enrollment-detail';
  protected readonly entitySample: Certificate[] = [];

  protected loadPage(): void {
    this.facade.load();
  }
}
