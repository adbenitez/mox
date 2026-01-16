package store

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"golang.org/x/crypto/bcrypt"

	"github.com/mjl-/bstore"

	"github.com/mjl-/mox/config"
	"github.com/mjl-/mox/mlog"
	"github.com/mjl-/mox/mox-"
	"github.com/mjl-/mox/smtp"
)

// AutoCreateAccount creates a new account on-the-fly for chatmail usage.
// This is called during authentication when an account doesn't exist yet but
// the domain has chatmail mode enabled.
//
// Returns the newly created account and account name, or an error.
func AutoCreateAccount(log mlog.Log, email, password string) (*Account, string, error) {
	addr, err := smtp.ParseAddress(email)
	if err != nil {
		return nil, "", fmt.Errorf("%w: parsing email address: %v", ErrUnknownCredentials, err)
	}

	// Check if domain is configured with chatmail mode
	domainConf, ok := mox.Conf.Domain(addr.Domain)
	if !ok {
		return nil, "", fmt.Errorf("%w: domain not configured", ErrUnknownCredentials)
	}
	if !domainConf.Chatmail {
		return nil, "", fmt.Errorf("%w: auto-create not enabled for domain", ErrUnknownCredentials)
	}

	// Generate a unique account name based on the email address
	// Use the localpart and a hash to avoid conflicts
	accountName := addr.Localpart.String()

	// Lock configuration to check and add account atomically
	defer mox.Conf.DynamicLockUnlock()()

	c := mox.Conf.Dynamic
	if _, exists := c.Accounts[accountName]; exists {
		// Race condition: account was created between check and now
		// Try to open and authenticate with the existing account
		acc, err := OpenAccount(log, accountName, false)
		if err != nil {
			return nil, "", fmt.Errorf("opening newly created account: %w", err)
		}
		defer func() {
			if err != nil {
				closeErr := acc.Close()
				log.Check(closeErr, "closing account after auth check")
			}
		}()

		// Verify password against existing account
		pw, err := bstore.QueryDB[Password](context.TODO(), acc.DB).Get()
		if err != nil {
			return nil, "", fmt.Errorf("checking password: %w", err)
		}
		if err := bcrypt.CompareHashAndPassword([]byte(pw.Hash), []byte(password)); err != nil {
			return nil, "", ErrUnknownCredentials
		}
		return acc, accountName, nil
	}

	// Ensure the directory does not exist
	accountDir := filepath.Join(mox.DataDirPath("accounts"), accountName)
	if _, err := os.Stat(accountDir); err == nil {
		return nil, "", fmt.Errorf("account directory already exists")
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, "", fmt.Errorf("checking account directory: %v", err)
	}

	// Create account configuration for chatmail - without junk filter
	accConf := config.Account{
		Domain: addr.Domain.Name(),
		Destinations: map[string]config.Destination{
			addr.String(): {},
		},
		// For chatmail, no junk filter
		JunkFilter: nil,
		// Allow messages to be rejected to a mailbox, not used in chatmail but keeping for compatibility
		RejectsMailbox: "",
		// Disable custom passwords - users get what they set
		NoCustomPassword: false,
	}

	// Helper to build updated config with account added/removed
	buildUpdatedConfig := func(addAccount bool) config.Dynamic {
		updated := config.Dynamic{
			Domains: c.Domains,
			Accounts: map[string]config.Account{},
		}
		for name, a := range c.Accounts {
			updated.Accounts[name] = a
		}
		if addAccount {
			updated.Accounts[accountName] = accConf
		} else {
			delete(updated.Accounts, accountName)
		}
		return updated
	}

	// Add account to configuration
	nc := buildUpdatedConfig(true)

	// Write updated configuration
	if err := mox.WriteDynamicLocked(context.Background(), log, nc); err != nil {
		return nil, "", fmt.Errorf("writing domains.conf: %w", err)
	}

	// Open the newly created account (this will initialize the database)
	acc, err := OpenAccount(log, accountName, false)
	if err != nil {
		// Try to remove the account from config on failure
		nc = buildUpdatedConfig(false)
		_ = mox.WriteDynamicLocked(context.Background(), log, nc)
		return nil, "", fmt.Errorf("opening new account: %w", err)
	}

	// Set the password
	if err := acc.SetPassword(log, password); err != nil {
		// Clean up on failure
		acc.Close()
		nc = buildUpdatedConfig(false)
		_ = mox.WriteDynamicLocked(context.Background(), log, nc)
		os.RemoveAll(accountDir)
		return nil, "", fmt.Errorf("setting password: %w", err)
	}

	log.Info("auto-created chatmail account", slog.String("account", accountName), slog.String("email", email))
	return acc, accountName, nil
}
