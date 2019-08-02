#!/bin/bash

apt-get update

for app in 
 vim ipcalc kazam shutter geany pinta
do
    apt install $app
done

pip install yq

