# Maintainer: Erik Martino <erik.martino@gmail.com>
pkgname=ls3-git
_pkgname=ls3
# This will be auto-updated by the pkgver function
pkgver=0.0.1
pkgrel=1
pkgdesc="A simple terminal-based user interface for browsing Amazon S3 buckets and objects."
arch=('x86_64' 'amd64' 'arm64')
url="https://github.com/erikmartino/ls3"
license=('GPL2')
depends=()
makedepends=('go')
source=("git+https://github.com/erikmartino/ls3.git")
sha256sums=('SKIP')

pkgver() {
  cd "$srcdir/$_pkgname"
  git describe --long --tags | sed 's/\([^-]*-g\)/r\1/;s/-/./g'
}

build() {
  cd "$srcdir/$_pkgname"
  go build
}

package() {
  install -Dm755 "$srcdir/$_pkgname/$_pkgname" "$pkgdir/usr/bin/$_pkgname"
  install -Dm644 "$srcdir/$_pkgname/LICENSE" "$pkgdir/usr/share/licenses/$_pkgname/LICENSE"
}
