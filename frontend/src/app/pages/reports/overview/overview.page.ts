import { CommonModule } from '@angular/common';
import { CdkDragDrop, DragDropModule, moveItemInArray } from '@angular/cdk/drag-drop';
import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatCheckboxModule } from '@angular/material/checkbox';
import { MatChipsModule } from '@angular/material/chips';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatSelectModule } from '@angular/material/select';
import type { ColumnMeta, SmartExportRequest, SourceMeta } from '@entities/smart-export';
import { ReportsAnalyticsFacade } from '@features/reports-analytics';

interface ColumnSelection {
  column: ColumnMeta;
  selected: boolean;
}

interface PreviewRow {
  id: number;
  cells: string[];
}

@Component({
  selector: 'app-page-reports-overview',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    MatButtonModule,
    MatCheckboxModule,
    MatChipsModule,
    MatFormFieldModule,
    MatIconModule,
    MatInputModule,
    MatProgressSpinnerModule,
    MatSelectModule,
    DragDropModule,
  ],
  templateUrl: './overview.page.html',
  styleUrl: './overview.page.scss',
})
export class ReportsOverviewPageComponent implements OnInit {
  protected readonly facade = inject(ReportsAnalyticsFacade);
  private readonly fb = inject(FormBuilder);

  private readonly previewColumnLimit = 6;
  private readonly previewRowCount = 6;
  private readonly excludedColumnLabels = new Set(['должность']);
  private readonly excludedColumnKeys = new Set(['position', 'job_title', 'jobtitle']);

  private readonly mockNames = [
    'Anna Petrova',
    'Ivan Smirnov',
    'Mariya Fedorova',
    'Sergey Volkow',
    'Nina Kovaleva',
    'Maksim Doronin',
  ];

  private readonly mockEmails = [
    'anna.p@company.ru',
    'ivan.s@company.ru',
    'mariya.f@company.ru',
    'sergey.v@company.ru',
    'nina.k@company.ru',
    'maksim.d@company.ru',
  ];

  private readonly mockStatuses = ['Approved', 'Pending', 'In progress', 'Draft', 'Completed', 'Approved'];
  private readonly mockLevels = ['Junior', 'Middle', 'Senior', 'Middle', 'Lead', 'Senior'];
  private readonly mockDepartments = ['HR', 'Finance', 'Operations', 'Product', 'Sales', 'IT'];
  private readonly mockCourses = [
    'Leadership Basics',
    'Excel Pro',
    'Data Mindset',
    'Project Delivery',
    'Negotiation 2.0',
    'Team Growth',
  ];

  private readonly mockManagers = [
    'Olga Kim',
    'Dmitriy Sokolov',
    'Elena Ivanova',
    'Roman Egorov',
    'Pavel Kozhin',
    'Irina Sova',
  ];

  private readonly mockMoney = [
    '126 000 RUB',
    '94 500 RUB',
    '188 000 RUB',
    '73 200 RUB',
    '142 900 RUB',
    '110 400 RUB',
  ];

  private readonly mockDates = [
    '2026-03-02',
    '2026-03-05',
    '2026-03-08',
    '2026-03-11',
    '2026-03-14',
    '2026-03-18',
  ];

  protected readonly columnSelections = signal<ColumnSelection[]>([]);
  protected readonly activeSource = signal<SourceMeta | null>(null);

  protected readonly selectedColumns = computed(() =>
    this.columnSelections().filter((columnSelection) => columnSelection.selected),
  );

  protected readonly selectedCount = computed(() => this.selectedColumns().length);
  protected readonly totalCount = computed(() => this.columnSelections().length);

  protected readonly previewColumns = computed(() => {
    const selected = this.selectedColumns().map((columnSelection) => columnSelection.column);
    if (selected.length > 0) {
      return selected.slice(0, this.previewColumnLimit);
    }

    const defaults = this.columnSelections()
      .filter((columnSelection) => columnSelection.column.default)
      .map((columnSelection) => columnSelection.column);

    if (defaults.length > 0) {
      return defaults.slice(0, this.previewColumnLimit);
    }

    return this.columnSelections()
      .slice(0, this.previewColumnLimit)
      .map((columnSelection) => columnSelection.column);
  });

  protected readonly previewRows = computed<PreviewRow[]>(() => {
    const columns = this.previewColumns();
    if (columns.length === 0) {
      return [];
    }

    return Array.from({ length: this.previewRowCount }, (_, rowIndex) => ({
      id: rowIndex + 1,
      cells: columns.map((column) => this.getMockValue(column, rowIndex)),
    }));
  });

