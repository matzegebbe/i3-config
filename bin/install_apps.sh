#!/bin/bash

apt-get update

for app in 
 curl vim ipcalc kazam shutter geany pinta tshark compton blueman bash-completion
do
    apt install $app
done

pip install yq

sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)"

echo "HISTSIZE=1000000000000" >> ~/.zshrc
echo "SAVEHIST=1000000000000" >> ~/.zshrc

add-apt-repository ppa:yubico/stable
apt-get install yubikey-manager-qt yubioath-desktop yubikey-personalization-gui
