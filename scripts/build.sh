#!/bin/sh

do_build()
{
    echo "building $* ..."

    TAG="$1:$2"
    DOCKER_FILE="-f $3/$4"

    echo "docker build --pull -t $TAG $DOCKER_FILE $3 --build-arg CodeSource=$5 --build-arg CodeBranch=$6 --build-arg CodeVersion=$7 --build-arg BuildDate=$8"
    check_call docker build --pull -t "$TAG" "$DOCKER_FILE" "$3" --build-arg CodeSource="$5" --build-arg CodeBranch="$6" --build-arg CodeVersion="$7" --build-arg BuildDate="$8"

	  echo "docker history $TAG"
	  check_call docker history "$TAG"

	  echo "docker push $TAG"
	  check_call docker push "$TAG"

    if [ "$CI_BUILD_REF_NAME" = "$MASTER_BRANCH" ] ; then
      echo "docker TAG $TAG $1";
      check_call docker TAG "$TAG" "$1";

      echo "docker push $1:latest";
      check_call docker push "$1":latest;
    fi
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

do_build "$@"
