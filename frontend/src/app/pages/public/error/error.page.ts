import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { RouterLink } from '@angular/router';
import { AuthFacade } from '@features/auth';
import type { User } from '@entities/user';

@Component({
  selector: 'app-page-public-error',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './error.page.html',
  styleUrl: './error.page.scss',
})
export class PublicErrorPageComponent {
  private readonly facade = inject(AuthFacade);
  protected readonly routePath = '/error';
  protected readonly entitySample: User[] = [];

  protected loadPage(): void {
    this.facade.load();
  }
}
