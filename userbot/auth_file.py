"""
Auth via file — sends code, polls for code in a file, signs in.
1. Run this script
2. It sends the verification code
3. Create a file called 'code.txt' with the code in it
4. Script picks it up and signs in
"""
import os
import time
import asyncio
from dotenv import load_dotenv
from pyrogram import Client

load_dotenv()

CODE_FILE = "code.txt"

async def main():
    # Remove old code file
    if os.path.exists(CODE_FILE):
        os.remove(CODE_FILE)

    app = Client(
        "rooted_userbot",
        api_id=int(os.getenv("API_ID", "0")),
        api_hash=os.getenv("API_HASH", ""),
        phone_number=os.getenv("PHONE_NUMBER", ""),
    )

    await app.connect()

    phone = os.getenv("PHONE_NUMBER", "")
    sent_code = await app.send_code(phone)
    print(f"Verification code sent to {phone}")
    print(f"Write the code to {CODE_FILE} and I'll pick it up...")

    # Poll for code file
    code = None
    for _ in range(60):  # wait up to 60 seconds
        if os.path.exists(CODE_FILE):
            with open(CODE_FILE) as f:
                code = f.read().strip()
            if code:
                break
        await asyncio.sleep(1)

    if not code:
        print("Timed out waiting for code.")
        await app.disconnect()
        return

    print(f"Got code: {code}")

    try:
        await app.sign_in(phone, sent_code.phone_code_hash, code)
    except Exception as e:
        print(f"Sign in failed: {e}")
        await app.disconnect()
        return

    me = await app.get_me()
    print(f"Authenticated as: {me.first_name} (ID: {me.id})")
    print("Session file: rooted_userbot.session")

    # Cleanup
    os.remove(CODE_FILE)
    await app.disconnect()

asyncio.run(main())
