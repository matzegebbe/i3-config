#!/bin/bash

apt-get update

for app in zsh git curl vim pip ipcalc cheese kazam shutter geany pinta tshark compton blueman bash-completion snapd i3	mc htop feh compiz parcellite
do
    apt-get install $app -y
done

pip install yq

sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)"

echo "HISTSIZE=1000000000000" >> ~/.zshrc
echo "SAVEHIST=1000000000000" >> ~/.zshrc

add-apt-repository ppa:yubico/stable
apt-get install yubikey-manager-qt yubioath-desktop yubikey-personalization-gui -y

# vbox 6
wget -q https://www.virtualbox.org/download/oracle_vbox_2016.asc -O- | apt-key add -
wget -q https://www.virtualbox.org/download/oracle_vbox.asc -O- | apt-key add -
add-apt-repository "deb http://download.virtualbox.org/virtualbox/debian bionic contrib"
apt-get update 
apt-get install virtualbox-6.0 -y

git config --global user.email "mathias.gebbe@gmail.com"
git config --global user.name "Mathias Gebbe"

apt-get autoremove -y && apt-get clean

snap install spotify
snap install vlc
snap install intellij-idea-community

apt-get install -y lightdm-webkit-greeter lightdm-settings
