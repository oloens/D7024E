
FROM larjim/kademlialab:latest

COPY main/main /home/go/src/main/main


#EXPOSE 8001
WORKDIR /home/go/src/main

CMD ["./main"]
