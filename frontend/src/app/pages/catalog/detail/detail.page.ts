import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { RouterLink } from '@angular/router';
import { LearningCatalogFacade } from '@features/learning-catalog';
import type { Course } from '@entities/course';

@Component({
  selector: 'app-page-catalog-detail',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './detail.page.html',
  styleUrl: './detail.page.scss',
})
export class CatalogDetailPageComponent {
  private readonly facade = inject(LearningCatalogFacade);
  protected readonly routePath = '/catalog/detail';
  protected readonly entitySample: Course[] = [];

  protected loadPage(): void {
    this.facade.load();
  }
}
