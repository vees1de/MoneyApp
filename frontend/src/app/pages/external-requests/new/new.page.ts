import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { RouterLink } from '@angular/router';
import { ExternalCourseRequestsFacade } from '@features/external-course-requests';
import type { Request as RequestEntity } from '@entities/request';

@Component({
  selector: 'app-page-external-requests-new',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './new.page.html',
  styleUrl: './new.page.scss',
})
export class ExternalRequestsNewPageComponent {
  private readonly facade = inject(ExternalCourseRequestsFacade);
  protected readonly routePath = '/external-requests/new';
  protected readonly entitySample: RequestEntity[] = [];

  protected loadPage(): void {
    this.facade.load();
  }
}
