#!/usr/bin/env bash
echo -n "Installing libglfw3 in vm"
apt-get update ; apt-get install -y libglfw3-dev gcc libgl1-mesa-dev xorg-dev xauth x11-apps # install for glfw environment
echo -n "Installing go environment"

## install go insie our virtualbox
curl -O https://dl.google.com/go/go1.4.linux-amd64.tar.gz
tar xvf go1.4.linux-amd64.tar.gz
sudo chown -R root:root ./go
sudo mv go /usr/local

# clean everything
echo -n "Cleaning installation"
rm ./go1.4.linux-amd64.tar.gz





