from web3 import Web3
import requests
from pymongo import MongoClient

abi = requests.get(
    "https://raw.githubusercontent.com/PiperFinance/CD/main/abi/token.abi"
).json()
rpc = "https://rpc.ankr.com/bsc"


def xbalRemote():
    client = MongoClient("mongodb://piperFinance:Piper2022@localhost:27017/")
    tok = input("tok> ")
    add = input("add> ")
    bl = input("block> ")
    if bl:
        return (
            Web3(Web3.HTTPProvider(rpc))
            .eth.contract(Web3.to_checksum_address(tok), abi=abi)
            .functions.balanceOf(Web3.to_checksum_address(add))
            .call(block_identifier=int(bl))
        ), client.BS_56.UsersBalance.find_one({"tokenStr": tok, "userStr": add})
    else:
        return (
            (
                Web3(Web3.HTTPProvider(rpc))
                .eth.contract(Web3.to_checksum_address(tok), abi=abi)
                .functions.balanceOf(Web3.to_checksum_address(add))
            ),
            client.BS_56.UsersBalance.find_one({"tokenStr": tok, "userStr": add}),
            Web3(Web3.HTTPProvider(rpc)).eth.block_number,
        )


if __name__ == "__main__":
    while True:
        try:
            __import__("pprint").pprint(xbalRemote())
        except KeyboardInterrupt:
            exit(0)
        except Exception as e:
            print(e)
            if input("Continue ?[Y/N]").upper() != "Y":
                break
