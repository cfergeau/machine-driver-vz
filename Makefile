.PHONY: all
all: build codesign

.PHONY: codesign
codesign:
	codesign --entitlements vf.entitlements -s - ./machine-driver-vf

.PHONY: build
build:
	go build -o machine-driver-vf ./cmd/machine-driver-vf
