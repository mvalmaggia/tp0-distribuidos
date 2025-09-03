from datetime import datetime
from common.utils import Bet

def decode_bet(raw_message: str) -> Bet:
    """
    Decode a bet string into a Bet object.
    Example input: "agency:1|dni:123|numero:5|nombre:Matias|apellido:Valmaggia|fecha:1995-07-21"
    """
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
