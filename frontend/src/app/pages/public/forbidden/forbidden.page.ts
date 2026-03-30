import { CommonModule, Location } from '@angular/common';
import { Component, inject } from '@angular/core';

@Component({
  selector: 'app-page-public-forbidden',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './forbidden.page.html',
  styleUrl: './forbidden.page.scss',
})
export class PublicForbiddenPageComponent {
  private readonly location = inject(Location);

  protected goBack(): void {
    this.location.back();
  }
}
