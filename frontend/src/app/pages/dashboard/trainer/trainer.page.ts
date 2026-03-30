import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { RouterLink } from '@angular/router';
import { AuthFacade } from '@features/auth';
import type { User } from '@entities/user';

@Component({
  selector: 'app-page-dashboard-trainer',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './trainer.page.html',
  styleUrl: './trainer.page.scss',
})
export class DashboardTrainerPageComponent {
  private readonly facade = inject(AuthFacade);
  protected readonly routePath = '/dashboard/trainer';
  protected readonly entitySample: User[] = [];

  protected loadPage(): void {
    this.facade.load();
  }
}
