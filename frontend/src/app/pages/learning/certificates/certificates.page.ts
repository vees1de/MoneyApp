import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { RouterLink } from '@angular/router';
import { EnrollmentsProgressFacade } from '@features/enrollments-progress';
import type { Certificate } from '@entities/certificate';

@Component({
  selector: 'app-page-learning-certificates',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './certificates.page.html',
  styleUrl: './certificates.page.scss',
})
export class LearningCertificatesPageComponent {
  private readonly facade = inject(EnrollmentsProgressFacade);
  protected readonly routePath = '/learning/certificates';
  protected readonly entitySample: Certificate[] = [];

  protected loadPage(): void {
    this.facade.load();
  }
}
