GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOINSTALL=$(GOCMD) install
GOTEST=$(GOCMD) test
GODEP=$(GOTEST) -i

GOPATH:=$(CURDIR)/_vendor:$(CURDIR)/

APPNAME = gopaperd
DISTDIR = dist

all: dist

clean:
	rm -rf bin
	rm -rf $(DISTDIR)

env:
	$(GO) env

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=386 $(GOINSTALL) gopaper/...
	CGO_ENABLED=0 $(GOINSTALL) gopaper/...

dist: clean build
	mkdir $(DISTDIR)
	cp bin/linux_386/$(APPNAME) $(DISTDIR)
	cp -R src/gopaper/cmd/gopaperd/static $(DISTDIR)
	cp -R src/gopaper/cmd/gopaperd/templates $(DISTDIR)
	tar -zcvf $(DISTDIR)/gopaper.tar.gz dist
	echo "Dist done"
