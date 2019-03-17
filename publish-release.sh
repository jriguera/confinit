#!/usr/bin/env bash -e
# set -o pipefail  # exit if pipe command fails
[ -z "$DEBUG" ] || set -x

##

RELEASE="confinit"
VERSION=${VERSION:-}
DESCRIPTION="Managing configuration at boot time"
GITHUB_REPO="jriguera/confinit"

###

MAKE="make"
JQ="jq"
CURL="curl -s"
SHA1="sha1sum -b"

# Create a personal github token to use this script
if [ -z "$GITHUB_TOKEN" ]
then
    echo "Github TOKEN not defined!"
    echo "See https://help.github.com/articles/creating-an-access-token-for-command-line-use/"
    exit 1
fi

# You need bosh installed and with you credentials
if ! [ -x "$(command -v $MAKE)" ]
then
    echo "ERROR: $MAKE command not found! Please install it and make it available in the PATH"
    exit 1
fi

# You need jq installed
if ! [ -x "$(command -v $JQ)" ]
then
    echo "ERROR: $JQ command not found! Please install it and make it available in the PATH"
    exit 1
fi

### Parse args
case $# in
    0)
        echo "*** Creating a new release. Automatically calculating version number"
        ;;
    1)
        if [ $1 == "-h" ] || [ $1 == "--help" ]
        then
            echo "Usage:  $0 [version-number]"
            echo "  Creates a release, commits the changes to this repository using tags and uploads "
            echo "  the release to Github Releases and the final Docker image to Docker Hub. "
            echo "  It also adds comments based on previous git commits."
            exit 0
        else
            VERSION=$1
            if ! [[ $VERSION =~ $RE_VERSION_NUMBER ]]
            then
                echo "ERROR: Incorrect version number!"
                exit 1
            fi
            echo "*** Creating a new release. Using release version number $VERSION."
        fi
        ;;
    *)
        echo "ERROR: incorrect argument. See '$0 --help'"
        exit 1
        ;;
esac

echo "* Removing old binaries ..."
make clean-build

# Get the last git commit made by this script
LASTCOMMIT=$(git show-ref --tags -d | tail -n 1)
if [ -z "$LASTCOMMIT" ]
then
    echo "* Changes since the beginning: "
    CHANGELOG=$(git log --pretty="%h %aI %s (%an)" | sed 's/^/- /')
else
    echo "* Changes since last version with commit $LASTCOMMIT: "
    CHANGELOG=$(git log --pretty="%h %aI %s (%an)" "$(echo $LASTCOMMIT | cut -d' ' -f 1)..@" | sed 's/^/- /')
fi
if [ -z "$CHANGELOG" ]
then
    echo "ERROR: no commits since last release with commit $LASTCOMMIT!. Please "
    echo "commit your changes to create and publish a new release!"
    exit 1
fi
echo "$CHANGELOG"

# Creating the release
if [ -z "$VERSION" ]
then
    VERSION=$(head -n1 VERSION)
    echo "* Creating final release version $VERSION (from VERSION) ..."
else
    echo "* Creating final release version $VERSION (from env/input)..."
fi

# Make
echo "* Generating binaries ..."
make all

# Create annotated tag
echo "* Creating a git tag ... "
git tag -a "v$VERSION" -m "$RELEASE v$VERSION"
git push --tags

# Create a release in Github
echo "* Creating a new release in Github ... "
DESC=$(cat <<-EOF
	# $RELEASE version $VERSION
	$DESCRIPTION
	## Changes since last version
	$CHANGELOG
EOF
)
printf -v data '{"tag_name": "v%s","target_commitish": "master","name": "v%s","body": %s,"draft": false, "prerelease": false}' "$VERSION" "$VERSION" "$(echo "$DESC" | $JQ -R -s '@text')"
releaseid=$($CURL -H "Authorization: token $GITHUB_TOKEN" -H "Content-Type: application/json" -XPOST --data "$data" "https://api.github.com/repos/$GITHUB_REPO/releases" | $JQ '.id')
# Upload the release
echo "* Uploading binaries to Github releases section ... "
for release in build/*
do
    echo -n "  URL: "
    $CURL -H "Authorization: token $GITHUB_TOKEN" -H "Content-Type: application/octet-stream" --data-binary @"${release}" "https://uploads.github.com/repos/$GITHUB_REPO/releases/$releaseid/assets?name=$(basename ${release})" | $JQ -r '.browser_download_url'
done

# Finish
echo
echo "*** Description https://github.com/$GITHUB_REPO/releases/tag/v$VERSION: "
echo
echo "$DESC"

exit 0
