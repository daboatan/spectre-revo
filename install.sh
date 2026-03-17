#!/bin/bash

echo "==BEGIN=="
echo "### DEBUG ONLY, PREFER CONTAINER SOLUTION ###"

function base_install()
{
  sudo apt install -y wget curl git unzip rsync

  # INSTALLATION PYGMENTS:
  #wget https://bootstrap.pypa.io/get-pip.py #error: externally-managed-environment
  #sudo python3 get-pip.py #error: externally-managed-environment
  sudo apt install -y python3-pip
  #sudo apt install python3 python3-distutils #distutils has been deprecated in Python 3.12
  sudo python3 -m pip install setuptools --break-system-packages
  sudo python3 -m pip install Pygments --break-system-packages

  # INSTALLATION NVM --> NODEJS/NPM
  wget -qO- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.1/install.sh | bash
  export NVM_DIR="$([ -z "${XDG_CONFIG_HOME-}" ] && printf %s "${HOME}/.nvm" || printf %s "${XDG_CONFIG_HOME}/nvm")"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"

  nvm install lts/Jod #node v22.14.0 | npm v10.9.2
  nvm use lts/jod
  nvm alias default lts/jod

  # INSTALLATION GO
  wget https://go.dev/dl/go1.24.0.linux-amd64.tar.gz
  export PATH=$PATH:/usr/local/go/bin
  sudo tar -C /usr/local -xzf go1.24.0.linux-amd64.tar.gz

  
  echo "#####"
  echo "add this 2 lines into bashrc/zshrc/... :"
  echo ""
  echo "export PATH=$PATH:/usr/local/go/bin"
  echo 'export NVM_DIR="$([ -z "${XDG_CONFIG_HOME-}" ] && printf %s "${HOME}/.nvm" || printf %s "${XDG_CONFIG_HOME}/nvm")"'
  echo ""
  echo "#####"

}


function dependancy_install()
{
  npm install
  go get
  go install

  if [ ! -d logs ]; then
    mkdir logs
  fi
  if [ ! -d data ]; then
    mkdir data
  fi

}

### (UN)COMMENT FOR INSTALL BASE:
base_install

### (UN)COMMENT FOR INSTALL DEPENDANCY:
dependancy_install

go build

PORT="${PORT:-8619}"
echo "# SPECTRE LAUNCHED ON PORT $PORT"
./spectre-updated -addr="0.0.0.0:$PORT" -log_dir="logs" -root="data" --logtostderr=1

echo "===END==="
