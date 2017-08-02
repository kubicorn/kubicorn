#!/bin/bash
VERBOSE=FALSE
VERBOSE_DOCKER_RUN=""
VERBOSE_DOCKER_BUILD="-q"
REMOVE_IMAGE=FALSE

for i in "$@"
do
case $i in
    -v|--verbose)
    VERBOSE=TRUE
    VERBOSE_DOCKER_RUN="-e 'VERBOSE=1'"
    VERBOSE_DOCKER_BUILD=""
    shift # past argument=value
    ;;
    --remove-image)
    REMOVE_IMAGE=TRUE
    shift # past argument=value
    ;;
esac
done

if ${REMOVE_IMAGE} ; then
    echo removing old container if exists
    docker image rm gobuilder
fi
echo Building container
docker build -t gobuilder "$(pwd)" ${VERBOSE_DOCKER_BUILD}
echo Running make script
docker run --rm -v "/$(pwd)/.."://go/src/github.com/kris-nova/kubicorn -w //go/src/github.com/kris-nova/kubicorn gobuilder make ${VERBOSE_DOCKER_RUN}
read -p "Done. Press enter to continue"