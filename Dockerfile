FROM golang:alpine
RUN mkdir /app 

ARG PORT=8080
ADD . /app/
WORKDIR /app 

RUN go build -o main .
RUN adduser -S -D -H -h /app appuser

EXPOSE $PORT
ENV PORT $PORT
USER appuser
CMD ["./main"]
