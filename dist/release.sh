#!/bin/sh
# Builds release tarballs with prebuilt binaries. Used by `make release`.
# Usage: release.sh VERSION OUTDIR ARCH...
set -eu

version=$1 out=$2
shift 2

for arch in "$@"; do
	case $arch in
	x86_64) goarch=amd64 ;;
	aarch64) goarch=arm64 ;;
	*) echo "unknown arch: $arch" >&2; exit 1 ;;
	esac

	name=bannin-$version-linux-$arch
	dir=$out/$name
	mkdir -p "$dir/bin" "$dir/plasmoid"

	# Static binaries: no C libraries to depend on.
	for cmd in bannin bannind; do
		CGO_ENABLED=0 GOOS=linux GOARCH=$goarch go build -trimpath \
			-ldflags "-s -w -X main.version=$version" -o "$dir/bin/$cmd" "./cmd/$cmd"
	done

	cp -r ui/plasma/. "$dir/plasmoid/"
	cp dist/bannind.service LICENSE README.md "$dir/"
	install -m 755 dist/install.sh dist/uninstall.sh "$dir/"

	# The binaries include these libraries, so their licenses travel with them.
	{
		printf '== Go standard library and runtime ==\n\n'
		cat "$(go env GOROOT)/LICENSE"
		GOOS=linux GOARCH=$goarch go list -deps \
			-f '{{with .Module}}{{if not .Main}}{{.Path}} {{.Version}} {{.Dir}}{{end}}{{end}}' ./cmd/... |
			sort -u |
			while read -r mod ver moddir; do
				[ -n "$mod" ] || continue
				printf '\n== %s %s ==\n\n' "$mod" "$ver"
				found=
				for file in "$moddir"/LICENSE* "$moddir"/COPYING*; do
					if [ -f "$file" ]; then cat "$file" && found=1; fi
				done
				[ -n "$found" ] || { echo "no license file found for $mod" >&2; exit 1; }
			done
	} >"$dir/THIRD_PARTY_LICENSES"

	tar -C "$out" --owner=0 --group=0 --numeric-owner -czf "$out/$name.tar.gz" "$name"
	rm -rf "$dir"
done

(cd "$out" && sha256sum ./*.tar.gz >SHA256SUMS)
