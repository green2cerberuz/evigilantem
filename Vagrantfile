# -*- mode: ruby -*-
# vi: set ft=ruby :

# All Vagrant configuration is done below. The "2" in Vagrant.configure
# configures the configuration version (we support older styles for
# backwards compatibility). Please don't change it unless you know what
# you're doing.
Vagrant.configure("2") do |config|
  # The most common configuration options are documented and commented below.
  # For a complete reference, please see the online documentation at
  # https://docs.vagrantup.com.

  # Every Vagrant development environment requires a box. You can search for
  # boxes at https://vagrantcloud.com/search.
  config.vm.box = "hashicorp/bionic64"
  config.vm.box_version = "1.0.282"
  config.ssh.forward_agent = true
  config.ssh.forward_x11 = true
  config.ssh.shell="bash"
  if Vagrant.has_plugin?("vagrant-vbguest") then
    config.vbguest.auto_update = false
  end
  if Vagrant.has_plugin?("vagrant-env") then
    config.env.enable
  end
  # command to run init script in our ubuntu machine
  config.vm.provision :shell, path: "bootstrap.sh"
  config.vm.provision "shell", env: {"GOROOT"=>ENV['GOROOT'], "GOPATH"=>ENV['GOPATH'], "CGO_LDFLAGS"=>ENV['CGO_LDFLAGS']}, inline: <<-SHELL
    echo "export GOROOT=$GOROOT" >> /home/vagrant/.profile
    echo "export GOPATH=$GOPATH" >> /home/vagrant/.profile
    echo "export PATH=$PATH:$GOROOT/bin:$GOPATH/bin" >> /home/vagrant/.profile
    echo "export CGO_LDFLAGS=$CGO_LDFLAGS" >> /home/vagrant/.profile
    source /home/vagrant/.profile
    echo -n "Installing go-gl bindings"
    go get -u github.com/go-gl/glfw/v3.3/glfw
    go get -u github.com/go-gl/gl/v4.1-core/gl
  SHELL

end
