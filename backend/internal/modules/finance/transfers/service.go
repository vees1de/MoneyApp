package transfers

import (
	"context"

	"moneyapp/backend/internal/modules/finance/transactions"
	"moneyapp/backend/internal/platform/httpx"

	"github.com/google/uuid"
)

type Service struct {
	transactions *transactions.Service
}

func NewService(transactionService *transactions.Service) *Service {
	return &Service{transactions: transactionService}
}

func (s *Service) Create(ctx context.Context, userID uuid.UUID, request CreateTransferRequest) (transactions.Transaction, error) {
	return s.transactions.Create(ctx, userID, transactions.CreateTransactionRequest{
		AccountID:         request.FromAccountID,
		TransferAccountID: &request.ToAccountID,
		Type:              transactions.TypeTransfer,
		Amount:            request.Amount,
		Currency:          request.Currency,
		Title:             request.Title,
		Note:              request.Note,
		OccurredAt:        request.OccurredAt,
	})
}

func (s *Service) Update(ctx context.Context, userID, transferID uuid.UUID, request UpdateTransferRequest) (transactions.Transaction, error) {
	current, err := s.transactions.Get(ctx, userID, transferID)
	if err != nil {
		return transactions.Transaction{}, err
	}
	if current.Type != transactions.TypeTransfer {
		return transactions.Transaction{}, httpx.BadRequest("not_a_transfer", "transaction is not a transfer")
	}

	return s.transactions.Update(ctx, userID, transferID, transactions.UpdateTransactionRequest{
		AccountID:         request.FromAccountID,
		TransferAccountID: request.ToAccountID,
		Type:              transferTypePtr(),
		Amount:            request.Amount,
		Currency:          request.Currency,
		Title:             request.Title,
		Note:              request.Note,
		OccurredAt:        request.OccurredAt,
	})
}

func (s *Service) Delete(ctx context.Context, userID, transferID uuid.UUID, reason *string) error {
	current, err := s.transactions.Get(ctx, userID, transferID)
	if err != nil {
		return err
	}
	if current.Type != transactions.TypeTransfer {
		return httpx.BadRequest("not_a_transfer", "transaction is not a transfer")
	}
	return s.transactions.Delete(ctx, userID, transferID, reason)
}

func (s *Service) Restore(ctx context.Context, userID, transferID uuid.UUID) (transactions.Transaction, error) {
	current, err := s.transactions.GetIncludingDeleted(ctx, userID, transferID)
	if err != nil {
		return transactions.Transaction{}, err
	}
	if current.Type != transactions.TypeTransfer {
		return transactions.Transaction{}, httpx.BadRequest("not_a_transfer", "transaction is not a transfer")
	}
	return s.transactions.Restore(ctx, userID, transferID)
}

func transferTypePtr() *transactions.Type {
	txType := transactions.TypeTransfer
	return &txType
}
