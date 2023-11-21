# Makefile

# Default target
all: build push

VERSION ?= v0.1.5

# Build the package
build:
	crossplane xpkg build --ignore=kustomization.yaml -f . -o crossplane-minecraft.xpkg

# Push the package
push:
	crossplane xpkg push -f crossplane-minecraft.xpkg index.docker.io/dirien/crossplane-minecraft:$(VERSION)
