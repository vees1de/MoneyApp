import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { RouterLink } from '@angular/router';
import { CorporateUniversityFacade } from '@features/corporate-university';
import type { Course } from '@entities/course';

@Component({
  selector: 'app-page-university-feedback',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './feedback.page.html',
  styleUrl: './feedback.page.scss',
})
export class UniversityFeedbackPageComponent {
  private readonly facade = inject(CorporateUniversityFacade);
  protected readonly routePath = '/university/feedback';
  protected readonly entitySample: Course[] = [];

  protected loadPage(): void {
    this.facade.load();
  }
}
