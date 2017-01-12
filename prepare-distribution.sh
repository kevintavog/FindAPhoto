# !/bin/bash

rm -Rf dist
mkdir dist
mkdir -p dist/content/dist

cd indexer
GOOS=linux GOARCH=amd64 go build -o ../dist/indexer

cd ../media-classifier
GOOS=linux GOARCH=amd64 go build -o ../dist/media-classifier

cd ../findaphotoserver
GOOS=linux GOARCH=amd64 go build -o ../dist/findaphotoserver

cd content
ng build --prod
cp -R dist ../../dist/content

cd ../..
