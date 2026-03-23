package ledger

import "github.com/google/uuid"

type Entry struct {
	AccountID         uuid.UUID
	TransferAccountID *uuid.UUID
}

func AffectedAccountIDs(entries ...Entry) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{})
	ids := make([]uuid.UUID, 0, len(entries)*2)
	for _, entry := range entries {
		appendUnique(&ids, seen, entry.AccountID)
		if entry.TransferAccountID != nil {
			appendUnique(&ids, seen, *entry.TransferAccountID)
		}
	}
	return ids
}

func appendUnique(ids *[]uuid.UUID, seen map[uuid.UUID]struct{}, id uuid.UUID) {
	if id == uuid.Nil {
		return
	}
	if _, ok := seen[id]; ok {
		return
	}
	seen[id] = struct{}{}
	*ids = append(*ids, id)
}
