import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { RouterLink } from '@angular/router';
import { ExternalCourseRequestsFacade } from '@features/external-course-requests';
import type { Request as RequestEntity } from '@entities/request';

@Component({
  selector: 'app-page-external-requests-list',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './list.page.html',
  styleUrl: './list.page.scss',
})
export class ExternalRequestsListPageComponent {
  private readonly facade = inject(ExternalCourseRequestsFacade);
  protected readonly routePath = '/external-requests/list';
  protected readonly entitySample: RequestEntity[] = [];

  protected loadPage(): void {
    this.facade.load();
  }
}
