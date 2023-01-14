# Maintainer: TeamSUEP <Arch is the best!>
pkgname=getmycourses-git
_pkgname=getMyCourses
pkgver=1.5.r1.d838f33
pkgrel=1
pkgdesc="从SUEP教务系统获取自己的课程表，并生成可导入日历的.ics文件。"
arch=('x86_64' 'aarch64')
url="https://github.com/TeamSUEP/getMyCourses"
license=('MIT')
groups=()
depends=('glibc')
makedepends=('git' 'go')
provides=("${pkgname%-git}")
conflicts=("${pkgname%-git}")
replaces=()
backup=()
options=()
install=
source=("$_pkgname::git+$url.git")
noextract=()
md5sums=('SKIP')

pkgver() {
    cd "$srcdir/$_pkgname"
    printf "%s" "$(git describe --long --tags | sed 's/\([^-]*-\)g/r\1/;s/-/./g')"
}

prepare() {
    cd "$srcdir/$_pkgname"
    mkdir -p build/
}

build() {
    cd "$srcdir/$_pkgname"
    export CGO_CPPFLAGS="${CPPFLAGS}"
    export CGO_CFLAGS="${CFLAGS}"
    export CGO_CXXFLAGS="${CXXFLAGS}"
    export CGO_LDFLAGS="${LDFLAGS}"
    export GOFLAGS="-buildmode=pie -trimpath -ldflags=-linkmode=external -mod=readonly -modcacherw"
    go build -o build .
}

check() {
    cd "$srcdir/$_pkgname"
    #go test ./...
}

package() {
    cd "$srcdir/$_pkgname"
    install -Dm755 build/$_pkgname "$pkgdir/usr/bin/${pkgname%-git}"
    install -Dm644 -t "$pkgdir/usr/share/licenses/$pkgname" LICENSE
    install -Dm644 -t "$pkgdir/usr/share/doc/$pkgname"      *.md
}
