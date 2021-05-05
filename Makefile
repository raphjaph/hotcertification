
binaries := cmd/certificationserver/certserver cmd/client/client cmd/keygen/keygen 

.PHONY: all clean $(binaries)

all: $(binaries)

$(binaries):
	go build -o ./$@ ./$(dir $@)

clean:
	rm -fv $(binaries)
