# Image de base
FROM golang:1.25.1-alpine

# Installer protoc, bash et git
RUN apk add --no-cache protobuf protobuf-dev bash git

# Installer les plugins Go pour protoc
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest \
    && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Ajouter le répertoire Go bin au PATH
ENV PATH="/go/bin:${PATH}"

# Dossier de travail
WORKDIR /defs

# L'image est prête à l'emploi, aucun build supplémentaire nécessaire
ENTRYPOINT ["protoc"]