
binaries := cmd/certificationserver/certserver cmd/client/client cmd/keygen/keygen 

.PHONY: all clean $(binaries)

all: $(binaries)

$(binaries):
	go build -o ./$@ ./$(dir $@)

client: 
	go build -o ./cmd/client/client ./cmd/client 

clean:
	rm -fv $(binaries)
