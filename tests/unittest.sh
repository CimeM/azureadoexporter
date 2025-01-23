#!/bin/bash

# build the image
docker build --tag docker-gs-ping ..
docker run --rm --env-file .env docker-gs-ping