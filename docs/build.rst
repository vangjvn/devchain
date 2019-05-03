===============
Build
===============

This document describes how to build the Second State DevChain.

Binary build
----------------------------

Binary builds of the Second State DevChain are limited to Ubuntu 16.04 and CentOS 7.


Ubuntu 16.04
````````````

First, let's install and update necessary packages on a clean Ubuntu 16.04 install.

.. code:: bash

  $ sudo apt update -y
  $ sudo apt install -y curl wget git bison build-essential

Next, you must have GO language version 1.10+ installed. The easiest way to get GO 1.10 is through the GVM. Below are the commands.

.. code:: bash

  $ bash < <(curl -s -S -L https://raw.githubusercontent.com/moovweb/gvm/master/binscripts/gvm-installer)
  $ gvm install go1.10.3 -B
  $ gvm use go1.10.3 --default


Now, pull the devchain source code from Github, and then build the binary executable for Ubuntu 16.04.

.. code:: bash

  $ go get github.com/second-state/devchain
  $ cd $GOPATH/src/github.com/second-state/devchain
  $ make all


Once successful, the binary executable from the build is `$GOPATH/bin/devchain`.


.. code:: bash

  $ which devchain
  /home/ubuntu/.gvm/pkgsets/go1.10.3/global/bin/devchain


CentOS 7
````````

TBD


Docker build
----------------------------

TBD


