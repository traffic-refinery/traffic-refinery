#!/bin/bash
PROJECT=$(basename $(dirname $(readlink -f $0)))

NAMES=$(ls cmd/* -d | xargs -n1 basename)
for NAME in $NAMES; do
    OSES=${OSS:-"linux"}
    ARCHS=${ARCHS:-"amd64"}
    for ARCH in $ARCHS; do
        for OS in $OSES; do
            echo $OS $ARCH $NAME
            GOOS=${OS} GOARCH=${ARCH} GOARM=7 go build -o build/${NAME}-${OS}-${ARCH} cmd/${NAME}/*.go
            if [ $? -eq 0 ]; then
                echo OK
            fi
        done
    done
done

echo "Resulting files:"
find build -type f
