import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { RouterLink } from '@angular/router';
import { ExternalCourseRequestsFacade } from '@features/external-course-requests';
import type { Request as RequestEntity } from '@entities/request';

@Component({
  selector: 'app-page-external-requests-detail',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './detail.page.html',
  styleUrl: './detail.page.scss',
})
export class ExternalRequestsDetailPageComponent {
  private readonly facade = inject(ExternalCourseRequestsFacade);
  protected readonly routePath = '/external-requests/detail';
  protected readonly entitySample: RequestEntity[] = [];

  protected loadPage(): void {
    this.facade.load();
  }
}
