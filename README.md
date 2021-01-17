# JumpGet [![JumpGet](https://circleci.com/gh/lsgrep/jumpget.svg?style=svg)](https://circleci.com/gh/lsgrep/jumpget)

![jumpget](https://raw.githubusercontent.com/lsgrep/jumpget/master/assets/jumpget.gif)
JumpGet client submits download url to the JumpGet server with `ssh` tunnel, then JumpGet server downloads & serves the
file only to the JumpGet client IPs(whitelisted).

## Why?

1. `scp`, `rsync` are slow

2. `wget` vs `JumpGet`
   ![jumpget](https://raw.githubusercontent.com/lsgrep/jumpget/master/assets/jumpget.png)

## Installation

### JumpGet Server

1. You need a VPS instance with fast network. You can get a cheap instance
   from [Linode](https://www.linode.com/?r=ceabf8f0da919a9253a7c5a8757366ad7bbfc30f) (
   referral link), Digital Ocean or AWS Lightsail etc

2. Add your SSH public key to the server(`~/.ssh/authorized_keys`).

3. Setup the JumpGet Server

#### Without TLS

```
docker run --name jumpget -p 3100:3100 \
  -p 127.0.0.1:4100:4100 -d \
  -e JUMPGET_PUBLIC_PORT=3100 -e JUMPGET_LOCAL_PORT=4100 \
  -e JUMPGET_PUBLIC_URL="http://example.com:3100"  lsgrep/jumpget

```

#### JumpGet Server with LetsEncrypt TLS

You can setup TLS with Traefik or Nginx. Checkout `docker-compose.yaml` for Traefik setup.

- replace `postmaster@example.com` with your own email address
- relpace `jumpget.example.com` with your own domain name,

  assuming the DNS has been configured correctly

- `docker-compose up -d`

### JumpGet Client

Checkout [Releases](https://github.com/lsgrep/jumpget/releases) for CircleCi built binaries.

For *nix systems:

```
sudo curl -L https://github.com/lsgrep/jumpget/releases/download/v0.1.30/jumpget_$(uname -s)_amd64 -o /usr/local/bin/jumpget
sudo chmod +x /usr/local/bin/jumpget
```

or just:
`go get -u github.com/lsgrep/jumpget`

## Configuration

#### Server Side Environment Variables

- `JUMPGET_LOCAL_PORT`
- `JUMPGET_PUBLIC_PORT`
- `JUMPGET_PUBLIC_URL`
- `JUMPGET_FILE_RETAIN_DURATION` default value 12, files will be deleted after 12 hours.

#### Client Configuration

`$HOME/.jumpget.yaml` stores client configuration info.

```
# $HOME/.jumpget.yaml
host: "example.com"
user: "example_user"
ssh-port: "2253"

```

`jumpget --help` to detailed information

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
# /etc/sysctl.conf
# you have to have >= 4.15 kernel to use this flag
net.ipv4.tcp_congestion_control=bbr
```

## TODO

- [ ] find a way to detect JumpGet server availability
- [ ] make fetch IP process robust

