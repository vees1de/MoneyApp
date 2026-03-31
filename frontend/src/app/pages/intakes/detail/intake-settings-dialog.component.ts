import { CommonModule } from '@angular/common';
import { Component, computed, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatButtonToggleModule } from '@angular/material/button-toggle';
import { MAT_DIALOG_DATA, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';

import type { CourseIntake } from '@core/api/contracts';
import { toDatetimeLocalValue } from '@core/domain/date-input.util';
import { resolveEndDateFromWeeks, toPositiveNumber } from '@core/domain/course-intake-form.util';

interface IntakeSettingsDialogData {
  intake: CourseIntake;
}

type IntakeSettingsFormValue = {
  title: string;
  description: string;
  max_participants: string;
  price: string;
  price_currency: string;
  start_date: string;
  schedule_mode: string;
  duration_weeks: string;
  end_date: string;
  application_deadline: string;
  status: string;
};

export type IntakeSettingsDialogResult =
  | { action: 'save'; values: IntakeSettingsFormValue }
  | { action: 'close' };

@Component({
  selector: 'app-intake-settings-dialog',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    MatButtonModule,
    MatButtonToggleModule,
    MatDialogModule,
    MatFormFieldModule,
    MatIconModule,
    MatInputModule,
    MatSelectModule,
  ],
  templateUrl: './intake-settings-dialog.component.html',
  styleUrl: './intake-settings-dialog.component.scss',
})
export class IntakeSettingsDialogComponent {
  protected readonly data = inject<IntakeSettingsDialogData>(MAT_DIALOG_DATA);
  private readonly dialogRef =
    inject<MatDialogRef<IntakeSettingsDialogComponent, IntakeSettingsDialogResult | null>>(
      MatDialogRef,
    );
  private readonly fb = inject(FormBuilder);

  protected readonly error = signal<string | null>(null);
  protected readonly form = this.fb.group({
    title: [this.data.intake.title ?? '', [Validators.required]],
    description: [this.data.intake.description ?? ''],
    max_participants: [this.data.intake.max_participants?.toString() ?? ''],
    price: [this.data.intake.price ?? ''],
    price_currency: [this.data.intake.price_currency ?? 'RUB'],
    start_date: [this.data.intake.start_date ?? '', [Validators.required]],
    schedule_mode: [this.data.intake.duration_weeks ? 'weeks' : 'end', [Validators.required]],
    duration_weeks: [this.data.intake.duration_weeks?.toString() ?? ''],
    end_date: [this.data.intake.end_date ?? ''],
    application_deadline: [toDatetimeLocalValue(this.data.intake.application_deadline)],
    status: [this.data.intake.status ?? 'open', [Validators.required]],
  });

  protected readonly statusOptions = [
    { value: 'open', label: 'Открыт' },
    { value: 'closed', label: 'Набор закрыт' },
    { value: 'canceled', label: 'Отменён' },
    { value: 'completed', label: 'Завершён' },
  ];

  protected readonly isWeeksMode = computed(
    () => this.form.controls.schedule_mode.value === 'weeks',
  );
  protected readonly calculatedEndDate = computed(() =>
    resolveEndDateFromWeeks(
      this.form.controls.start_date.value,
      toPositiveNumber(this.form.controls.duration_weeks.value) ?? null,
    ),
  );

  protected save(): void {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    const values = this.form.getRawValue() as IntakeSettingsFormValue;
    if (!values.title.trim()) {
      this.error.set('Название набора обязательно.');
      return;
    }
    if (this.isWeeksMode() && !toPositiveNumber(values.duration_weeks)) {
      this.error.set('Укажите количество недель.');
      return;
    }
    if (!this.isWeeksMode() && !values.end_date.trim()) {
      this.error.set('Укажите дату окончания.');
      return;
    }

    this.error.set(null);
    this.dialogRef.close({ action: 'save', values });
  }

  protected requestClose(): void {
    this.dialogRef.close({ action: 'close' });
  }

  protected cancel(): void {
    this.dialogRef.close(null);
  }
}
