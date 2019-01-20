#!/bin/bash

cd indexer
GOOS=linux GOARCH=amd64 go build -o ../dist/indexer

cd ../media-classifier
GOOS=linux GOARCH=amd64 go build -o ../dist/media-classifier

cd ../findaphotoserver
GOOS=linux GOARCH=amd64 go build -o ../dist/findaphotoserver

cd ../..
