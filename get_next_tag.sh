#!/bin/bash

usage() {
    echo "USAGE:"
    echo "./get_next_tag.sh COMMIT_MESSAGE CURRENT_TAG"
    echo "COMMIT_MESSAGE should have the format Merge pull request <PR #> from Yuruh/<PR Name>"
    echo "CURRENT_TAG should have the format <major>.<minor>.<fix>"
}

commit_msg=$1
tag=$2

if [ $# -lt 2 ]
then
  usage;
  exit 1;
fi



# Coud be refactored to find n value
major=$(gawk 'match($0, /[0-9]+/) {
  print substr($0, RSTART, RLENGTH)
};' <<< $tag)

minor=$(gawk 'match($0, /[0-9]+/) {
  remaining=substr($0, RLENGTH + 2)
}; match(remaining, /[0-9]+/) {
  print substr(remaining, RSTART, RLENGTH)
};' <<< $tag)


name_position=$(gawk -v a="$commit_msg" -v b="Yuruh/" "BEGIN{print index(a,b)}")
let name_position+=6

branch_type=$(gawk -v a="$commit_msg" -v b=$name_position "BEGIN{print substr(a,b,5)}")

if [ $name_position -gt 0 ]
then
  if [ $branch_type = "MAJOR" ]
  then
    let "major+=1"
    let "minor=0"
  elif [ $branch_type = "MINOR" ]
  then
    let "minor+=1"
  fi
fi

result_version="$major.$minor"

echo "$result_version"