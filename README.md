# ipupdater
IP updater for dynamically changing IPs

Supporting currently PorkbunDNS


> NOTICE! Currently being changed to use apikeys from ENV and to be modular with DNS APIs, probably with dynamic libraries, and does not work before modifications! Will be fixed shortly!


> Current situation bit too busy and on hold:
> - Add modules for different DNS service APIs
> - Maybe using dynamically linking libraries, which would mean adding new services need only an added .so library file
> - Refactoring code to be modular and designing required exported functions from libraries, e.g. update and about, perhaps



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

Docker pull

```bash
docker pull 0x6a4b.dev/ipupdater:latest
```

Docker build

```bash
docker build -t ipupdater .
```

Run with API key as environment variable

```bash
export DNS_APIKEY="apikey" DNS_APISECRET="secret"
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
