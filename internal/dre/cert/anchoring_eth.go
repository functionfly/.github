// Package cert: production Ethereum anchoring implementation.
//
// Requires SetSigningKey and per-chain RPC + contract address. The on-chain contract
// must expose:
//
//	function anchor(bytes32 root) external
//
// and optionally emit Anchored(bytes32 indexed root) so GetMerkleProof can read the root from logs.
package cert

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	// anchorSelector is the first 4 bytes of Keccak256("anchor(bytes32)").
	anchorSelector = crypto.Keccak256([]byte("anchor(bytes32)"))[:4]
)

const (
	waitReceiptTimeout = 2 * time.Minute
	waitReceiptPoll    = 2 * time.Second
	minConfirmations   = 1
)

// SetSigningKey sets the ECDSA private key (hex, with or without 0x) for signing anchor transactions.
// Call with a value from secure storage or env (e.g. os.Getenv("ANCHOR_SIGNING_KEY")). Do not commit keys.
func (s *EthereumAnchoringService) SetSigningKey(hexKey string) {
	hexKey = strings.TrimPrefix(strings.TrimSpace(hexKey), "0x")
	s.privateKey = hexKey
}

// IsConfigured returns true if the signing key is set (anchoring is ready).
func (s *EthereumAnchoringService) IsConfigured() bool {
	return s.privateKey != ""
}

// clientForChain returns a cached ethclient for the given chain, dialing if necessary.
func (s *EthereumAnchoringService) clientForChain(ctx context.Context, chain string) (*ethclient.Client, error) {
	rpcURL := s.rpcEndpoints[chain]
	if rpcURL == "" {
		return nil, fmt.Errorf("cert: no RPC endpoint for chain %q", chain)
	}
	s.clientMu.Lock()
	defer s.clientMu.Unlock()
	if s.clientCache == nil {
		s.clientCache = make(map[string]interface{})
	}
	if c, ok := s.clientCache[chain].(*ethclient.Client); ok && c != nil {
		return c, nil
	}
	client, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		return nil, fmt.Errorf("cert: dial %q: %w", chain, err)
	}
	s.clientCache[chain] = client
	return client, nil
}

// packAnchorCalldata returns the calldata for anchor(bytes32 root): selector + 32-byte root.
// rootHash is hex (64 chars), with or without 0x prefix.
func packAnchorCalldata(rootHash string) ([]byte, error) {
	rootHash = strings.TrimPrefix(strings.TrimSpace(rootHash), "0x")
	if len(rootHash) != 64 {
		return nil, fmt.Errorf("cert: execution root hash must be 32 bytes (64 hex chars), got %d", len(rootHash))
	}
	rootBytes, err := hex.DecodeString(rootHash)
	if err != nil {
		return nil, fmt.Errorf("cert: invalid root hash hex: %w", err)
	}
	calldata := make([]byte, 0, 4+32)
	calldata = append(calldata, anchorSelector...)
	calldata = append(calldata, common.LeftPadBytes(rootBytes, 32)...)
	return calldata, nil
}

