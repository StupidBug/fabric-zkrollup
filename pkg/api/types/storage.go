package types

type ProofStorageReq struct {
	Evidence []string `json:"evidences"`
}

type ProofStorageResp struct {
	Hashs []string `json:"hashs"`
}

type ProofStatusResp struct {
	Status string `json:"status"`
}
