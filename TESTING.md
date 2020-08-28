Instructions for testing
========================

This check implements testing via go test, which can be easily run like:

```
go test -v ./...
```

## Integration testing

Some testing requires a real WinRM host to test against, this can be enabled by using environment variables.

On a Unix/Linux shell:

```bash
export WINRM_HOST=win10
export WINRM_USER=administrator
export WINRM_PASSWORD=secret
```

On PowerShell:

```powershell
$Env:WINRM_HOST = "win10"
$Env:WINRM_USER = "administrator"
$Env:WINRM_PASSWORD = "secret"
```

With the vars set, just run the normal test command, and no tests should be skipped.

```
go test -v ./...
```

Other vars have some additional behavior, that can be overridden, set or enabled:

* `WINRM_BASIC_USER` Use a different user for Basic Auth
* `WINRM_BASIC_PASSWORD` Use a different password for Basic Auth
* `WINRM_NTLM_USER` Use a different user for NTLM
* `WINRM_NTLM_PASSWORD` Use a different password for NTLM
