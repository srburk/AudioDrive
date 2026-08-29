
# AudioDrive

Serves a folder of audio files as a private podcast feed. A podcast client
subscribes to one unguessable HTTPS URL and plays episodes from it.

Anyone who has that URL can fetch the feed, the cover image, and every
episode. Treat it like a password. Do not post it, commit it, or share it
beyond the people who should subscribe.

## Subscribe

Start the server. On startup it prints the local subscribe URL:

```
Subscribe URL: http://127.0.0.1:8080/<token>/rss.xml
```

Podcast apps need HTTPS. Behind Tailscale Funnel (or another reverse proxy)
the URL is:

```
https://<your-host>/<token>/rss.xml
```

Paste that into Overcast, Apple Podcasts, Pocket Casts, AntennaPod, or any
client that accepts a feed URL. The same token is on enclosure and image
links in the feed, so clients can fetch audio without cookies or a login.

`/rss.xml`, `/image.png`, and `/filename.mp3` without the token are denied.
The audio folder is not listed as a directory.

## Token

The token is a 32-character secret stored in a file so the subscribe URL stays
the same across restarts.

- **Default path:** `token` in the working directory (mode 0600). The
  systemd unit below uses `WorkingDirectory=/var/lib/audiodrive`, so the
  file is `/var/lib/audiodrive/token`.
- **First start:** if the file is missing, AudioDrive generates a token and
  writes it there.
- **Set or rotate:** write a new token into that file, or pass
  `-token <value>` (it is saved to the token file). Restart. Update every
  podcast client. The old URL stops working.
- **Custom path:** `-token-file /path/to/token`.

## Flags

```
-p, -port        listen port (default 8080)
-folder          audio directory (default ./audio)
-image           cover image (default ./image.png)
-token-file      token file path (default token)
-token           set the subscribe token and save it to token-file
```

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

The subscribe URL is in the process stdout on start. `journalctl -u audiodrive`
shows it.

## TODO

* Manage uploads on a dashboard
* Edit upload info (title, etc.)
