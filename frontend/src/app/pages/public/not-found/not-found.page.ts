import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { RouterLink } from '@angular/router';
import { AuthFacade } from '@features/auth';
import type { User } from '@entities/user';

@Component({
  selector: 'app-page-public-not-found',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './not-found.page.html',
  styleUrl: './not-found.page.scss',
})
export class PublicNotFoundPageComponent {
  private readonly facade = inject(AuthFacade);
  protected readonly routePath = '/not-found';
  protected readonly entitySample: User[] = [];

  protected loadPage(): void {
    this.facade.load();
  }
}
