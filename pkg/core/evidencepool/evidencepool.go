package evidencepool

import (
	"log"
	"sync"
	"time"

	"github.com/StupidBug/fabric-zkrollup/pkg/chaincode"
	"github.com/StupidBug/fabric-zkrollup/pkg/core/storage"
	"github.com/StupidBug/fabric-zkrollup/pkg/types/evidence"
	"github.com/StupidBug/fabric-zkrollup/pkg/zk"
)

const (
	maxPoolSize = 50
)

type EvidencePool struct {
	Evidence        []*evidence.Evidence
	appendMutex     sync.Mutex
	packageCh       chan struct{}
	storage         *storage.Storage
	lastPackageTime time.Time
}

func NewEvidencePool() *EvidencePool {
	genesisStateRoot := zk.ComputeStorageMerkleRoot([]string{"initial"})
	log.Printf("genesis state root %s", genesisStateRoot)
	return &EvidencePool{
		Evidence:  make([]*evidence.Evidence, 0, 100),
		storage:   storage.NewStorage(genesisStateRoot),
		packageCh: make(chan struct{}),
	}
}

func (ep *EvidencePool) Add(proofs []string) ([]string, error) {
	log.Printf("append proofs %v", proofs)
	hashs := make([]string, 0, len(proofs))
	for _, proof := range proofs {
		evidence := evidence.NewEvidence(proof)
		hashs = append(hashs, evidence.Hash)
		ep.appendMutex.Lock()
		ep.Evidence = append(ep.Evidence, evidence)
		ep.appendMutex.Unlock()
	}
	if len(ep.Evidence) >= maxPoolSize {
		ep.packageCh <- struct{}{}
	}
	return hashs, nil
}

func (ep *EvidencePool) PopAll() []*evidence.Evidence {
	ep.appendMutex.Lock()
	all := ep.Evidence
	ep.Evidence = make([]*evidence.Evidence, 0, 100)
	ep.appendMutex.Unlock()
	return all
}

// StartAutoBlock starts automatic block creation
func (ep *EvidencePool) StartAutoPackage() {
	ticker := time.NewTicker(2 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				ep.Package()
			case <-ep.packageCh:
				ep.Package()
			}
		}
	}()
}

func (ep *EvidencePool) Package() {
	if time.Since(ep.lastPackageTime) < 2*time.Second && len(ep.Evidence) < maxPoolSize {
		return
	}
	all := ep.PopAll()
	if len(all) == 0 {
		log.Print("pool dont have proofs")
		return
	}
	proofs := make([]string, 0, len(all))
	for _, evidence := range all {
		proofs = append(proofs, evidence.Proof)
	}
	input := zk.StorageProofInput{
		OldStateRoot: ep.storage.LastStateRoot(),
		Evidence:     proofs,
	}
	output, err := zk.Package(input)
	if err != nil {
		log.Fatalf("generate proof failed [input: %#v]: %s", input, err.Error())
	}
	chaincode.Validate(output)
	ep.storage.AddBatch(all, output.NewStateRoot)
}

func (ep *EvidencePool) Query(hash string) string {
	return ep.storage.GetEvidenceStatus(hash)
}
