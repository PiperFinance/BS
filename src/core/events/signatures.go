package events

import (
	"github.com/ethereum/go-ethereum/crypto"
)

const (

	// Event Names
	// ERC20 + ERC721 + ERC1155

	TransferE       = EventName("Transfer")
	WithdrawalE     = EventName("Withdrawal")
	DepositE        = EventName("Deposit")
	MintE           = EventName("Mint")
	BurnE           = EventName("Burn")
	ApprovalE       = EventName("Approval")
	ApprovalForAllE = EventName("ApprovalForAll")
	URI_E           = EventName("URI")
	TransferBatchE  = EventName("TransferBatch")
	TransferSingleE = EventName("TransferSingle")

	// Event Signatures
	// transfer actual asset + 	approve + NFT related

	TransferESig       = EventSignature("Transfer(address,address,uint256)")
	DepositESig        = EventSignature("Deposit(address,uint256)")
	WithdrawalESig     = EventSignature("Withdrawal(address,uint256)")
	ApprovalESig       = EventSignature("Approval(address,address,uint256)")
	ApprovalForAllESig = EventSignature("ApprovalForAll(address,address,bool)")
	URI_ESig           = EventSignature("URI(string,uint256)")
	TransferBatchESig  = EventSignature("TransferBatch(address,address,address,uint256[],uint256[])")
	TransferSingleESig = EventSignature("TransferSingle(address,address,address,uint256,uint256)")
)

var (
	URIESigHash            = crypto.Keccak256Hash([]byte(URI_ESig))
	DepositESigHash        = crypto.Keccak256Hash([]byte(DepositESig))
	ApprovalESigHash       = crypto.Keccak256Hash([]byte(ApprovalESig))
	TransferESigHash       = crypto.Keccak256Hash([]byte(TransferESig))
	WithdrawalESigHash     = crypto.Keccak256Hash([]byte(WithdrawalESig))
	TransferBatchESigHash  = crypto.Keccak256Hash([]byte(TransferBatchESig))
	ApprovalForAllESigHash = crypto.Keccak256Hash([]byte(ApprovalForAllESig))
	TransferSingleESigHash = crypto.Keccak256Hash([]byte(TransferSingleESig))
)
