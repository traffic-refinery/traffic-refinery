all:
	go build -o tr cmd/traffic-refinery/tr.go

ring:
	go build -o tr -tags=ring cmd/traffic-refinery/tr.go
	
pf:
	go build -o tr -tags=pf cmd/traffic-refinery/tr.go

clean:
	rm -f tr