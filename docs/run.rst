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

TBD

Single node
```````````````

TBD


Multiple nodes
```````````````

TBD


