IMAGE=thzpub/suntrigger
VERSION=latest

run: build-amd64
	[ -e ./env ] && . ./env ; ./suntrigger-amd64


ARCHS:=amd64 arm32v7 arm64v8


build: $(ARCHS:%=build-%)
push: $(ARCHS:%=push-%)

build-amd64:
	CGO_ENABLED=0 GOARCH=$(@:build-%=%) go build -o suntrigger-$(@:build-%=%)
build-arm32v7:
	CGO_ENABLED=0 GOARCH=arm go build -o suntrigger-$(@:build-%=%)
build-arm64v8:
	CGO_ENABLED=0 GOARCH=arm64 go build -o suntrigger-$(@:build-%=%)

push-%: dockerbuild-%
	docker push $(IMAGE):$(@:push-%=%)-$(VERSION)

dockerbuild-%: build-% qemu
	docker build --build-arg ARCH=$(@:dockerbuild-%=%) -t $(IMAGE):$(@:dockerbuild-%=%)-$(VERSION) .

manifest: push
	@echo "manifest can be purged with: docker manifest push --purge $(IMAGE):$(VERSION)"
	docker manifest create $(IMAGE):$(VERSION) \
		$(ARCHS:%=$(IMAGE):%-$(VERSION))
	@echo "manifest can be pushed with: make manifest-push"

manifest-purge:
	docker manifest push --purge $(IMAGE):$(VERSION)

manifest-push:
	docker manifest push $(IMAGE):$(VERSION)

release: build-amd64 manifest-purge manifest manifest-push

# binfmt support for cross-building docker images
qemu:
	mkdir -p qemu-statics
	cp /usr/bin/qemu-arm-static qemu-statics/
	cp /usr/bin/qemu-aarch64-static qemu-statics/
