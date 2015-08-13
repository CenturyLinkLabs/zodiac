#!/bin/bash

function main {
  displayLogo
  checkForDocker
  checkForCompose
  checkForZodiac
}

function checkForDocker {
  echo "Checking if the proper version of Docker exists..."
  which docker >/dev/null
  if [[ "$?" -ne "0" ]]; then
    alertNoDocker
  else
    docker -v | grep -w '1.[6-9]' 2>&1 >/dev/null
    if [[ "$?" -ne "0" ]]; then
      alertNoDocker
    else
      echo 'Supported Docker version already exists, continuing...'
    fi
  fi
}

function alertNoDocker {
  echo "Please install Docker version 1.6.* or later, then re-run this script.";
  exit 1;
}

function checkForCompose {
  echo "Checking if the proper version of Docker Compose exists..."
  which docker-compose >/dev/null
  if [[ "$?" -ne "0" ]]; then
    alertNoCompose
  else
    docker-compose -v | grep -w '1.[2-4].[0-9]' 2>&1 >/dev/null
    if [[ "$?" -ne "0" ]]; then
      alertNoCompose
    else
      echo 'Supported Docker Compose version already exists, skipping...'
    fi
  fi
}

function alertNoCompose {
  echo "Please install Docker Compose version 1.4.*, then re-run this script";
  exit 1;
}

function promptForComposeInstall {
  echo "Zodiac requires Docker Compose version 1.2 or higher.";
  read -p "Would you like us to install Docker Compose for you? [Y/n]" ic
  if [[ "$ic" == "" || "$ic" == "Y" || "$ic" == "y" ]]; then
    installCompose
  else
    exit 1;
  fi
}

function checkForZodiac {
  echo "Checking for existing Zodiac install..."
  which zodiac >/dev/null
  if [[ "$?" -ne "0" ]]; then
    installZodiac
  else
    zodiac -v | grep -w '0.2.1' 2>&1 >/dev/null
    if [[ "$?" -ne "0" ]]; then
      installZodiac
    else
      echo 'Zodiac is already installed.'
      exit 1;
    fi
  fi
}

function installCompose {
  echo "Installing Docker Compose..."
  curl -L https://github.com/docker/compose/releases/download/1.4.0/docker-compose-`uname -s`-`uname -m` > /usr/local/bin/docker-compose
  chmod +x /usr/local/bin/docker-compose
}

function installZodiac {
  echo "Installing Zodiac..."
  curl -L https://github.com/CenturyLinkLabs/zodiac/releases/download/0.2.1/zodiac-`uname -s`-`uname -m` > /usr/local/bin/zodiac
  chmod +x /usr/local/bin/zodiac
}


function displayLogo {
  tput clear
cat << "EOF"

MMMMMMMMMMMMd  `+dmmmmmmmy:   ommmmmmmmmdo`   hmmh      :mmmo       .sdmmmmmmdo`
yyyyyyyhMMMMh  mMMmhhhhdMMMo  sMMMdhhhhmMMN`  mMMm     `mMMMN-     .NMMmhhhhhhhy
      :mMMNs`  NMMy    -MMMs  sMMM/    sMMM`  mMMm     sMMNMMd`    -MMMs        
    `sNMMm/    NMMy    -MMMs  sMMM/    sMMM`  mMMm    -NMModMMo    -MMMs        
   -dMMMy.     NMMy    -MMMs  sMMM/    sMMM`  mMMm   `dMMd -NMN-   -MMMs        
 `oNMMN+       NMMy    -MMMs  sMMM/    sMMM`  mMMm   oMMMdyyNMMh   -MMMs        
.hMMMd-        NMMy    -MMMs  sMMM/    sMMM`  mMMm  -NMMmddddNMM+  -MMMs        
NMMMNyyyyyyyo  NMMmyyyyhMMMo  sMMMdyyyydMMN`  mMMm  dMMm.````/MMN. .MMMmyyyyyyyy
MMMMMMMMMMMMd  `oNMMMMMMMd/   sMMMMMMMMMNs.   mMMm +MMM+      hMMh  -yNMMMMMMNs.

by: CenturyLink Labs - https://labs.ctl.io/

Installing...

EOF
}

main "$@";
