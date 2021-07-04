#!/bin/bash -eu

readonly OUTFILE="/tmp/$(basename $2)"
echo "$1" >$OUTFILE

if test -v FETTER_VAR1; then
    echo "$FETTER_VAR1" >>$OUTFILE
fi
