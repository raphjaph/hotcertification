
binaries := cmd/certificationserver/certserver cmd/client/client cmd/keygen/keygen benchmark/benchmark

.PHONY: all clean $(binaries) 

all: $(binaries)

$(binaries):
	go build -o ./$@ ./$(dir $@)

benchmark: 
	go build -o ./benchmark/benchmark ./benchmark

clean:
	rm -fv $(binaries)
