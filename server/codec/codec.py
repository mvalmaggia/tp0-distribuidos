from datetime import datetime
from common.utils import Bet

def decode_bet(raw_message: str) -> Bet:
    fields = {}
    for part in raw_message.strip().split("|"):
        if ":" in part:
            key, value = part.split(":", 1)
            fields[key] = value

    return Bet(
        agency=fields.get("agency", "0"),
        first_name=fields.get("first_name", ""),
        last_name=fields.get("last_name", ""),
        document=fields.get("dni", ""),
        birthdate=fields.get("birthdate", "1900-01-01"),
        number=fields.get("number", "0")
    )

def decode_bet_batch(raw_message: str) -> list[Bet]:
    if raw_message.startswith("BET_BATCH;"):
        raw_message = raw_message[len("BET_BATCH;"):]

    individual_bets = raw_message.strip().split(';')

    decoded_bets = []
    for bet_str in individual_bets:
        if bet_str:
            decoded_bets.append(decode_bet(bet_str))
    
    return decoded_bets

def encode_winners(winners: list[str]) -> str:
    return ';'.join(winners)