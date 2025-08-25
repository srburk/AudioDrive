
# AudioDrive

## Service File

```
[Unit]
Description=Podcast Server
After=network.target

[Service]
ExecStart=/usr/local/bin/audiodrive-server --folder ~/audio
Restart=always

[Install]
WantedBy=multi-user.target
```

## My notes:

hosting on tailscale: