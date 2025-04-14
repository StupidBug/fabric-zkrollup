package evidence

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/StupidBug/fabric-zkrollup/pkg/types/status"
)

type Evidence struct {
	Proof  string
	Hash   string
	Status status.Status
}

func NewEvidence(proof string) *Evidence {
	return &Evidence{
		Proof:  proof,
		Status: status.StatusPending,
		Hash:   ComputeHash(proof),
	}
}

func ComputeHash(proof string) string {
	data := fmt.Appendf(nil, "%s", proof)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
