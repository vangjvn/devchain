===============
Run
===============

This document describes how to run the Second State DevChain.

Binary
----------------------------

The binary executable `devchain` is the software that runs blockchain nodes.

Single node
````````````

First, you need to initialize the configurations and settings on the node computer.

.. code:: bash

  $ devchain node init --home $HOME/.devchain

The `genesis.json` and `config.toml` files will be created under the `$HOME/.devchain/config` directory. You can make changes to those files to customize your blockchain.
Then, you can start the node.

.. code:: bash

  $ devchain node start

Next, in a new terminal window, run the following command to connect to the local DevChain node.

.. code:: bash

  $ devchain attach http://localhost:8545
  > cmt.syncing
  {
    catching_up: false,
    latest_app_hash: "07FA113DF14AAC49773DD7EE2B8418740D9DD552",
    latest_block_hash: "AF1415AF0057C52C4A1F7DC80298217A33291AEE",
    latest_block_height: 23,
    latest_block_time: "2019-05-03T21:41:14.581000291Z"
  }


Multiple nodes
```````````````

TBD


Docker
----------------------------

In the previous section, we have built a Docker image for the node software under the name `second-state/devchain`.

Single node
```````````````

First, you need to initialize the configurations and settings on the node computer.

.. code:: bash

  $ docker run --rm -v $HOME/.devchain:/devchain second-state/devchain node init --home /devchain

The `genesis.json` and `config.toml` files will be created under the `$HOME/.devchain/config` directory. You can make changes to those files to customize your blockchain. You may need to `sudo su -` in order to edit those files since they are created by the root user. Then, you can start the node.

.. code:: bash

  $ docker run --rm -v $HOME/.devchain:/devchain -p 26657:26657 -p 8545:8545 second-state/devchain node start --home /devchain

You can run the ID of the running Docker container.

.. code:: bash

  $ docker container ls
  CONTAINER ID        IMAGE                   COMMAND                  CREATED             STATUS              PORTS                                                         NAMES
  0bcd9da5bf05        second-state/devchain   "./devchain node sta…"   4 minutes ago       Up 4 minutes        0.0.0.0:8545->8545/tcp, 0.0.0.0:26657->26657/tcp, 26656/tcp   pedantic_mendeleev

Next, log into that container.

.. code:: bash

  $ docker exec -i -t 0bcd9da5bf05 bash
  root@0bcd9da5bf05:/app# ls
  devchain  devchain.sha256  lib

Finally, you can attach a console to the node to run web3 commands.

.. code:: bash

  root@0bcd9da5bf05:/app# ./devchain attach http://localhost:8545
  ...
  > cmt.syncing
  {
    catching_up: false,
    latest_app_hash: "C7D8AECE081DF06FFC9BF6144A50B37CA5DD8A8E",
    latest_block_hash: "B592D63AB78C571E0FB695A052681E65F6DFE15B",
    latest_block_height: 35,
    latest_block_time: "2019-05-04T02:59:30.542783017Z"
  }


Multiple nodes
```````````````

TBD


