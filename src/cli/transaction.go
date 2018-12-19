package cli

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/skycoin/skycoin/src/api/webrpc"
	"github.com/skycoin/skycoin/src/cipher"
	"github.com/skycoin/skycoin/src/coin"
	"github.com/skycoin/skycoin/src/readable"

	gcli "github.com/urfave/cli"
)

func transactionCmd() gcli.Command {
	name := "transaction"
	return gcli.Command{
		Name:         name,
		Usage:        "Show detail info of specific transaction",
		ArgsUsage:    "[transaction id]",
		OnUsageError: onCommandUsageError(name),
		Action: func(c *gcli.Context) error {
			txid := c.Args().First()
			if txid == "" {
				return errors.New("txid is empty")
			}

			// validate the txid
			_, err := cipher.SHA256FromHex(txid)
			if err != nil {
				return errors.New("invalid txid")
			}

			client := APIClientFromContext(c)

			txn, err := client.Transaction(txid)
			if err != nil {
				return err
			}

			return printJSON(webrpc.TxnResult{
				Transaction: txn,
			})
		},
	}
}

func decodeRawTxCmd() gcli.Command {
	name := "decodeRawTransaction"
	return gcli.Command{
		Name:         name,
		Usage:        "Decode raw transaction",
		ArgsUsage:    "[raw transaction]",
		OnUsageError: onCommandUsageError(name),
		Action: func(c *gcli.Context) error {
			rawTxStr := c.Args().First()
			if rawTxStr == "" {
				printHelp(c)
				return errors.New("missing raw transaction value")
			}

			b, err := hex.DecodeString(rawTxStr)
			if err != nil {
				return fmt.Errorf("invalid raw transaction: %v", err)
			}

			txn, err := coin.TransactionDeserialize(b)
			if err != nil {
				return fmt.Errorf("Unable to deserialize transaction bytes: %v", err)
			}

			// Assume the transaction is not malformed and if it has no inputs
			// that it is the genesis block's transaction
			isGenesis := len(txn.In) == 0
			rTxn, err := readable.NewTransaction(txn, isGenesis)
			if err != nil {
				return err
			}

			return printJSON(rTxn)
		},
	}
}

func addressTransactionsCmd() gcli.Command {
	name := "addressTransactions"
	return gcli.Command{
		Name:      name,
		Usage:     "Show detail for transaction associated with one or more specified addresses",
		ArgsUsage: "[address list]",
		Description: `Display transactions for specific addresses, seperate multiple addresses with a space,
        example: addressTransactions addr1 addr2 addr3`,
		OnUsageError: onCommandUsageError(name),
		Action:       getAddressTransactionsCmd,
	}
}

func getAddressTransactionsCmd(c *gcli.Context) error {
	client := APIClientFromContext(c)

	addrs := make([]string, c.NArg())
	var err error
	for i := 0; i < c.NArg(); i++ {
		addrs[i] = c.Args().Get(i)
		if _, err = cipher.DecodeBase58Address(addrs[i]); err != nil {
			return fmt.Errorf("invalid address: %v, err: %v", addrs[i], err)
		}
	}

	outputs, err := client.GetTransactions(addrs)
	if err != nil {
		return err
	}

	return printJSON(outputs)
}
