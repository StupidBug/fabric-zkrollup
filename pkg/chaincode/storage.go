package chaincode

import "github.com/StupidBug/fabric-zkrollup/pkg/zk"

type StorageClient interface {
	func(*zk.StorageProofOutput) error
}
