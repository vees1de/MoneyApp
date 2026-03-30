import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { RouterLink } from '@angular/router';
import { CorporateUniversityFacade } from '@features/corporate-university';
import type { Course } from '@entities/course';

@Component({
  selector: 'app-page-university-directions',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './directions.page.html',
  styleUrl: './directions.page.scss',
})
export class UniversityDirectionsPageComponent {
  private readonly facade = inject(CorporateUniversityFacade);
  protected readonly routePath = '/university/directions';
  protected readonly entitySample: Course[] = [];

  protected loadPage(): void {
    this.facade.load();
  }
}
