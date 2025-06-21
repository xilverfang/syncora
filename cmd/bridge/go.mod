module github.com/xilverfang/syncora/cmd/bridge

go 1.24.4

replace github.com/xilverfang/syncora/internal/core/crypto => ../../internal/core/crypto

replace github.com/xilverfang/syncora/internal/core/database => ../../internal/core/database

replace github.com/xilverfang/syncora/internal/bridge-engine => ../../internal/bridge-engine

require (
	github.com/spf13/cobra v1.9.1
	github.com/xilverfang/syncora/internal/core/crypto v0.0.0-00010101000000-000000000000
	github.com/xilverfang/syncora/internal/core/database v0.0.0-00010101000000-000000000000
	golang.org/x/term v0.32.0
)

require (
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.0.1 // indirect
	github.com/ethereum/go-ethereum v1.15.11 // indirect
	github.com/holiman/uint256 v1.3.2 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	golang.org/x/crypto v0.35.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
)
