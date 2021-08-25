.PHONY: all
all: build codesign

.PHONY: codesign
codesign:
	codesign --entitlements vz.entitlements -s - ./machine-driver-vz

.PHONY: build
build:
	go build -o machine-driver-vz ./cmd/machine-driver-vz
