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


