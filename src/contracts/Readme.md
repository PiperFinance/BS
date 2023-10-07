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

## LP tokens

### AMM V2

Uniswap V2 based tokens emit following :

- Swap :
  ```sol
  emit Swap(msg.sender, amount0In, amount1In, amount0Out, amount1Out, to);
  ```
- Mint - LP :
  ```sol
  event Mint(address indexed sender, uint amount0, uint amount1);
  ```
- Burn - LP :
  ```sol
  event Burn(address indexed sender, uint amount0, uint amount1, address indexed to);
  ```
- PairCreated - Factory:
  ```sol
  emit PairCreated(token0, token1, pair, allPairs.length);
  ```
