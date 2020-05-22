FROM golang:1.13

# Set the working directory outside $GOPATH to enable the support for modules.
WORKDIR /usr/src/diary

# Fetch dependencies first; they are less susceptible to change on every build
# and will therefore be cached for speeding up the next build
COPY ./go.mod ./go.sum ./

RUN go mod download

# Import the code from the context.
COPY . .

RUN go get github.com/pilu/fresh

EXPOSE 8080

# Build the executable to `/app`. Mark the build as statically linked.
#RUN go build -o /app .

# Run the compiled binary.
#ENTRYPOINT ["/app"]

# Run with live reload, for dev purposes.
CMD ["fresh"]
