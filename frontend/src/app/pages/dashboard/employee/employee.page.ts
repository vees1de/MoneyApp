import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { RouterLink } from '@angular/router';
import { AuthFacade } from '@features/auth';
import type { User } from '@entities/user';

@Component({
  selector: 'app-page-dashboard-employee',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './employee.page.html',
  styleUrl: './employee.page.scss',
})
export class DashboardEmployeePageComponent {
  private readonly facade = inject(AuthFacade);
  protected readonly routePath = '/dashboard/employee';
  protected readonly entitySample: User[] = [];

  protected loadPage(): void {
    this.facade.load();
  }
}
