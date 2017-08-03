#!/bin/bash
VERBOSE=FALSE
VERBOSE_DOCKER_RUN=""
VERBOSE_DOCKER_BUILD="-q"
REMOVE_IMAGE=FALSE
SHOW_HELP=FALSE

which docker >/dev/null 2>&1 || { echo >&2 "Docker is required but it's not installed. Aborting."; exit 1; }

for i in "$@"
do
case $i in
    -v|--verbose)
    VERBOSE=TRUE
    VERBOSE_DOCKER_RUN="-e 'VERBOSE=1'"
    VERBOSE_DOCKER_BUILD=""
    shift # past argument=value
    ;;
    -i|--remove-image)
    REMOVE_IMAGE=TRUE
    shift # past argument=value
    ;;
    -h|--help)
    SHOW_HELP=TRUE
    shift # past argument=value
    ;;
esac
done

if ${SHOW_HELP} ; then
    echo ""
    echo "Usage:  ./build.sh"
    echo ""
    echo "A script to build Kubicorn with Docker."
    echo ""
    echo "Flags:"
    echo "    -v, --verbose        Writes output to terminal"
    echo "    -i, --remove-image   Removes the docker image before building"
    echo "    -h, --help           Outputs all flags"
    echo ""
    exit 0;
fi

if ${REMOVE_IMAGE} ; then
    echo removing old container if exists
    docker image rm gobuilder-kubicorn
fi

if ${VERBOSE} ; then
    echo Building container
fi
docker build -t gobuilder-kubicorn "$(pwd)" ${VERBOSE_DOCKER_BUILD}

if ${VERBOSE} ; then
    echo Running make script
fi
docker run --rm -v "/$(pwd)/.."://go/src/github.com/kris-nova/kubicorn -w //go/src/github.com/kris-nova/kubicorn gobuilder-kubicorn make ${VERBOSE_DOCKER_RUN}

read -p "Done. Press enter to continue"