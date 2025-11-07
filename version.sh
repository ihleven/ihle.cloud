#! /bin/bash

# https://gist.github.com/ericbmerritt/f52f1c48b86150704270
# Assumes that you tag versions with the version number (e.g., "1.1")
# and then the build number is that plus the number of commits since
# the tag (e.g., "1.1.17")

DESCRIBE=`git describe --tags --always --long`


TAG=`echo $DESCRIBE  | awk '{split($0,a,"-"); print a[1]}'`
NUM=`echo $DESCRIBE  | awk '{n=split($0,a,"-"); print a[n-1]}'`
HASH=`echo $DESCRIBE | awk '{n=split($0,a,"-"); print a[n]}'`
N=`echo $DESCRIBE | awk '{n=split($0,a,"-"); print n}'`

if [[ "${N}" = 4 ]]; then
    TAG+='-'
    TAG+=`echo $DESCRIBE  | awk '{split($0,a,"-"); print a[2]}'`
fi

echo ${TAG}.${NUM}-${HASH}