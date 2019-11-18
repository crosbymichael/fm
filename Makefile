all:
	@go build -v

install:
	@install fm /usr/local/bin/
