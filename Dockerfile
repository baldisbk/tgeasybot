# syntax=docker/dockerfile:1

FROM golang AS builder

# pre-copy/cache go.mod for pre-downloading dependencies
# and only redownloading them in subsequent builds if they change
WORKDIR /usr/src/bot
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# too much random bullsh#t in root directory
# dont wanna even bother with dockerignore
COPY ./cmd/ ./cmd/
COPY ./internal/ ./internal/

# need to mount output directory...
# FROM builder as unittest
# RUN go test -coverprofile=coverage.out ./...
# RUN go tool cover -html=coverage.out -o cover.html

FROM builder AS botbuild
RUN CGO_ENABLED=0 go build -o ./cmd/api/bot ./cmd/api/main.go

FROM alpine AS bot
RUN apk add yq
RUN wget "https://storage.yandexcloud.net/cloud-certs/CA.pem" --output-document root.crt
COPY --from=botbuild /usr/src/bot/cmd/api/bot .
COPY --from=botbuild /usr/src/bot/cmd/api/config.yaml .
RUN --mount=type=secret,env=TGBOT_TG_TOKEN,id=TGBOT_TG_TOKEN yq -i '.tgapi.token = "'$TGBOT_TG_TOKEN'"' config.yaml
RUN --mount=type=secret,env=TGBOT_DB_PASSWORD,id=TGBOT_DB_PASSWORD yq -i '.db.password = "'$TGBOT_DB_PASSWORD'"' config.yaml
ENTRYPOINT [ "./bot",  "--config", "config.yaml" ]
