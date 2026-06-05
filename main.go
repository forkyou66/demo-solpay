package main

import (
	"context"
	"encoding/base64"
	"log"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gofiber/fiber/v3"
)

var merchantWallet = solana.MustPublicKeyFromBase58("aNq6vs7ytCUUG4V2qXS12fkN94enx5BJQijy1YjP9va")

type TransactionRequest struct {
	Account string `json:"account"`
}

type TransactionResponse struct {
	Transaction string `json:"transaction"`
	Message     string `json:"message"`
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

func buildSolTransferTx(_ context.Context, sender solana.PublicKey) (*solana.Transaction, error) {
	const lamports = uint64(1_000_000_000) // 1 SOL = 1,000,000,000 lamports

	transferIx, err := system.NewTransferInstruction(
		lamports,
		sender,
		merchantWallet,
	).ValidateAndBuild()
	if err != nil {
		return nil, err
	}

	var zeroHash solana.Hash
	tx, err := solana.NewTransaction(
		[]solana.Instruction{transferIx},
		zeroHash,
		solana.TransactionPayer(sender),
	)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func main() {
	app := fiber.New()
	app.Post("/transaction", postHandler)
	log.Fatal(app.Listen(":7542"))
}
