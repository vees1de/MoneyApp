import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';
import { MatCardModule } from '@angular/material/card';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatIconModule } from '@angular/material/icon';
import { MatButtonModule } from '@angular/material/button';
import { MatChipsModule } from '@angular/material/chips';
import { MatExpansionModule } from '@angular/material/expansion';

import { AIApiService, type AIRecommendResponse } from '@core/api/ai-api.service';

@Component({
  selector: 'app-page-ai',
  standalone: true,
  imports: [
    CommonModule,
    RouterLink,
    MatCardModule,
    MatProgressSpinnerModule,
    MatIconModule,
    MatButtonModule,
    MatChipsModule,
    MatExpansionModule,
  ],
  templateUrl: './ai.page.html',
  styleUrl: './ai.page.scss',
})
export class AIPageComponent implements OnInit {
  private readonly api = inject(AIApiService);

  protected readonly loading = signal(true);
  protected readonly error = signal<string | null>(null);
  protected readonly data = signal<AIRecommendResponse | null>(null);

  ngOnInit(): void {
    this.loadRecommendations();
  }

  protected loadRecommendations(): void {
    this.loading.set(true);
    this.error.set(null);

    this.api.getRecommendations().subscribe({
      next: (result) => {
        this.data.set(result);
        this.loading.set(false);
      },
      error: (err) => {
        const message = err?.error?.message || err?.message || 'Не удалось получить рекомендации';
        this.error.set(message);
        this.loading.set(false);
      },
    });
  }
}
