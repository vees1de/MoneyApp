import { CommonModule } from '@angular/common';
import { Component, OnInit, inject, signal, computed } from '@angular/core';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatCheckboxModule } from '@angular/material/checkbox';
import { MatChipsModule } from '@angular/material/chips';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { CdkDragDrop, DragDropModule, moveItemInArray } from '@angular/cdk/drag-drop';
import { ReportsAnalyticsFacade } from '@features/reports-analytics';
import type { ColumnMeta, SourceMeta, SmartExportRequest } from '@entities/smart-export';

interface ColumnSelection {
  column: ColumnMeta;
  selected: boolean;
}

@Component({
  selector: 'app-page-reports-overview',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    MatButtonModule,
    MatCardModule,
    MatCheckboxModule,
    MatChipsModule,
    MatFormFieldModule,
    MatIconModule,
    MatInputModule,
    MatSelectModule,
    MatProgressSpinnerModule,
    DragDropModule,
  ],
  templateUrl: './overview.page.html',
  styleUrl: './overview.page.scss',
})
export class ReportsOverviewPageComponent implements OnInit {
  protected readonly facade = inject(ReportsAnalyticsFacade);
  private readonly fb = inject(FormBuilder);

  protected readonly columnSelections = signal<ColumnSelection[]>([]);
  protected readonly activeSource = signal<SourceMeta | null>(null);

  protected readonly selectedColumns = computed(() =>
    this.columnSelections().filter((cs) => cs.selected),
  );

  protected readonly selectedCount = computed(() => this.selectedColumns().length);

  protected readonly totalCount = computed(() => this.columnSelections().length);

  protected readonly canExport = computed(
    () => this.activeSource() !== null && this.selectedCount() > 0 && !this.facade.exporting(),
  );

  protected readonly filterForm = this.fb.group({
    search: [''],
    statuses: [[] as string[]],
    levels: [[] as string[]],
    date_from: [''],
    date_to: [''],
    sort_by: [''],
    sort_dir: ['asc' as 'asc' | 'desc'],
  });

  protected readonly statusOptions = [
    { value: 'draft', label: 'Черновик' },
    { value: 'pending', label: 'На рассмотрении' },
    { value: 'approved', label: 'Одобрено' },
    { value: 'rejected', label: 'Отклонено' },
    { value: 'in_progress', label: 'В процессе' },
    { value: 'completed', label: 'Завершено' },
    { value: 'cancelled', label: 'Отменено' },
    { value: 'published', label: 'Опубликовано' },
    { value: 'archived', label: 'В архиве' },
  ];

  protected readonly levelOptions = [
    { value: 'junior', label: 'Junior' },
    { value: 'middle', label: 'Middle' },
    { value: 'senior', label: 'Senior' },
    { value: 'lead', label: 'Lead' },
  ];

  protected readonly sortDirOptions = [
    { value: 'asc', label: 'По возрастанию' },
    { value: 'desc', label: 'По убыванию' },
  ];

  ngOnInit(): void {
    this.facade.load();
  }

  protected selectSource(source: SourceMeta): void {
    this.activeSource.set(source);
    this.columnSelections.set(
      source.columns.map((col) => ({ column: col, selected: col.default })),
    );
    this.filterForm.patchValue({ sort_by: '', sort_dir: 'asc' });
  }

  protected toggleColumn(index: number): void {
    this.columnSelections.update((items) => {
      const updated = [...items];
      updated[index] = { ...updated[index], selected: !updated[index].selected };
      return updated;
    });
  }

  protected selectAllColumns(): void {
    this.columnSelections.update((items) => items.map((cs) => ({ ...cs, selected: true })));
  }

  protected deselectAllColumns(): void {
    this.columnSelections.update((items) => items.map((cs) => ({ ...cs, selected: false })));
  }

  protected selectDefaultColumns(): void {
    this.columnSelections.update((items) =>
      items.map((cs) => ({ ...cs, selected: cs.column.default })),
    );
  }

  protected onColumnDrop(event: CdkDragDrop<ColumnSelection[]>): void {
    this.columnSelections.update((items) => {
      const updated = [...items];
      moveItemInArray(updated, event.previousIndex, event.currentIndex);
      return updated;
    });
  }

  protected getColumnTypeIcon(type: string): string {
    switch (type) {
      case 'number':
        return 'tag';
      case 'date':
        return 'calendar_today';
      case 'currency':
        return 'payments';
      default:
        return 'text_fields';
    }
  }

  protected getColumnTypeLabel(type: string): string {
    switch (type) {
      case 'number':
        return 'Число';
      case 'date':
        return 'Дата';
      case 'currency':
        return 'Валюта';
      default:
        return 'Текст';
    }
  }

  protected exportExcel(): void {
    const source = this.activeSource();
    if (!source) return;

    const selected = this.selectedColumns().map((cs) => cs.column.key);
    if (selected.length === 0) return;

    const filters = this.filterForm.getRawValue();

    const request: SmartExportRequest = {
      source: source.key,
      columns: selected,
      format: 'xlsx',
    };

    if (filters.search?.trim()) {
      request.filters = { ...request.filters, search: filters.search.trim() };
    }
    if (filters.statuses && filters.statuses.length > 0) {
      request.filters = { ...request.filters, statuses: filters.statuses };
    }
    if (filters.levels && filters.levels.length > 0) {
      request.filters = { ...request.filters, levels: filters.levels };
    }
    if (filters.date_from) {
      request.filters = { ...request.filters, date_from: filters.date_from };
    }
    if (filters.date_to) {
      request.filters = { ...request.filters, date_to: filters.date_to };
    }
    if (filters.sort_by) {
      request.sort_by = filters.sort_by;
      request.sort_dir = filters.sort_dir || 'asc';
    }

    this.facade.exportToExcel(request);
  }

  protected resetFilters(): void {
    this.filterForm.reset({ search: '', statuses: [], levels: [], date_from: '', date_to: '', sort_by: '', sort_dir: 'asc' });
  }
}
