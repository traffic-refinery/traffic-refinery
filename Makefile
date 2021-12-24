all:
	go build -o tr cmd/traffic-refinery/tr.go


# go build -o tr -tags=ring,afpakcet cmd/traffic-refinery/tr.go