  protected readonly previewTableMinWidth = computed(() =>
    Math.max(980, this.previewColumns().length * 180),
  );

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
    const columns = source.columns.filter((column) => this.isColumnAllowed(column));

    this.activeSource.set(source);
    this.columnSelections.set(
      columns.map((column) => ({ column, selected: column.default })),
    );

    this.filterForm.patchValue({
      sort_by: '',
      sort_dir: 'asc',
    });
  }

  protected toggleColumn(index: number): void {
    this.columnSelections.update((items) => {
      const updated = [...items];
      updated[index] = { ...updated[index], selected: !updated[index].selected };
      return updated;
    });
  }

  protected selectAllColumns(): void {
    this.columnSelections.update((items) => items.map((columnSelection) => ({ ...columnSelection, selected: true })));
  }

  protected deselectAllColumns(): void {
    this.columnSelections.update((items) => items.map((columnSelection) => ({ ...columnSelection, selected: false })));
  }

  protected selectDefaultColumns(): void {
    this.columnSelections.update((items) =>
      items.map((columnSelection) => ({ ...columnSelection, selected: columnSelection.column.default })),
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
    if (!source) {
      return;
    }

    const selectedColumns = this.selectedColumns().map((columnSelection) => columnSelection.column.key);
    if (selectedColumns.length === 0) {
      return;
    }

    const filters = this.filterForm.getRawValue();

    const request: SmartExportRequest = {
      source: source.key,
      columns: selectedColumns,
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
      const dateFrom = this.normalizeDateFromFilter(filters.date_from);
      if (dateFrom) {
        request.filters = { ...request.filters, date_from: dateFrom };
      }
    }
    if (filters.date_to) {
      const dateTo = this.normalizeDateToFilter(filters.date_to);
      if (dateTo) {
        request.filters = { ...request.filters, date_to: dateTo };
      }
    }
    if (filters.sort_by) {
      request.sort_by = filters.sort_by;
      request.sort_dir = filters.sort_dir || 'asc';
    }

    this.facade.exportToExcel(request);
  }

  protected resetFilters(): void {
    this.filterForm.reset({
      search: '',
      statuses: [],
      levels: [],
      date_from: '',
      date_to: '',
      sort_by: '',
      sort_dir: 'asc',
    });
  }

  private normalizeDateFromFilter(value: string): string | null {
    const date = new Date(`${value}T00:00:00`);
    return Number.isNaN(date.getTime()) ? null : date.toISOString();
  }

  private normalizeDateToFilter(value: string): string | null {
    const date = new Date(`${value}T23:59:59.999`);
    return Number.isNaN(date.getTime()) ? null : date.toISOString();
  }

  private getMockValue(column: ColumnMeta, rowIndex: number): string {
    const key = column.key.toLowerCase();

    if (key.includes('id')) {
      return `EMP-${1040 + rowIndex}`;
    }

    if (key.includes('email')) {
      return this.mockEmails[rowIndex % this.mockEmails.length];
    }

    if (key.includes('status')) {
      return this.mockStatuses[rowIndex % this.mockStatuses.length];
    }

    if (key.includes('level')) {
      return this.mockLevels[rowIndex % this.mockLevels.length];
    }

    if (key.includes('department') || key.includes('team')) {
      return this.mockDepartments[rowIndex % this.mockDepartments.length];
    }

    if (key.includes('course') || key.includes('program') || key.includes('training')) {
      return this.mockCourses[rowIndex % this.mockCourses.length];
    }

    if (key.includes('manager') || key.includes('owner')) {
      return this.mockManagers[rowIndex % this.mockManagers.length];
    }

    if (key.includes('date') || column.type === 'date') {
      return this.mockDates[rowIndex % this.mockDates.length];
    }

    if (key.includes('budget') || key.includes('cost') || key.includes('price') || key.includes('amount') || column.type === 'currency') {
      return this.mockMoney[rowIndex % this.mockMoney.length];
    }

    if (key.includes('percent') || key.includes('rate')) {
      return `${76 + rowIndex * 3}%`;
    }

    if (column.type === 'number') {
      return String(70 + rowIndex * 6);
    }

    return this.mockNames[rowIndex % this.mockNames.length];
  }

  private isColumnAllowed(column: ColumnMeta): boolean {
    const label = column.label.trim().toLowerCase();
    const key = column.key.trim().toLowerCase();

    if (this.excludedColumnLabels.has(label) || label.includes('должность')) {
      return false;
    }

    if (
      this.excludedColumnKeys.has(key) ||
      key.endsWith('_position') ||
      key.includes('job_title')
    ) {
      return false;
    }

    return true;
  }
}
