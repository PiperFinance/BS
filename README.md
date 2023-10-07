# BS

For Setting up Contract Bindings ...
```bash
## ERC20
solc --abi erc20.sol -o erc20
abigen --abi erc20/ERC20.abi --out erc20.go --pkg contracts --type ERC20
## erc721
```