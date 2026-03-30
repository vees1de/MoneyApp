import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { RouterLink } from '@angular/router';
import { EnrollmentsProgressFacade } from '@features/enrollments-progress';
import type { Enrollment } from '@entities/enrollment';

@Component({
  selector: 'app-page-my-learning-list',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './list.page.html',
  styleUrl: './list.page.scss',
})
export class MyLearningListPageComponent {
  private readonly facade = inject(EnrollmentsProgressFacade);
  protected readonly routePath = '/my-learning/list';
  protected readonly entitySample: Enrollment[] = [];

  protected loadPage(): void {
    this.facade.load();
  }
}
