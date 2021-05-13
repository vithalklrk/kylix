#!/bin/bash
#
# Install kylix
#
cp ./kylix /usr/bin
mkdir -p /usr/share/kylix/pictures/flags
cp ./kylix.ui /usr/share/kylix/
cp ./images/* /usr/share/kylix/pictures
cp ./flags/* /usr/share/kylix/pictures/flags
