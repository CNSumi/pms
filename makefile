GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
TARGET=pms
TARGET7=$(TARGET)7

all: clean build
build:
	$(GOBUILD) -ldflags "-s -w" -o $(TARGET7)
	upx -9 -f -o $(TARGET) $(TARGET7)
	rm -f $(TARGET7)
clean:
	$(GOCLEAN)
	rm -f $(TARGET) $(TARGET7)
pms:
	$(GOCLEAN)
	$(GOBUILD) -o $(TARGET)
