# POST

Part of BAD2BEEF's Toolbox

## Overview

A simple HTTP POST handler (though other verbs are fine, too) which dumps request data to disk. If values for redirect, redir, uri or url are included the client will be redirected to the contents of the variable.

## Building

    go build post.go

## Usage

    post [options]

### Options

- **-listen** *:8080* Golang net/http listener string
- **-route** */* URL route for BITS handler (Leading / required)
- **-cert** Certificate chain if TLS is desired
- **-key** Certificate's private key

### Examples

#### Start POST Handler

    $ ./post -listen :8080 -route /
    0000-00-00T00:00:00-00:00 Listening on :8080/
