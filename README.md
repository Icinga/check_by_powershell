# check_by_powershell

check_by_powershell is a check plugin for Icinga 2 written in GO. It's build on the go-library [winrm](https://github.com/masterzen/winrm), which can execute remote commands on Windows machines through
the use of WinRM/WinRS.

## Preparing the remote Windows machine for Basic authentication
This project supports only basic authentication for local accounts (domain users are not supported). The remote windows system must be prepared for winrm:

### HTTP
		winrm quickconfig
		winrm set winrm/config/service/Auth '@{Basic="true"}'
		winrm set winrm/config/service '@{AllowUnencrypted="true"}'
		winrm set winrm/config/winrs '@{MaxMemoryPerShellMB="1024"}'

### HTTPS
By default WinRM uses Kerberos for authentication so Windows never sends the password to the system requesting validation.
WinRM HTTPS requires a local computer "Server Authentication" certificate with a CN matching the hostname, that is not expired, revoked, or self-signed to be installed.
On the remote host, a PowerShell prompt, using the __Run as Administrator__ :

       $CertThumbprint = 'cert_thumbprint';

       $certificate = Get-ChildItem -Path cert:\ -Recurse | Where-Object Thumbprint -eq $CertThumbprint;

       if ($null -eq $certificate) {
           throw 'The provided thumbprint was not found in any certificate stores';
       }

       # Allow PS-Remote configuration
       Enable-PSRemoting -SkipNetworkProfileCheck -Force;

       # Disable HTTP transport for PS-Remoting to ensure encryption
       Get-ChildItem WSMan:\Localhost\listener | Where-Object Keys -eq "Transport=HTTP" | Remove-Item -Recurse;

       # Set the HTTPS Transport with our provided Thumbprint for the SSL certificate
       New-Item -Path WSMan:\LocalHost\Listener -Transport HTTPS -Address * -CertificateThumbPrint $CertThumbprint -Force;

       # Set Firewall Rule for allowing communication
       New-NetFirewallRule -DisplayName "Windows Remote Management (HTTPS-In)" -Name "Windows Remote Management (HTTPS-In)" -Profile Any -LocalPort 5986 -Protocol TCP;

       # Enable the HTTPS lisener
       Set-Item WSMan:\localhost\Service\EnableCompatibilityHttpsListener -Value true;

       # Disable possible old HTTP firewall rules
       Disable-NetFirewallRule -DisplayName "Windows Remote Management (HTTP-In)";
       Disable-NetFirewallRule -DisplayName "Windows-Remoteverwaltung (HTTP eingehend)";

       winrm set winrm/config/service/auth '@{Basic="true"}';
       winrm set winrm/config/client '@{TrustedHosts="*"}';

       Restart-Service winrm;

If it's necessary to use a self-signed-certificate, you can follow this [guide](https://www.visualstudiogeeks.com/devops/how-to-configure-winrm-for-https-manually)

## GO build example
To compile for a specific platform, you have to set the GOOS and GOARCH environment variables. For more [information](https://golang.org/pkg/go/build/)

| **Platform**        | **GOOS**           | **GOARCH**  |
| ------------- |:-------------:| -----:|
| Mac      | darwin | amd64 |
| Linux      | linux      |   amd64 |
| Windows | windows      |    amd64 |

### Example - Linux
     GOOS=linux GOARCH=amd64 go build -o check_by_powershell main.go 

## Usage
    ./check_by_powershell -h
    Usage of check_by_powershell

    This Plugin executes remote commands on Windows machines through the use of WinRM.

    Arguments:
      -H, --host string          Host name, IP Address of the remote host (default "127.0.0.1")
      -p, --port int             Port number WinRM (default 5985)
          --user string          Username of the remote host
          --password string      Password of the user
          --tls                  Use TLS connection (default: false)
      -u, --unsecure             Verify the hostname on the returned certificate
          --ca string            CA certificate
          --cert string          Client certificate
          --key string           Client Key
          --cmd string           Command to execute on the remote machine
          --icingacmd string     Executes commands of Icinga PowerShell Framework (e.g. Invoke-IcingaCheckCPU)
          --auth string          Authentication mechanism - NTLM | SSH
          --sshhost string       SSH Host (mandatory if --auth=SSH)
          --sshuser string       SSH Username (mandatory if --auth=SSH)
          --sshpassword string   SSH Password (mandatory if --auth=SSH)
      -t, --timeout int          Abort the check after n seconds (default 10)
      -d, --debug                Enable debug mode
      -v, --verbose              Enable verbose mode
      -V, --version              Print version and exit

### Execute a script over http
    ./check_by_powershell -H 192.168.172.217 -p 5985 --cmd "cscript.exe /T:30 /NoLogo C:\Windows\system32\check_time.vbs 1.de.pool.ntp.org 20 240" --user "windowsuser" --password 'secret!pw'

    OK - NTP OK: Offset +0.0556797 secs|'offset'=+0.0556797s;20;240;

It is necessary that the PowerShell script exits with an exitcode like *exit 2*, otherwise the plugin could exit with an unexpected exitcode.

### Execute a Icinga PowerShell Framework Commandlet over https
    ./check_by_powerhsell  -H "example.local.de" --icingacmd "Invoke-IcingaCheckCPU" --user "windowsuser" --password '!secret!pw'  --tls --unsecure -t 30

    [OK] Check package "CPU Load"
    | 'core_23_10'=2.31%;;;0;100 'core_23_3'=2.54%;;;0;100 'core_23_15'=2.12%;;;0;100 'core_23_5'=2.39%;;;0;100
      'core_23_1'=2.04%;;;0;100 'core_23'=1.93%;;;0;100 'core_2_15'=2.78%;;;0;100 'core_2_10'=2.89%;;;0;100 [...]