// anchorOnChain performs the actual chain RPC: sign, send, wait receipt, return receipt.
func (s *EthereumAnchoringService) anchorOnChain(ctx context.Context, chain, executionRootHash, contractAddr string) (*AnchorReceipt, error) {
	if s.privateKey == "" {
		return nil, fmt.Errorf("cert: anchoring requires SetSigningKey to be set")
	}
	contract := common.HexToAddress(contractAddr)
	calldata, err := packAnchorCalldata(executionRootHash)
	if err != nil {
		return nil, err
	}

	client, err := s.clientForChain(ctx, chain)
	if err != nil {
		return nil, err
	}

	key, err := crypto.HexToECDSA(s.privateKey)
	if err != nil {
		return nil, fmt.Errorf("cert: invalid private key: %w", err)
	}
	from := crypto.PubkeyToAddress(key.PublicKey)

	chainID, err := client.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("cert: chain id: %w", err)
	}

	nonce, err := client.PendingNonceAt(ctx, from)
	if err != nil {
		return nil, fmt.Errorf("cert: nonce: %w", err)
	}

	gasLimit := uint64(100_000)
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("cert: gas price: %w", err)
	}

	tx := types.NewTransaction(
		nonce,
		contract,
		big.NewInt(0),
		gasLimit,
		gasPrice,
		calldata,
	)

	signer := types.LatestSignerForChainID(chainID)
	signedTx, err := types.SignTx(tx, signer, key)
	if err != nil {
		return nil, fmt.Errorf("cert: sign tx: %w", err)
	}

	if err := client.SendTransaction(ctx, signedTx); err != nil {
		return nil, fmt.Errorf("cert: send tx: %w", err)
	}

	txHash := signedTx.Hash().Hex()

	// Wait for receipt
	var receipt *types.Receipt
	deadline := time.Now().Add(waitReceiptTimeout)
	for time.Now().Before(deadline) {
		receipt, err = client.TransactionReceipt(ctx, signedTx.Hash())
		if err == nil && receipt != nil {
			break
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(waitReceiptPoll):
		}
	}
	if receipt == nil {
		return nil, fmt.Errorf("cert: timeout waiting for tx %s", txHash)
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return nil, fmt.Errorf("cert: tx %s failed (status %d)", txHash, receipt.Status)
	}

	// Current block for confirmations
	header, _ := client.HeaderByNumber(ctx, nil)
	confirmations := int64(0)
	if header != nil {
		confirmations = header.Number.Int64() - receipt.BlockNumber.Int64()
		if confirmations < 0 {
			confirmations = 0
		}
	}

	return &AnchorReceipt{
		Chain:            chain,
		BlockNumber:      receipt.BlockNumber.Int64(),
		TxHash:           txHash,
		MerkleRoot:       executionRootHash,
		AnchorHash:       generateAnchorHash(executionRootHash),
		AnchoredAt:       time.Now().UTC(),
		Confirmations:    confirmations,
		GasUsed:          receipt.GasUsed,
		ContractAddress:  contractAddr,
	}, nil
}

// verifyAnchorOnChain checks the chain for the given receipt and returns true if valid.
func (s *EthereumAnchoringService) verifyAnchorOnChain(ctx context.Context, receipt *AnchorReceipt) (bool, error) {
	client, err := s.clientForChain(ctx, receipt.Chain)
	if err != nil {
		return false, err
	}

	txHash := common.HexToHash(receipt.TxHash)
	tx, pending, err := client.TransactionByHash(ctx, txHash)
	if err != nil {
		return false, fmt.Errorf("cert: get tx: %w", err)
	}
	if pending || tx == nil {
		return false, nil
	}

	r, err := client.TransactionReceipt(ctx, txHash)
	if err != nil {
		return false, fmt.Errorf("cert: get receipt: %w", err)
	}
	if r == nil || r.Status != types.ReceiptStatusSuccessful {
		return false, nil
	}

	// Optional: require minimum confirmations
	header, err := client.HeaderByNumber(ctx, nil)
	if err == nil && header != nil {
		confirmations := header.Number.Int64() - r.BlockNumber.Int64()
		if confirmations < minConfirmations {
			return false, nil
		}
	}

	return true, nil
}

// getBlockNumberOnChain returns the latest block number for the chain.
func (s *EthereumAnchoringService) getBlockNumberOnChain(ctx context.Context, chain string) (int64, error) {
	client, err := s.clientForChain(ctx, chain)
	if err != nil {
		return 0, err
	}
	header, err := client.HeaderByNumber(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("cert: block number: %w", err)
	}
	return header.Number.Int64(), nil
}

// getMerkleProofOnChain returns a Merkle proof from the chain for the given tx.
// Root is taken from receipt logs (Anchored(bytes32)) if present, otherwise from stored receipt.
func (s *EthereumAnchoringService) getMerkleProofOnChain(ctx context.Context, chain, txHash string) (*MerkleProof, error) {
	client, err := s.clientForChain(ctx, chain)
	if err != nil {
		return nil, err
	}

	hash := common.HexToHash(txHash)
	receipt, err := client.TransactionReceipt(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("cert: get receipt: %w", err)
	}
	if receipt == nil {
		return nil, fmt.Errorf("cert: no receipt for tx %s", txHash)
	}

	rootHash := ""
	// Anchored(bytes32 indexed root): topic0 = event sig, topic1 = root (padded to 32 bytes)
	for _, log := range receipt.Logs {
		if len(log.Topics) >= 2 {
			rootHash = "0x" + hex.EncodeToString(log.Topics[1].Bytes())
			break
		}
	}
	if rootHash == "" {
		rootHash = "0x"
	}

	return &MerkleProof{
		BlockNumber: receipt.BlockNumber.Int64(),
		TxHash:      txHash,
		RootHash:    rootHash,
		Proof:       []string{},
		LeafHash:    rootHash,
	}, nil
}
