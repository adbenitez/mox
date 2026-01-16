package store

import (
	"path/filepath"
	"testing"

	"github.com/mjl-/mox/dns"
	"github.com/mjl-/mox/mox-"
)

func TestChatmailConfig(t *testing.T) {
	// Test that chatmail configuration is properly read
	mox.Context = ctxbg
	mox.ConfigStaticPath = filepath.FromSlash("../testdata/chatmail/mox.conf")
	mox.ConfigDynamicPath = filepath.FromSlash("../testdata/chatmail/domains.conf")
	mox.MustLoadConfig(true, false)

	domain, err := dns.ParseDomain("chatmail.example")
	tcheck(t, err, "parsing domain")

	conf, ok := mox.Conf.Domain(domain)
	if !ok {
		t.Fatal("chatmail.example domain not found in config")
	}

	if !conf.Chatmail {
		t.Fatal("expected Chatmail to be true for chatmail.example")
	}

	// Check regular domain doesn't have chatmail enabled
	domain2, err := dns.ParseDomain("regular.example")
	tcheck(t, err, "parsing regular domain")

	conf2, ok := mox.Conf.Domain(domain2)
	if !ok {
		t.Fatal("regular.example domain not found in config")
	}

	if conf2.Chatmail {
		t.Fatal("expected Chatmail to be false for regular.example")
	}

	t.Log("Chatmail configuration test passed - auto-create functionality requires integration test")
}
