# transctl

`transctl` is a command-line utility for controlling remote Torrent clients
([Transmission][transmission], [qBittorrent][qbittorrent], [Deluge][deluge],
[rTorrent][rtorrent]).

[Installing][] | [Building][] | [Using][] | [Features][] | [Releases][]

[Installing]: #installing (Installing)
[Building]: #building (Building)
[Using]: #using (Using)
[Features]: #features-and-compatibility (Features and Compatibility)
[Releases]: https://github.com/kenshaw/transctl/releases (Releases)

## Overview

`transctl` provides a unified command-line tool for controlling remote torrent
clients.  Inspired by Git, `kubectl`, and other modern command-line tools,
`transctl` provides a standard and easy way to manage torrents from the
command-line.  The `transctl` project also provides Go client packages for each
of the clients supported, in a standard Go-idiomatic fashion.

## Installing

`transctl` can be installed [via Release][], [via Homebrew][], [via Scoop][] or [via Go][]:

[via Release]: #installing-via-release
[via Homebrew]: #installing-via-homebrew-macos
[via Scoop]: #installing-via-scoop-windows
[via Go]: #installing-via-go

### Installing via Release

1. [Download a release for your platform][Releases]
2. Extract the `transctl` or `transctl.exe` file from the `.tar.bz2` or `.zip` file
3. Move the extracted executable to somewhere on your `$PATH` (macOS/Linux) or
`%PATH%` (Windows)

### Installing via Homebrew (macOS)

`transctl` is available in the [`kenshaw/kenshaw` tap][kenshaw-tap], and can be installed in the
usual way with the [`brew` command][homebrew]:

```sh
# add tap
$ brew tap kenshaw/kenshaw

# install transctl
$ brew install transctl
```

### Installing via Scoop (Windows)

`transctl` can be installed using [Scoop](https://scoop.sh):

```powershell
# install scoop if not already installed
iex (new-object net.webclient).downloadstring('https://get.scoop.sh')

scoop install transctl
```

### Installing via Go

`transctl` can be installed in the usual Go fashion:

```sh
# install transctl
$ go get -u github.com/kenshaw/transctl
```

[deluge]: https://www.deluge-torrent.org/
[homebrew]: https://brew.sh/
[kenshaw-tap]: https://github.com/kenshaw/homebrew-kenshaw
[qbittorrent]: https://www.qbittorrent.org/
[rtorrent]: https://rakshasa.github.io/rtorrent/
[transmission]: https://transmissionbt.com/
