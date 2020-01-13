package config

import "time"

var (
	// GrpcTimeOut grpc timeout
	GrpcTimeOut = 5000 * time.Millisecond
	// HTTPTimeOut http timeout
	HTTPTimeOut = 5000 * time.Millisecond

	queryHost = "http://yourHostHere"

	// QueryEncryptedAccountURL query encrypted account endpoint
	QueryEncryptedAccountURL = queryHost + "/exp/api/payaddr/query"
	// QueryPlainAccountURL query original account endpoint
	QueryPlainAccountURL = queryHost + "/exp/api/origaddr/query"
	// SaveEncryptedAccountURL save encrypted account endpoint
	SaveEncryptedAccountURL = queryHost + "/exp/api/payaddr/save"

	// APIPublicKey is API publick key
	APIPublicKey = "Input Key Here"
)
