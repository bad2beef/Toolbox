# BITS

Part of BAD2BEEF's Toolbox

## Overview

This is a minimal BITS upload handler suitable for small one-off or containerized deployments. The goal is to be fairly small and simple, allowing use without the need to stand up a Windows Server, IIS, or any full-fledged web server with language runtimes.

## Building

    go build bits.go

## Usage

    bits [options]

### Options

- **-listen** *:8080* Golang net/http listener string
- **-route** */bits* URL route for BITS handler (Leading / required)
- **-cert** Certificate chain if TLS is desired
- **-key** Certificate's private key
- **-bits** *bits* Directory to store BITS data
- **-logs** *logs* Directory to store logs

### Examples

#### Start BITS Handler

    $ ./bits -listen :8080 -route /bits -bits bits -logs logs
    0000-00-00T00:00:00-00:00 :8080 - Listening on :8080/

#### Perform Upload

Perform a BITS upload transfer from PowerShell on a Windows host.

    PS> Start-BitsTransfer -TransferType Upload -Source .\some-file.ext -Destination http://my-bits-host.tld:8080/bits

## References

1. [BITS Upload Protocol](https://docs.microsoft.com/en-us/windows/win32/bits/bits-upload-protocol)
