import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { RouterLink } from '@angular/router';
import { LearningCatalogFacade } from '@features/learning-catalog';
import type { Course } from '@entities/course';

@Component({
  selector: 'app-page-catalog-list',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './list.page.html',
  styleUrl: './list.page.scss',
})
export class CatalogListPageComponent {
  private readonly facade = inject(LearningCatalogFacade);
  protected readonly routePath = '/catalog/list';
  protected readonly entitySample: Course[] = [];

  protected loadPage(): void {
    this.facade.load();
  }
}
