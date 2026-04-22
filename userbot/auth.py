"""
First-time authentication for the Rooted userbot.
Run this once to create the session file, then main.py will use it.
"""
import os
from dotenv import load_dotenv
from pyrogram import Client

load_dotenv()

app = Client(
    "rooted_userbot",
    api_id=int(os.getenv("API_ID", "0")),
    api_hash=os.getenv("API_HASH", ""),
    phone_number=os.getenv("PHONE_NUMBER", ""),
)

with app:
    me = app.get_me()
    print(f"Authenticated as: {me.first_name} (@{me.username})")
    print(f"Session file created: rooted_userbot.session")
    print("You can now run main.py without entering the code again.")
