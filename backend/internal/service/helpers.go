package service

import "github.com/google/uuid"

func uniqueIDs(ids []uuid.UUID) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{}, len(ids))
	out := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if id == uuid.Nil {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

// mergePermissions appends extra codes without duplicating grants that are
// already present and without mutating the input slice.
func mergePermissions(existing, extra []string) []string {
	seen := make(map[string]struct{}, len(existing)+len(extra))
	merged := make([]string, 0, len(existing)+len(extra))
	for _, group := range [][]string{existing, extra} {
		for _, code := range group {
			if _, ok := seen[code]; ok {
				continue
			}
			seen[code] = struct{}{}
			merged = append(merged, code)
		}
	}
	return merged
}
