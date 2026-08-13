An!KOTO
=======

a tui for browsing and streaming anime, scraping from **anidb.app**.
navigate with vim keys or arrows, and play in your preferred video player.

> this project is a rebranded fork of [anitui](https://github.com/typechecks/anitui)
> with a fresh source (anidb.app) and some home-screen polish.

features
--------

- search anidb.app and browse results without a browser
- sub / dub toggle per episode
- multiple video sources (1080p / 720p / 360p) with automatic selection
- mouse support: click the **made by t4fs** button under the search bar
  to open the author's github
- plays in mpv, vlc, iina, or haruna

installation
------------

### windows

1. build and install:

   ```powershell
   go build -trimpath -o "$HOME\.anitui\anitui.exe" ./cmd/anitui
   ```

2. (optional) make `anikoto` available in any terminal by adding a shim
   directory to your `PATH`:

   ```powershell
   New-Item -ItemType Directory -Force -Path "$HOME\scoop\shims" | Out-Null
   Set-Content "$HOME\scoop\shims\anikoto.cmd" '@"%USERPROFILE%\.anitui\anitui.exe" %*'
   ```

   then run `anikoto` from cmd, powershell, or git bash.

### linux / macos

```bash
make build
./build/anitui   # or: sudo ./build/anitui install
```

### from source

```bash
git clone https://github.com/T4fs/anikoto.git
cd anikoto
make build
```

quick start
-----------

```bash
# build
make build

# run
./build/anitui

# or via Go
go run ./cmd/anitui
```

controls
--------

### navigation

| key | action |
|-----|--------|
| enter | confirm / select |
| esc | return |
| j / ↓ | down |
| k / ↑ | up |
| g g | jump to top |
| G | jump to bottom |
| ctrl+u / ⌘u | page up |
| ctrl+d / ⌘d | page down |
| / | search |
| ? | toggle help popup |
| ctrl+c / ⌘c | exit |

### episode screen

| key | action |
|-----|--------|
| j / ↓ | down |
| k / ↑ | up |
| enter | play episode |
| space | toggle synopsis expand |
| d | toggle sub / dub |
| esc | back to results |

### watching screen

| key | action |
|-----|--------|
| h / ← | previous episode |
| l / → | next episode |
| r | replay current episode |
| space | replay current episode |
| s | cycle video source |
| d | toggle sub / dub |
| esc | back to episode list |

player support
--------------

anikoto auto-detects and uses:

1. mpv
2. iina (macOS)
3. vlc
4. haruna

override via `ANITUI_PLAYER` environment variable.

platform support
----------------

| os | architectures |
|----|---------------|
| linux | amd64, arm64 |
| macos | amd64, arm64 |
| windows | amd64, arm64 |

building from source
--------------------

```bash
git clone https://github.com/T4fs/anikoto.git
cd anikoto
make build
```

cross-compile:

```bash
make build-linux-amd64    # linux x86_64
make build-linux-arm64    # linux arm64
make build-windows-amd64  # windows x86_64
make build-all            # build for all supported platforms
```

license
-------

this project is licensed under the GNU General Public License v3.0 —
see [LICENSE](LICENSE). original anitui by [typechecks](https://github.com/typechecks).
