
FROM larjim/kademlialab:latest

COPY . /home/go/src/main/



WORKDIR /home/go/src/main

CMD ["./main"]
