import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { RouterLink } from '@angular/router';
import { AuthFacade } from '@features/auth';
import type { User } from '@entities/user';

@Component({
  selector: 'app-page-dashboard-hr',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './hr.page.html',
  styleUrl: './hr.page.scss',
})
export class DashboardHrPageComponent {
  private readonly facade = inject(AuthFacade);
  protected readonly routePath = '/dashboard/hr';
  protected readonly entitySample: User[] = [];

  protected loadPage(): void {
    this.facade.load();
  }
}
