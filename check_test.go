package main

import (
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
	"time"
)

const DefaultTimeout = 5 * time.Second

func TestConfig_Validate(t *testing.T) {
	c := &Config{}
	assert.Error(t, c.Validate())

	// Most basic settings
	c.Host = "localhost"
	c.Command = "Get-Something"
	c.User = "administrator"
	c.Password = "verysecret"

	assert.NoError(t, c.Validate())
	assert.Equal(t, c.Port, Port)
	assert.Equal(t, c.AuthType, AuthBasic)
	assert.True(t, c.validated)
}

func TestBuildConfigFlags(t *testing.T) {
	fs := &pflag.FlagSet{}
	config := BuildConfigFlags(fs)

	assert.True(t, fs.HasFlags())
	assert.False(t, config.validated)
}

func TestConfig_BuildCommand(t *testing.T) {
	c := &Config{Command: "Get-Something"}
	assert.Contains(t, c.BuildCommand(), "powershell.exe")
	assert.Contains(t, c.BuildCommand(), "dAByA") // try {

	c = &Config{IcingaCommand: "Icinga-CheckSomething"}
	assert.Contains(t, c.BuildCommand(), "powershell.exe")
	assert.Contains(t, c.BuildCommand(), "dAByAHkAIAB7ACAAVQBzAGUALQBJAGMAaQBuAGcAYQA7") // try { Use-Icinga;
}

func TestConfig_Run_WithError(t *testing.T) {
	c := &Config{
		Host:     "192.0.2.11",
		User:     "admin",
		Password: "test",
		Command:  "Get-Host",
	}

	err := c.Validate()
	assert.NoError(t, err)

	err, _, _ = c.Run(1 * time.Microsecond)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dial tcp 192.0.2.11:")
}

func TestConfig_Run_Basic(t *testing.T) {
	c := buildEnvConfig(t, AuthBasic)

	runCheck(t, c)
}

func TestConfig_Run_Basic_WithTLS(t *testing.T) {
	c := buildEnvConfig(t, AuthBasic)
	setupTlsFromEnv(t, c)

	runCheck(t, c)
}

func TestConfig_Run_NTLM(t *testing.T) {
	c := buildEnvConfig(t, AuthNTLM)

	runCheck(t, c)
}

func TestConfig_Run_NTLM_WithTls(t *testing.T) {
	c := buildEnvConfig(t, AuthNTLM)
	setupTlsFromEnv(t, c)
	runCheck(t, c)
}

func TestConfig_Run_TLS(t *testing.T) {
	c := buildEnvConfig(t, AuthTLS)
	setupTlsFromEnv(t, c)

	if c.TlsCertPath == "" {
		t.Skip("WINRM_TLS_CERT not set")
	}

	runCheck(t, c)
}

func runCheck(t *testing.T, c *Config) {
	err := c.Validate()
	assert.NoError(t, err)

	err, rc, output := c.Run(DefaultTimeout)
	assert.NoError(t, err)
	assert.Equal(t, 0, rc)
	assert.Contains(t, output, "ConsoleHost")
}

func buildEnvConfig(t *testing.T, auth string) *Config {
	host := os.Getenv("WINRM_HOST")
	if host == "" {
		t.Skip("No env config for WINRM_*")
	}

	c := &Config{
		Host:     host,
		User:     os.Getenv("WINRM_USER"),
		Password: os.Getenv("WINRM_PASSWORD"),
		Command:  "Get-Host",
		AuthType: auth,
	}

	verb := strings.ToUpper(auth)

	if user := os.Getenv("WINRM_" + verb + "_USER"); user != "" {
		c.User = user
	}

	if password := os.Getenv("WINRM_" + verb + "_PASSWORD"); password != "" {
		c.Password = password
	}

	return c
}

func setupTlsFromEnv(t *testing.T, c *Config) {
	if os.Getenv("WINRM_SKIP_TLS") != "" {
		t.Skip("WINRM_SKIP_TLS has been set")
	}

	c.Tls = true
	if os.Getenv("WINRM_INSECURE") != "" {
		c.Insecure = true
	}

	if file := os.Getenv("WINRM_TLS_CA"); file != "" {
		c.TlsCAPath = file
	}

	if file := os.Getenv("WINRM_TLS_CERT"); file != "" {
		c.TlsCertPath = file
	}

	if file := os.Getenv("WINRM_TLS_KEY"); file != "" {
		c.TlsKeyPath = file
	}
}