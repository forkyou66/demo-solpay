package main

import (
	"context"
	"encoding/base64"
	"log"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gofiber/fiber/v3"
)

const oneSol = 1_000_000_000

var (
	rpcClient     = rpc.New(rpc.DevNet_RPC)
	merchantWallet = solana.MustPublicKeyFromBase58("aNq6vs7ytCUUG4V2qXS12fkN94enx5BJQijy1YjP9va")
)

type TransactionRequest struct {
	Account string `json:"account"`
}

type TransactionResponse struct {
	Transaction string `json:"transaction"`
	Message     string `json:"message"`
}

type GetResponse struct {
	Label string `json:"label"`
	Icon  string `json:"icon"`
}

func getHandler(c fiber.Ctx) error {
	return c.JSON(GetResponse{
		Label: "Exiled Apes Academy",
		Icon:  "https://exiledapes.academy/wp-content/uploads/2021/09/X_share.png",
	})
}

func postHandler(c fiber.Ctx) error {
	var body TransactionRequest
	if err := c.Bind().Body(&body); err != nil || body.Account == "" {
		return fiber.NewError(fiber.StatusBadRequest, "missing account")
	}

	sender, err := solana.PublicKeyFromBase58(body.Account)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid account")
	}

	tx, err := buildSolTransferTx(context.Background(), sender)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	serialized, err := tx.MarshalBinary()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to serialize transaction")
	}

	return c.JSON(TransactionResponse{
		Transaction: base64.StdEncoding.EncodeToString(serialized),
		Message:     "Thank you for your purchase of ExiledApe #518",
	})
}

func buildMemoInstruction(memo string, sender solana.PublicKey) solana.Instruction {
	return solana.NewInstruction(
		solana.MemoProgramID,
		solana.AccountMetaSlice{
			solana.NewAccountMeta(sender, true, true),
		},
		[]byte(memo),
	)
}


func buildSolTransferTx(ctx context.Context, sender solana.PublicKey) (*solana.Transaction, error) {
	const lamports = uint64(oneSol / 100) // 0.01 SOL

	transferIx, err := system.NewTransferInstruction(
		lamports,
		sender,
		merchantWallet,
	).ValidateAndBuild()
	if err != nil {
		return nil, err
	}

	memoIx := buildMemoInstruction("RqSfVF1fNFXk5QrMMXc6YbascbKceAVXn7Trw3776vP4HM44Q", sender)
	recent, err := rpcClient.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
    if err != nil {
        return nil, err
    }

	tx, err := solana.NewTransaction(
		[]solana.Instruction{transferIx, memoIx},
		recent.Value.Blockhash,
		solana.TransactionPayer(sender),
	)

	if err != nil {
		return nil, err
	}

	return tx, nil
}

func healthHandler(c fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "ok",
	})
}

func main() {
	app := fiber.New()
	app.Get("/", healthHandler)
	app.Get("/transaction", getHandler)
	app.Post("/transaction", postHandler)
	log.Fatal(app.Listen(":7542"))
}
