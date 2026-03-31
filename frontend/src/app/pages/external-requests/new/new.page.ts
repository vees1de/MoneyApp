import { CommonModule } from '@angular/common';
import { Component, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';

import { ExternalRequestsApiService } from '@core/api/external-requests-api.service';

@Component({
  selector: 'app-page-external-requests-new',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    RouterLink,
    MatButtonModule,
    MatCardModule,
    MatFormFieldModule,
    MatIconModule,
    MatInputModule,
  ],
  templateUrl: './new.page.html',
  styleUrl: './new.page.scss',
})
export class ExternalRequestsNewPageComponent {
  private readonly api = inject(ExternalRequestsApiService);
  private readonly fb = inject(FormBuilder);
  private readonly router = inject(Router);

  protected readonly submitting = signal(false);
  protected readonly error = signal<string | null>(null);

  protected readonly form = this.fb.group({
    title: ['', [Validators.required]],
    provider_name: ['', [Validators.required]],
    course_url: ['', [Validators.required]],
    program_description: ['', [Validators.required]],
    planned_start_date: [''],
    planned_end_date: [''],
    duration_hours: [''],
    cost_amount: ['', [Validators.required]],
    currency: ['RUB', [Validators.required]],
    business_goal: [''],
    employee_comment: [''],
  });

  protected async submit(): Promise<void> {
    if (this.form.invalid || this.submitting()) {
      this.form.markAllAsTouched();
      return;
    }

    this.submitting.set(true);
    this.error.set(null);

    const payload = this.form.getRawValue();

    this.api.create(payload).subscribe({
      next: async (created) => {
        await this.router.navigateByUrl(`/external-requests/${created.id}`);
        this.submitting.set(false);
      },
      error: () => {
        this.error.set('Не удалось создать заявку');
        this.submitting.set(false);
      },
    });
  }
}
