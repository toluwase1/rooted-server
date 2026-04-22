"""
Non-interactive auth — sends code, waits, then signs in.
Usage: python auth_noninteractive.py
Then enter the code when prompted (within 30 seconds).
"""
import os
import sys
import asyncio
from dotenv import load_dotenv
from pyrogram import Client

load_dotenv()

async def main():
    app = Client(
        "rooted_userbot",
        api_id=int(os.getenv("API_ID", "0")),
        api_hash=os.getenv("API_HASH", ""),
        phone_number=os.getenv("PHONE_NUMBER", ""),
    )

    await app.connect()

    phone = os.getenv("PHONE_NUMBER", "")
    sent_code = await app.send_code(phone)
    print("Verification code sent to Telegram. Waiting for code via stdin...")
    print("Enter code:", flush=True)

    # Read from stdin with a timeout
    loop = asyncio.get_event_loop()
    code = await loop.run_in_executor(None, sys.stdin.readline)
    code = code.strip()

    if not code:
        print("No code entered.")
        await app.disconnect()
        return

    try:
        await app.sign_in(phone, sent_code.phone_code_hash, code)
    except Exception as e:
        print(f"Sign in failed: {e}")
        await app.disconnect()
        return

    me = await app.get_me()
    print(f"Authenticated as: {me.first_name} (ID: {me.id})")
    print("Session file: rooted_userbot.session")

    await app.disconnect()

asyncio.run(main())
