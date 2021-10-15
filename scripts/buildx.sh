#!/bin/sh

do_buildx()
{
    echo "building $* ..."

    if [ -n "$4" ];then
        docker_file="-f $3/$4"
    fi

    # enabled Docker CLI experimental feature
    cat "$HOME"/.docker/config.json
    echo ""

    # Docker Engine 19.03+ with --experimental=true and enable buildkit feature (check docker daemon config doc)
    #cat /etc/docker/daemon.json
    docker version

    # docker build multi-arch images and push to registry (--push)
    # local docker build image, use '--platform linux/amd64' or '--platform linux/arm64' (-o type=docker)
    # example: docker buildx build --platform linux/amd64 --force-rm=true -t $3 docker_file $3/ -o type=docker
    echo "docker buildx build --platform linux/arm64,linux/amd64 -t $1:$2 -f $3/$4 $3 --build-arg CodeSource=$5 --build-arg CodeBranch=$6 --build-arg CodeVersion=$7 --build-arg BuildDate=$8 --push"
    check_call docker buildx build --platform linux/arm64,linux/amd64 -t "$1:$2" "$docker_file" "$3" --build-arg CodeSource="$5" --build-arg CodeBranch="$6" --build-arg CodeVersion="$7" --build-arg BuildDate="$8" --push

	  # list multi-arch images digests
	  echo "docker buildx imagetools inspect $1:$2"
    check_call docker buildx imagetools inspect "$1:$2"

    # due to multi-arch build docker local image doesn't contain the built image, need to pull linux/amd64 image from registry
    echo "docker pull $1:$2"
    check_call docker pull "$1:$2"

	  # re-tag linux/amd64 image as latest
	  echo "docker tag $1:$2 $1"
	  check_call docker tag "$1:$2" "$1"

	  # show linux/amd64 image
	  echo "docker history $1:$2"
	  check_call docker history "$1:$2"
}

check_call()
{
    eval "$@"
    rc=$?
    if [ ${rc} -ne 0 ]; then
        echo "[$*] execute fail: $rc"
        exit 1
    fi
}

do_buildx "$@"
