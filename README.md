# golang_api for Book Keeper

First of all, run this command in the terminal of the project
 1) cp .env.example .env

Installation 
 1) uncomment app, networks, volumes(to save data of DB) services in Dockerfile or if you have golang locally then follow third point
 2) docker-compose build
 3) docker-compose up

Using go
 1) follow second point or use "go mod init book-keeper" after deletion of go.mod and go.sum 
 2) go mod tidy
 3) go run main.go or make run

