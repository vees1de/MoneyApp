import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { RouterLink } from '@angular/router';
import { ApprovalsFacade } from '@features/approvals';
import type { ApprovalStep } from '@entities/approval-step';

@Component({
  selector: 'app-page-approvals-inbox',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './inbox.page.html',
  styleUrl: './inbox.page.scss',
})
export class ApprovalsInboxPageComponent {
  private readonly facade = inject(ApprovalsFacade);
  protected readonly routePath = '/approvals/inbox';
  protected readonly entitySample: ApprovalStep[] = [];

  protected loadPage(): void {
    this.facade.load();
  }
}
