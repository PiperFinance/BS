import requests
from pprint import pprint
from web3 import Web3
from pymongo import MongoClient

client = MongoClient("mongodb://localhost:27017/")
rpc = "https://bsc-dataseed3.binance.org"
w3 = Web3(Web3.HTTPProvider(rpc))
abi = requests.get(
    "https://raw.githubusercontent.com/PiperFinance/CD/main/abi/token.abi"
).json()


def scanner_bal(tok, add):
    q = client["BS_T_1_56"]["UsersBalance"].find(
        filter={"userStr": add, "tokenStr": tok}
    )
    return q.next()["bal"]


def bal(tok, add):
    return (
        Web3(Web3.HTTPProvider(rpc))
        .eth.contract(Web3.to_checksum_address(tok), abi=abi)
        .functions.balanceOf(Web3.to_checksum_address(add))
        .call()
    )


def run():
    cases = {
        (
            "0x9A2478C4036548864d96a97Fbf93f6a3341fedac",
            "0x0718a4633C8cc350a52580c13C7d1303b8C684Ef",
        ),
    }
    for case in cases:
        pprint({"case": case, "bs": scanner_bal(*case), "w3": bal(*case)}, indent=2)


def run2():
    while True:
        case = (input("tok> "), input("add> "))
        pprint({"case": case, "bs": scanner_bal(*case), "w3": bal(*case)}, indent=2)


run2()
