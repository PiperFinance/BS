# What

Binding contracts to golang modules are done here

## How

```bash
abigen --abi erc1155/ERC1155.abi --out erc1155.go --pkg contracts -type ERC1155

solc erc20.sol --abi -o erc20 && \
    abigen --abi erc20/ERC20.abi --out erc20.go --pkg contracts --type ERC20
```

```zsh
☁  erc20 [dev] ⚡  solc erc20.sol  --overwrite --abi -o ./
Compiler run successful. Artifact(s) can be found in directory "./".
☁  erc20 [dev] ⚡  abigen --abi ERC20.abi --out ../erc20.go --pkg contracts --type ERC20
```
