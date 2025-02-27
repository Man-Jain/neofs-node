package morph

import (
	"errors"
	"fmt"
	"path"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neo-go/pkg/smartcontract"
	"github.com/nspcc-dev/neo-go/pkg/wallet"
	"github.com/nspcc-dev/neofs-node/cmd/neofs-adm/internal/modules/config"
	"github.com/nspcc-dev/neofs-node/pkg/innerring"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	singleAccountName    = "single"
	committeeAccountName = "committee"
	consensusAccountName = "consensus"
)

func generateAlphabetCreds(cmd *cobra.Command, args []string) error {
	// alphabet size is not part of the config
	size, err := cmd.Flags().GetUint(alphabetSizeFlag)
	if err != nil {
		return err
	}
	if size == 0 {
		return errors.New("size must be > 0")
	}

	walletDir := viper.GetString(alphabetWalletsFlag)
	pwds, err := initializeWallets(walletDir, int(size))
	if err != nil {
		return err
	}

	cmd.Println("size:", size)
	cmd.Println("alphabet-wallets:", walletDir)
	for i := range pwds {
		cmd.Printf("wallet[%d]: %s\n", i, pwds[i])
	}

	return nil
}

func initializeWallets(walletDir string, size int) ([]string, error) {
	wallets := make([]*wallet.Wallet, size)
	pubs := make(keys.PublicKeys, size)
	passwords := make([]string, size)

	for i := range wallets {
		password, err := config.AlphabetPassword(viper.GetViper(), i)
		if err != nil {
			return nil, fmt.Errorf("can't fetch password: %w", err)
		}

		p := path.Join(walletDir, innerring.GlagoliticLetter(i).String()+".json")
		// TODO(@fyrchik): file is created with 0666 permissions, consider changing.
		w, err := wallet.NewWallet(p)
		if err != nil {
			return nil, fmt.Errorf("can't create wallet: %w", err)
		}
		if err := w.CreateAccount(singleAccountName, password); err != nil {
			return nil, fmt.Errorf("can't create account: %w", err)
		}

		passwords[i] = password
		wallets[i] = w
		pubs[i] = w.Accounts[0].PrivateKey().PublicKey()
	}

	// Create committee account with N/2+1 multi-signature.
	majCount := smartcontract.GetMajorityHonestNodeCount(size)
	for i, w := range wallets {
		if err := addMultisigAccount(w, majCount, committeeAccountName, passwords[i], pubs); err != nil {
			return nil, fmt.Errorf("can't create committee account: %w", err)
		}
	}

	// Create consensus account with 2*N/3+1 multi-signature.
	bftCount := smartcontract.GetDefaultHonestNodeCount(size)
	for i, w := range wallets {
		if err := addMultisigAccount(w, bftCount, consensusAccountName, passwords[i], pubs); err != nil {
			return nil, fmt.Errorf("can't create consensus account: %w", err)
		}
	}

	for _, w := range wallets {
		if err := w.Save(); err != nil {
			return nil, fmt.Errorf("can't save wallet: %w", err)
		}
		w.Close()
	}

	return passwords, nil
}

func addMultisigAccount(w *wallet.Wallet, m int, name, password string, pubs keys.PublicKeys) error {
	acc := wallet.NewAccountFromPrivateKey(w.Accounts[0].PrivateKey())
	acc.Label = name

	if err := acc.ConvertMultisig(m, pubs); err != nil {
		return err
	}
	if err := acc.Encrypt(password, keys.NEP2ScryptParams()); err != nil {
		return err
	}
	w.AddAccount(acc)
	return nil
}

func generateStorageCreds(cmd *cobra.Command, args []string) error {
	// storage wallet path is not part of the config
	storageWalletPath, err := cmd.Flags().GetString(storageWalletFlag)
	if err != nil {
		return err
	}

	cmd.Println("endpoint:", viper.GetString(endpointFlag))
	cmd.Println("alphabet-wallets:", viper.GetString(alphabetWalletsFlag))
	cmd.Println("storage-wallet:", storageWalletPath)

	return nil
}
