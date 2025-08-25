
# AudioDrive

![AudioDrive Icon](image.png)

## Service File

`/etc/systemd/system/audiodrive.service`
```
[Unit]
Description=Podcast Server
After=network.target

[Service]
ExecStart=/usr/local/bin/audiodrive --folder ~/audio
WorkingDirectory=/var/lib/audiodrive
Restart=always

[Install]
WantedBy=multi-user.target
```

## My notes:

hosting on tailscale: