# !/bin/bash

rm -Rf dist
mkdir dist
mkdir -p dist/content/dist

cd indexer
GOOS=linux GOARCH=amd64 go build -o ../dist/indexer

cd ../findaphotoserver
GOOS=linux GOARCH=amd64 go build -o ../dist/findaphotoserver

cd content
ng build
cp -R dist ../../dist/content

cd ../..
