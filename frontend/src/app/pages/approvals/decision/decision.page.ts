import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { RouterLink } from '@angular/router';
import { ApprovalsFacade } from '@features/approvals';
import type { ApprovalStep } from '@entities/approval-step';

@Component({
  selector: 'app-page-approvals-decision',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './decision.page.html',
  styleUrl: './decision.page.scss',
})
export class ApprovalsDecisionPageComponent {
  private readonly facade = inject(ApprovalsFacade);
  protected readonly routePath = '/approvals/decision';
  protected readonly entitySample: ApprovalStep[] = [];

  protected loadPage(): void {
    this.facade.load();
  }
}
