import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { RouterLink } from '@angular/router';
import { AdminFacade } from '@features/admin';
import type { User } from '@entities/user';

@Component({
  selector: 'app-page-admin-roles',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './roles.page.html',
  styleUrl: './roles.page.scss',
})
export class AdminRolesPageComponent {
  private readonly facade = inject(AdminFacade);
  protected readonly routePath = '/admin/roles';
  protected readonly entitySample: User[] = [];

  protected loadPage(): void {
    this.facade.load();
  }
}
