FROM golang:1.18-bullseye

RUN go install github.com/cosmtrek/air@latest

ENV APP_HOME /go/src/mathapp
RUN mkdir -p "$APP_HOME"

#COPY ./src "$APP_HOME"
WORKDIR "$APP_HOME"


EXPOSE 8010

CMD [ "air","run" ]