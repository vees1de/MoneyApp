import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { RouterLink } from '@angular/router';
import { CorporateUniversityFacade } from '@features/corporate-university';
import type { Course } from '@entities/course';

@Component({
  selector: 'app-page-university-groups',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './groups.page.html',
  styleUrl: './groups.page.scss',
})
export class UniversityGroupsPageComponent {
  private readonly facade = inject(CorporateUniversityFacade);
  protected readonly routePath = '/university/groups';
  protected readonly entitySample: Course[] = [];

  protected loadPage(): void {
    this.facade.load();
  }
}
