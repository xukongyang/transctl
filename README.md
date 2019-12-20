# transctl

`transctl` is a command-line utility for controlling remote [Transmission
RPC][transmission-rpc] hosts.

[Installing][] | [Building][] | [Using][] | [Features][] | [Releases][]

[Installing]: #installing (Installing)
[Building]: #building (Building)
[Using]: #using (Using)
[Features]: #features-and-compatibility (Features and Compatibility)
[Releases]: https://github.com/kenshaw/transctl/releases (Releases)

## Overview

`transctl` provides a modern command-line tool for controlling remote
[Transmission RPC][transmission-rpc] hosts. `transctl` is inspired by Git,
`kubectl` and other command-line tools for managing and working with multiple
remote host contexts. `transctl` wraps the simple [`transrpc` client
package][transrpc].

## Installing

`transctl` can be installed [via Release][], [via Homebrew][], [via Scoop][] or [via Go][]:

[via Release]: #installing-via-release
[via Homebrew]: #installing-via-homebrew-macos
[via Scoop]: #installing-via-scoop-windows
[via Go]: #installing-via-go

### Installing via Release

1. [Download a release for your platform][Releases]
2. Extract the `transctl` or `transctl.exe` file from the `.tar.bz2` or `.zip` file
3. Move the extracted executable to somewhere on your `$PATH` (Linux/macOS) or
`%PATH%` (Windows)

### Installing via Homebrew (macOS)

`transctl` is available in the [`kenshaw/kenshaw` tap][kenshaw-tap], and can be installed in the
usual way with the [`brew` command][homebrew]:

```sh
# add tap
$ brew tap kenshaw/kenshaw

# install transctl with "most" drivers
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

[transmission-rpc]: https://github.com/transmission/transmission/blob/master/extras/rpc-spec.txt
[transrpc]: https://github.com/kenshaw/transrpc
[kenshaw-tap]: https://github.com/kenshaw/homebrew-kenshaw
