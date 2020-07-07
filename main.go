package main

import (
	"bytes"
	"fmt"
	"github.com/NETWAYS/go-check"
	"github.com/masterzen/winrm"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"os"
	"time"
)

func main() {

	defer check.CatchPanic()

	config := check.NewConfig()

	config.Name = "check_by_powershell"
	config.Readme = `This Plugin executes remote commands on Windows machines through the use of WinRM.`
	config.Version = "1.0.0"
	config.Timeout = 10

	// WinRM / Powershell
	host := config.FlagSet.StringP("host", "H", "127.0.0.1", "Host name, IP Address of the remote host")
	port := config.FlagSet.IntP("port", "p", 5985, "Port number WinRM")

	user := config.FlagSet.String("user", "", "Username of the remote host")
	password := config.FlagSet.String("password", "", "Password of the user")

	tls := config.FlagSet.Bool("tls", false, "Use TLS connection (default: false)")
	unsecure := config.FlagSet.BoolP("unsecure", "u", false, "Verify the hostname on the returned certificate")

	ca := config.FlagSet.String("ca", "", "CA certificate")
	cert := config.FlagSet.String("cert", "", "Client certificate")
	key := config.FlagSet.String("key", "", "Client Key")

	command := config.FlagSet.String("cmd", "", "Command to execute on the remote machine")
	icingacmd := config.FlagSet.String("icingacmd", "", "Executes commands of Icinga PowerShell Framework (e.g. Invoke-IcingaCheckCPU)")

	auth := config.FlagSet.String("auth", "", "Authentication mechanism - NTLM | SSH")
	sshHost := config.FlagSet.String("sshhost", "", "SSH Host (mandatory if --auth=SSH)")
	sshUser := config.FlagSet.String("sshuser", "", "SSH Username (mandatory if --auth=SSH)")
	sshPassword := config.FlagSet.String("sshpassword", "", "SSH Password (mandatory if --auth=SSH)")

	config.ParseArguments()

	if *command == "" && *icingacmd == "" {
		check.Exit(3, "PowerShell command is empty. Please enter the command to execute.")
	}

	var (
		clientCert []byte
		clientKey  []byte
		clientCa   []byte
		cmd        string
		debugCmd   string
		buf        bytes.Buffer
		errmsg     bytes.Buffer
		exitcode   int
		err        error
	)

	if *tls {
		*port = 5986
	}

	if *tls && *unsecure == false {
		if *cert != "" {
			clientCert, err = ioutil.ReadFile(*cert)
			if err != nil {
				check.Exit(3, "failed to read client certificate: %q", err)
				// panic(err)
			}
		} else {
			check.Exit(3, "please specify a client certificate")
			// panic(err)
		}

		if *key != "" {
			clientKey, err = ioutil.ReadFile(*key)
			if err != nil {
				check.Exit(3, "failed to read client key: %q", err)
				// panic(err)
			}
		} else {
			check.Exit(3, "please specify a client key")
			// panic(err)
		}

		if *ca != "" {
			clientCa, err = ioutil.ReadFile(*ca)
			if err != nil {
				check.Exit(3, "failed to read client ca: %q", err)
				// panic(err)
			}
		}

		winrm.DefaultParameters.TransportDecorator = func() winrm.Transporter {
			return &winrm.ClientAuthRequest{}
		}
	}

	endpoint := winrm.NewEndpoint(
		*host,      // Host to connect to
		*port,      // Winrm port
		*tls,       // Use TLS
		*unsecure,  // Allow insecure connection
		clientCa,   // CA certificate
		clientCert, // Client Certificate
		clientKey,  // Client Key
		time.Duration(config.Timeout)*time.Second, // Timeout
	)
	params := winrm.DefaultParameters

	if *auth == "NTML" {
		params.TransportDecorator = func() winrm.Transporter { return &winrm.ClientNTLM{} }
	}

	if *auth == "SSH" {
		if *sshUser == "" || *sshPassword == "" || *sshHost == "" {
			check.Exit(3, "Please specify a sshuser, sshpassword, sshhost")
		} else {
			sshClient, err := ssh.Dial("tcp", *sshHost+":22", &ssh.ClientConfig{
				User:            *sshUser,
				Auth:            []ssh.AuthMethod{ssh.Password(*sshPassword)},
				HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			})
			params.Dial = sshClient.Dial

			if err != nil {
				check.Exit(3, "%q", err)
				// panic(err)
			}
		}
	}

	client, err := winrm.NewClientWithParameters(endpoint, *user, *password, params)
	if err != nil {
		check.Exit(3, "%q", err)
		// panic(err)
	}

	if *icingacmd != "" && *command != "" {
		check.Exit(3, "icingacmd and command are mutually exclusive")
	}

	if *icingacmd != "" {
		cmd = winrm.Powershell("try { Use-Icinga; exit (" + *icingacmd + ") } catch { Write-Host ('UNKNOWN: ' + $error); exit 3 }")
		debugCmd = "try { Use-Icinga; exit (" + *icingacmd + ") } catch { Write-Host ('UNKNOWN: ' + $error); exit 3 }"
	} else if *command != "" {
		cmd = winrm.Powershell("try { " + *command + "; exit $LASTEXITCODE } catch { Write-Host ('UNKNOWN: ' + $error); exit 3 }")
		debugCmd = "try { " + *command + "; exit $LASTEXITCODE } catch { Write-Host ('UNKNOWN: ' + $error); exit 3 }"
	}

	if config.Debug {
		fmt.Println("Executed PowerShell command:\n" + debugCmd + "\n")
	}

	exitcode, err = client.Run(cmd, &buf, &errmsg)
	if err != nil {
		check.Exit(3, "%q", err)
		// panic(err)
	}

	fmt.Print(buf.String())
	os.Exit(exitcode)
}
