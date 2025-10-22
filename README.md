# ipupdater
IP updater for dynamically changing IPs

Supporting currently PorkbunDNS

## Config

### YAML


### ENV


This shows the API keys in ps -ax and closes program on disconnect

```bash
DNS_APIKEY="apikey" DNS_APISECRET="secret" ipupdater &
```

Better solution and won't close on disconnect

```bash
export DNS_APIKEY="apikey" DNS_APISECRET="secret"
nohup ipupdater &
```


Add systemd service

```bash
# /etc/systemd/system/ipupdater.service
[Unit]
Description=IP Updater
After=network.target

[Service]
Type=simple
User=youruser
WorkingDirectory=/path/to/ipupdater
Environment=DNS_APIKEY=apikey
Environment=DNS_APISECRET=secret
ExecStart=/path/to/ipupdater
Restart=always

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl start ipupdater
sudo systemctl enable ipupdater
```

Use Docker

```bash
export DNS_APIKEY="apikey" DNS_APISECRET="secret"
docker pull 0x6a4b.dev/ipupdater:latest
docker run -d --rm -e DNS_APIKEY -e DNS_APISECRET --name ipupdater 0x6a4b.dev/ipupdater
```

Docker with .env

.env file contents

```bash
DNS_APIKEY=apikey
DNS_APISECRET=secretapikey
```

```bash
docker run -d --rm --env-file .env --name ipupdater 0x6a4b.dev/ipupdater
```
