EXECUTABLE := server

default:
	go build -o $(EXECUTABLE) main.go

clean:
	rm $(EXECUTABLE)
