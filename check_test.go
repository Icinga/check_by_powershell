package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

const DefaultTimeout = 15 * time.Second

func TestConfig_Validate(t *testing.T) {
	c := &Config{}
	assert.Error(t, c.Validate())

	// Most basic settings
	c.Host = "localhost"
	c.Command = "Get-Something"
	c.User = "administrator"
	c.Password = "verysecret"

	assert.NoError(t, c.Validate())
	assert.Equal(t, c.Port, TlsPort)
	assert.False(t, c.NoTls)
	assert.Equal(t, c.AuthType, AuthDefault)
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
	assert.Contains(t, c.BuildCommand(), "powershell.exe -EncodedCommand")

	c = &Config{IcingaCommand: "Icinga-CheckSomething"}
	assert.Contains(t, c.BuildCommand(), "powershell.exe -EncodedCommand")
}

func TestConfig_Run_WithError(t *testing.T) {
	c := &Config{
		Host:     "192.0.2.11",
		User:     "admin",
		Password: "test",
		Command:  "Get-Host",
		NoTls:    true,
	}

	err := c.Validate()
	assert.NoError(t, err)

	err, _, _ = c.Run(1 * time.Second)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dial tcp 192.0.2.11:")
}

func TestConfig_Run_Basic(t *testing.T) {
	if os.Getenv("WINRM_SKIP_BASIC") != "" {
		t.Skip("WINRM_SKIP_BASIC has been set")
	}

	if os.Getenv("WINRM_SKIP_UNENCRYPTED") != "" {
		t.Skip("WINRM_SKIP_UNENCRYPTED has been set")
	}

	c := buildEnvConfig(t, AuthBasic)
	c.NoTls = true

	fmt.Printf("%v\n", c)

	runCheck(t, c)
}

func TestConfig_Run_Basic_WithTLS(t *testing.T) {
	if os.Getenv("WINRM_SKIP_BASIC") != "" {
		t.Skip("WINRM_SKIP_BASIC has been set")
	}

	c := buildEnvConfig(t, AuthBasic)
	setupTlsFromEnv(t, c)

	err := c.Validate()
	assert.NoError(t, err)

	fmt.Printf("%v\n", c)

	runCheck(t, c)
}

func TestConfig_Run_NTLM(t *testing.T) {
	if os.Getenv("WINRM_SKIP_UNENCRYPTED") != "" {
		t.Skip("WINRM_SKIP_UNENCRYPTED has been set")
	}

	c := buildEnvConfig(t, AuthNTLM)
	c.NoTls = true

	err := c.Validate()
	assert.NoError(t, err)

	fmt.Printf("%v\n", c)

	runCheck(t, c)
}

func TestConfig_Run_NTLM_WithTls(t *testing.T) {
	c := buildEnvConfig(t, AuthNTLM)
	setupTlsFromEnv(t, c)

	err := c.Validate()
	assert.NoError(t, err)

	fmt.Printf("%v\n", c)

	runCheck(t, c)
}

func TestConfig_Run_TLS(t *testing.T) {
	c := buildEnvConfig(t, AuthTLS)
	setupTlsFromEnv(t, c)

	if c.TlsCertPath == "" {
		t.Skip("WINRM_TLS_CERT not set")
	}

	err := c.Validate()
	assert.NoError(t, err)

	fmt.Printf("%v\n", c)

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

	if file := os.Getenv("WINRM_TLS_PORT"); file != "" {
		tmp, err := strconv.ParseInt(file, 10, 16)
		assert.NoError(t, err)

		c.Port = int(tmp)
	}
}
