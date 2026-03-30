import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { RouterLink } from '@angular/router';
import { CorporateUniversityFacade } from '@features/corporate-university';
import type { Course } from '@entities/course';

@Component({
  selector: 'app-page-university-programs',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './programs.page.html',
  styleUrl: './programs.page.scss',
})
export class UniversityProgramsPageComponent {
  private readonly facade = inject(CorporateUniversityFacade);
  protected readonly routePath = '/university/programs';
  protected readonly entitySample: Course[] = [];

  protected loadPage(): void {
    this.facade.load();
  }
}
