# JumpGet [![JumpGet](https://circleci.com/gh/lsgrep/jumpget.svg?style=svg)](https://circleci.com/gh/lsgrep/jumpget)

## Why?

![demo](https://raw.githubusercontent.com/lsgrep/jumpget/master/assets/lulu.png)

## Installation

1. You need a JumpGet server. You can get a cheap VPS instance from Linode, Digital Ocean or AWS Lightsail
2. Add your SSH public key to the server(`~/.ssh/authorized_keys`). The download tasks are submitted through the `SSH`
   connection.
3. Setup the JumpGet Server

#### Without TLS

```
docker run --name jumpget -p 3100:3100 \
  -p 127.0.0.1:4100:4100 -d \
  -e JUPMGET_PUBLIC_PORT=3100 -e JUMPGET_LOCAL_PORT=4100 \
  -e JUMPGET_PUBLIC_URL="example.com:3100"  lsgrep/jumpget

```

#### JumpGet Server with TLS

You can setup TLS with Traefik or Nginx. Checkout `docker-compose.yaml` for Traefik setup.

- replace `postmaster@example.com` with your own email address
- relpace `jumpget.example.com` with your own domain name,

  assuming the DNS has been configured correctly

- `docker-compose up -d`

4. Install JumpGet client, Checkout [Releases](https://github.com/lsgrep/jumpget/releases)

For *nix systems:

```
sudo curl -L https://github.com/lsgrep/jumpget/releases/download/v0.1.8/jumpget_$(uname -s)_amd64 -o /usr/local/bin/jumpget
sudo chmod +x /usr/local/bin/jumpget
```

## Configuration

#### Server Side Environment Variables

- `JUMPGET_LOCAL_PORT`
- `JUMPGET_PUBLIC_PORT`
- `JUMPGET_PUBLIC_URL`
- `JUMPGET_FILE_RETAIN_DURATION` default value 12, files will be deleted after 12 hours.

#### Client Configuration

`$HOME/.jumpget.yaml` stores client configuration info. `jumpget --help` to detailed information

- `JUMPGET_LOCAL_PORT` default value 4100, this value should be consistent with server

## Security

- Download tasks are submitted through `ssh`
- Public File Server can only accessed through whitelisted IPs
- whitelisted IPs are fetched through:
    - https://api.ipify.org
    - $SSH_CONNECTION on JumpGet server

## Performance

1. Enable better TCP congestion control method on JumpGet server

```
# bbr
net.ipv4.tcp_congestion_control=bbr
```

