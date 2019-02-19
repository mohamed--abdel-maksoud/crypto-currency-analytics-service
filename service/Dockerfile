FROM golang

WORKDIR /go/src/app
COPY . .

RUN go get github.com/go-chi/chi github.com/go-chi/render github.com/mongodb/mongo-go-driver/mongo

CMD [ "go", "run", "service/service.go" ]
