import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { RouterLink } from '@angular/router';
import { AuthFacade } from '@features/auth';
import type { User } from '@entities/user';

@Component({
  selector: 'app-page-public-login',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './login.page.html',
  styleUrl: './login.page.scss',
})
export class PublicLoginPageComponent {
  private readonly facade = inject(AuthFacade);
  protected readonly routePath = '/login';
  protected readonly entitySample: User[] = [];

  protected loadPage(): void {
    this.facade.load();
  }
}
