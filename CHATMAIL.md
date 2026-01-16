# Chatmail Configuration for Mox

Mox supports chatmail mode, which is designed for use with Delta Chat and similar instant messaging over email clients.

## Features

When chatmail mode is enabled for a domain:

1. **Automatic Account Creation**: Accounts are created automatically on first login attempt via IMAP or SMTP submission
2. **No Spam Filtering**: Junk filtering is disabled for fast message delivery
3. **TLS Required**: All connections require encrypted transport (STARTTLS/TLS)

## Configuration

### Enable Chatmail for a Domain

In your `domains.conf` file, add the `Chatmail` field to the domain configuration:

```
Domains:
	chat.example.com:
		Chatmail: true
		LocalpartCaseSensitive: false
```

### Ensure TLS is Required

In your `mox.conf` file, configure listeners to require TLS:

```
Listeners:
	public:
		IPs:
			- 0.0.0.0
		SMTP:
			Enabled: true
			RequireSTARTTLS: true  # Recommended for incoming SMTP
		Submission:
			Enabled: true
			NoRequireSTARTTLS: false  # Required for chatmail
		IMAP:
			Enabled: true
			NoRequireSTARTTLS: false  # Required for chatmail
		# Or use IMAPS which is always TLS:
		IMAPS:
			Enabled: true
```

## How It Works

### Account Creation

1. A user configures their email client (e.g., Delta Chat) with:
   - Email: `alice@chat.example.com`
   - Password: `their-chosen-password`

2. The client attempts to login via IMAP or SMTP

3. If the account doesn't exist and the domain has `Chatmail: true`:
   - Mox automatically creates the account
   - Sets the user's password
   - Creates the account directory structure
   - Initializes the mailbox database
   - Adds the account to `domains.conf`

4. The user is now logged in and can send/receive messages

### Account Configuration

Auto-created chatmail accounts have the following defaults:

- **No Junk Filter**: `JunkFilter: nil` - All messages are delivered immediately without spam analysis
- **No Rejects Mailbox**: Messages are never stored in a rejects folder
- **Standard Mailboxes**: Inbox and other standard folders are created
- **Account Name**: Uses the localpart of the email address (e.g., "alice" for alice@chat.example.com)

## Security Considerations

### TLS Requirements

Chatmail mode **requires** TLS for all client connections:

- **Submission (SMTP)**: `NoRequireSTARTTLS: false` must be set
- **IMAP**: Either `NoRequireSTARTTLS: false` or use `IMAPS`
- **Incoming SMTP**: `RequireSTARTTLS: true` is recommended

Mox will log warnings during configuration loading if TLS is not properly configured for chatmail domains.

### Password Security

- Passwords are hashed using bcrypt before storage
- Users set their own passwords on first login
- Consider using the `NoCustomPassword` account setting if you want to enforce randomly-generated passwords (not enabled by default for chatmail)

## Example Configuration

### Complete chatmail setup

**mox.conf:**
```
DataDir: /var/lib/mox
LogLevel: info
Hostname: mail.chat.example.com

Listeners:
	public:
		IPs:
			- 0.0.0.0
			- '::'
		SMTP:
			Enabled: true
			Port: 25
			RequireSTARTTLS: true
		Submission:
			Enabled: true
			Port: 587
			NoRequireSTARTTLS: false
		IMAPS:
			Enabled: true
			Port: 993

Postmaster:
	Account: postmaster
	Mailbox: Postmaster
```

**domains.conf:**
```
Domains:
	chat.example.com:
		Chatmail: true
		LocalpartCaseSensitive: false

Accounts:
	postmaster:
		Domain: chat.example.com
		Destinations:
			postmaster@chat.example.com: nil
```

## Differences from Standard Mox Configuration

| Feature | Standard Mox | Chatmail Mode |
|---------|-------------|---------------|
| Account Creation | Manual via admin/API | Automatic on first login |
| Spam Filtering | Enabled by default | Disabled |
| Junk Filter | Bayesian filter | None |
| TLS Requirement | Optional | Required |
| Use Case | General email | Instant messaging |

## Troubleshooting

### Accounts Not Auto-Creating

Check the following:

1. Domain has `Chatmail: true` in domains.conf
2. Listener has submission/IMAP TLS properly configured
3. Check logs for authentication attempts and errors
4. Verify the email domain matches the chatmail domain

### TLS Warnings in Logs

If you see warnings about TLS configuration:

```
chatmail domain requires TLS for submission, but no listener has submission with TLS required
```

Solution: Set `NoRequireSTARTTLS: false` in your submission configuration.

### Permission Errors

Auto-created accounts need write access to the data directory. Ensure:

- Mox process has write permissions to DataDir/accounts/
- File system has sufficient space
- SELinux/AppArmor policies allow account creation

## Migration

### Converting Existing Accounts to Chatmail

If you have existing accounts and want to enable chatmail mode:

1. Enable `Chatmail: true` on the domain
2. Existing accounts continue to work normally  
3. New accounts will be auto-created on first login
4. Optionally remove junk filters from existing accounts in domains.conf

### Disabling Chatmail Mode

To disable chatmail mode:

1. Set `Chatmail: false` on the domain
2. Auto-created accounts remain and continue to work
3. New accounts must be created manually
4. Consider adding junk filters back to accounts

## See Also

- [Delta Chat](https://delta.chat/) - Chat over email
- [Mox Documentation](https://www.xmox.nl)
- [Mox Configuration](https://pkg.go.dev/github.com/mjl-/mox/config)
