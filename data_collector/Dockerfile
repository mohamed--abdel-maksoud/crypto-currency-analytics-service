FROM golang

WORKDIR /go/src/app
COPY . .

RUN go get github.com/Jeffail/gabs github.com/mongodb/mongo-go-driver/mongo

CMD [ "go", "run", "data_collector/data_collector.go" ]
