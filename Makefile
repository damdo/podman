all: _gokrazy/extrafiles_arm64.tar _gokrazy/extrafiles_amd64.tar

_gokrazy/extrafiles_amd64.tar:
	mkdir -p _gokrazy/extrafiles_amd64
	curl -fsSL https://github.com/mgoltzsche/podman-static/releases/latest/download/podman-linux-amd64.tar.gz | \
		tar xzv --strip-components=1 -C _gokrazy/extrafiles_amd64/ --exclude README.md
	cd _gokrazy/extrafiles_amd64 && tar cf ../extrafiles_amd64.tar *
	rm -rf _gokrazy/extrafiles_amd64

_gokrazy/extrafiles_arm64.tar:
	mkdir -p _gokrazy/extrafiles_arm64
	curl -fsSL https://github.com/mgoltzsche/podman-static/releases/latest/download/podman-linux-arm64.tar.gz | \
		tar xzv --strip-components=1 -C _gokrazy/extrafiles_arm64/ --exclude README.md
	cd _gokrazy/extrafiles_arm64 && tar cf ../extrafiles_arm64.tar *
	rm -rf _gokrazy/extrafiles_arm64

clean:
	rm -f _gokrazy/extrafiles_*.tar
