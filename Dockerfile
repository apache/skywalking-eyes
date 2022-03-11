# TODO: minimize the docker image size, now 524MB !!!

FROM golang:1.16-alpine AS build

WORKDIR /license-eye

COPY . .

RUN apk add --no-cache make curl && make linux

FROM alpine:3 AS bin

COPY --from=build /license-eye/bin/linux/license-eye /bin/license-eye

# Go
COPY --from=build /usr/local/go/bin/go /usr/local/go/bin/go
ENV PATH="/usr/local/go/bin:$PATH"
RUN apk add --no-cache bash gcc musl-dev npm
# Go

WORKDIR /github/workspace/

ENTRYPOINT ["/bin/license-eye"]
