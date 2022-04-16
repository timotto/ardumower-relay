# ArduMower Sunray Relay Server

The ArduMower Relay Server exposes 
an [ArduMower Modem](https://github.com/timotto/ardumower-modem)
or vanilla [Sunray esp32_ble](https://github.com/Ardumower/Sunray/tree/master/esp32_ble) Sketch 
on the Internet.
It acts as a bridge 
between the ArduMower Sunray App and your ArduMower, 
so you can

- reach your ArduMower from the Internet without port forwarding
- connect securely with TLS
- forget the ArduMower WiFi IP address
- keep all your browser settings on secure defaults

You can either run your own server with or without authentication,
or you can apply for credentials to use an already existing server (*).

_(*) currently there are no public servers available_

## Running your own Relay Server

### HTTP vs HTTPS

When you run the Relay Server without any arguments it will listen on port `8080` in single user mode, without any authentication or authorization.

To fulfill all the claims above you need to expose the Relay Server to the Internet at a publicly trusted HTTPS address.

You can run the Relay Server behind a reverse proxy like [Caddy server](https://caddyserver.com/) which takes care of Lets' Encrypt certificates.

### Authentication

#### No Authentication

By default, the server runs without authentication. This allows only a single ArduMower per server.

#### Static credentials

Store username and password combinations in a text file. Only those credentials will allow access to the server.
The [example credentials file](docs/example/users.example.plaintext) is used in the [exanple configuration file](docs/example/config.example.yml).

#### Free For All

The server allows any credentials.
As long as you use the same credentials in the ArduMower Modem and in the Sunray App, you can control your ArduMower. 
Add this to your `config.yml` to enable Free For All:

```yaml
auth:
  enabled: false
  free_for_all: true
```

### Using the binary

Executable binaries 
of the Relay Server 
are available for download 
on the [GitHub release page](https://github.com/timotto/ardumower-relay/releases). 

### Using Docker

The executable binaries 
from the [GitHub release page](https://github.com/timotto/ardumower-relay/releases)
are also available as Docker image.
The exact Docker repository is listed in the release notes.

There is no `latest` tag, and there are no other non-immutable tags either.
You currently need to specify the exact version.
I'm planning on publishing non-immutable tags for major and minor semantic version aliases.

### Running from source

To test, build, and run the Relay Server you need [Go](https://go.dev/) 1.17 or later.

## License

Copyright (c) 2021 Tim Otto

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR
OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE
OR OTHER DEALINGS IN THE SOFTWARE.