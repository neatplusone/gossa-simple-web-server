gossa-sws
=============

![gossasws_screenshot](https://github.com/user-attachments/assets/239fb243-be60-4513-a06b-2a371c8e02a6)

[![build status](https://github.com/neatplusone/gossa-simple-web-server/workflows/ci/badge.svg)](https://github.com/neatplusone/gossa-simple-web-server/actions)
[![github downloads](https://img.shields.io/github/downloads/neatplusone/gossa-simple-web-server/total.svg?logo=github)](https://github.com/neatplusone/gossa-simple-web-server/releases)

a fast and simple webserver for your files, that's dependency-free and with a small, easy-to-review codebase.

a simple UI comes as default, featuring :

  * 🔍 files/directories browser & handler
  * 📩 drag-and-drop uploader
  * 💾 90s web UI that prints in milliseconds
  * 📸 video streaming, picture browser, pdf viewer
  * ✍️ simple note editor
  * ⌨️ keyboard navigation
  * 🚀 lightweight and dependency free codebase
  * 🔒 >95% test coverage and reproducible builds
  * 🥂 fast golang static server
  * 💑 easy multi account setup, read-only mode
  * ✨ PWA-able
  * 🖥️ multi-platform support
  * &plus; easy simple API for automation, see [See CURL/Powershell example](https://github.com/neatplusone/gossa-simple-web-server/blob/gossa-sws/API.md)
  * &plus; folder size calculation (optional)
  * &plus; sorting by name, file modification date, file size 
  * &plus; multi-select batch delete
  * &plus; optional built-in HTTPS and basic auth
  * &plus; optional per-file upload size limit
  * &plus; ui/ux improvements

### install / build
<!-- [arch linux (AUR)](https://aur.archlinux.org/packages/gossa/) - e.g. `yay -S gossa` 

[nix](https://search.nixos.org/packages?channel=unstable&show=gossa&from=0&size=50&sort=relevance&type=packages&query=gossa) - e.g. `nix-shell -p gossa`

[mpr](https://mpr.makedeb.org/packages/gossa) -->

binaries are available on the [release page](https://github.com/neatplusone/gossa-simple-web-server/releases)

all builds are reproducible, checkout the hashes on the release page.

### usage
```sh
% ./gossa --help

% ./gossa -h 192.168.100.33 ~/storage
```

### options
| flag | default | description |
| --- | --- | --- |
| `-h` | `0.0.0.0` | host to listen on |
| `-p` | `8000` | port to listen on |
| `-prefix` | `/` | url prefix at which gossa is reached, e.g. `/gossa/` |
| `-symlinks` | `false` | follow symlinks (allows escaping the served path) |
| `-k` | `true` | skip hidden files (dot-prefixed) |
| `-ro` | `false` | read-only mode (no upload, rename, move, delete) |
| `-calcfoldersize` | `false` | calculate and display folder sizes |
| `-verb` | `false` | verbose logging |
| `-maxupload` | `0` | max upload size in MB per file (`0` = unlimited) |
| `-tls-cert` | `""` | path to TLS certificate to serve HTTPS (with `-tls-key`) |
| `-tls-key` | `""` | path to TLS key to serve HTTPS (with `-tls-cert`) |
| `-auth` | `""` | enable HTTP basic auth, format `user:pass` |

```sh
# serve HTTPS with basic auth and a 500MB per-file upload cap
% ./gossa -tls-cert cert.pem -tls-key key.pem -auth alice:s3cret -maxupload 500 ~/storage
```

### shortcuts
press `Ctrl/Cmd + h` to see all the UI/keyboard shortcuts.

to delete several items at once, tick the checkboxes next to them and click the "delete selected" button.

### fancier setups
<!--
release images are pushed to [ghcr](https://github.com/neatplusone/gossa-simple-web-server/pkgs/container/gossa-simple-web-server), e.g. :

```sh
# pull from dockerhub and run
% mkdir ~/LocalDirToShare
% sudo docker run -v ~/LocalDirToShare:/shared -p 8000:8000 neatplusone/gossa-simple-web-server
```
 -->

for quick setups, HTTPS (`-tls-cert`/`-tls-key`) and basic auth (`-auth`) are available built-in. for richer needs (multi-user, rate limiting, etc.) they can still be delegated to a reverse proxy - [sample caddy configs](https://github.com/neatplusone/gossa-simple-web-server/blob/master/support/) are available to quickly setup multi users setups along with https.

automatic boot-time startup can be handled with a user systemd service - see [support](https://github.com/neatplusone/gossa-simple-web-server/tree/master/support)

