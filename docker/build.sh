#!/bin/bash
echo Building container
docker build -t gobuilder "$(pwd)"
echo Running make script
docker run --rm -v "/$(pwd)/.."://usr/src/myapp -w //usr/src/myapp gobuilder make
read -p "Done. Press enter to continue"