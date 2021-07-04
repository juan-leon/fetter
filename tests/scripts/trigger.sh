#!/bin/bash -eu

readonly OUTFILE="/tmp/$(basename $2)"
echo "$1" >$OUTFILE

if test -v VAR1; then
    echo "$VAR1" >>$OUTFILE
fi
