"""
Rooted Userbot — Creates Telegram groups for matched users and syncs messages.

This service runs alongside the Go backend. It uses Pyrogram (MTProto) to:
1. Create a supergroup when two users match
2. Add both users to the group
3. Listen for messages in match groups and save to DB
4. Send messages from Mini App to Telegram groups

The Go backend calls this service's HTTP API to trigger group creation
and to send messages from the Mini App to Telegram groups.
"""

import os
import asyncio
import logging
from contextlib import asynccontextmanager

import psycopg2
from dotenv import load_dotenv
from pyrogram import Client, filters
from pyrogram.errors import FloodWait
from pyrogram.types import Message
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel

load_dotenv()

# ==========================================
# SESSION PERSISTENCE (Cloud Storage)
# ==========================================

SESSION_FILE = "rooted_userbot.session"
GCS_BUCKET = os.getenv("GCS_SESSION_BUCKET", "")
GCS_BLOB = "userbot/rooted_userbot.session"


def download_session():
    """Download session file from Cloud Storage on startup."""
    if not GCS_BUCKET:
        log.info("No GCS_SESSION_BUCKET — using local session file")
        return
    try:
        from google.cloud import storage
        client = storage.Client()
        bucket = client.bucket(GCS_BUCKET)
        blob = bucket.blob(GCS_BLOB)
        if blob.exists():
            blob.download_to_filename(SESSION_FILE)
            log.info("Session downloaded from gs://%s/%s", GCS_BUCKET, GCS_BLOB)
        else:
            log.warning("No session in GCS — will need initial auth")
    except Exception as e:
        log.warning("Failed to download session: %s", e)


def upload_session():
    """Upload session file to Cloud Storage for persistence."""
    if not GCS_BUCKET or not os.path.exists(SESSION_FILE):
        return
    try:
        from google.cloud import storage
        client = storage.Client()
        bucket = client.bucket(GCS_BUCKET)
        blob = bucket.blob(GCS_BLOB)
        blob.upload_from_filename(SESSION_FILE)
        log.info("Session uploaded to gs://%s/%s", GCS_BUCKET, GCS_BLOB)
    except Exception as e:
        log.warning("Failed to upload session: %s", e)

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
log = logging.getLogger("userbot")

# Pyrogram client (MTProto user account)
app_client = Client(
    "rooted_userbot",
    api_id=int(os.getenv("API_ID", "0")),
    api_hash=os.getenv("API_HASH", ""),
    phone_number=os.getenv("PHONE_NUMBER", ""),
)

# Database connection (optional — userbot works without it)
def get_db():
    url = os.getenv("DATABASE_URL")
    if not url:
        return None
    try:
        return psycopg2.connect(url)
    except Exception as e:
        log.warning("DB connection failed (non-fatal): %s", e)
        return None

BACKEND_URL = os.getenv("BACKEND_URL", "http://localhost:8080")

# Track which groups we're managing (telegram_group_id → conversation_id)
managed_groups: dict[int, str] = {}


# ==========================================
# FASTAPI HTTP API (called by Go backend)
# ==========================================

@asynccontextmanager
async def lifespan(app: FastAPI):
    """Start Pyrogram client on startup, stop on shutdown."""
    download_session()
    await app_client.start()
    log.info("Userbot started: %s", (await app_client.get_me()).first_name)

    # Load managed groups from DB (skip if DB unavailable)
    db = get_db()
    if db:
        try:
            cur = db.cursor()
            cur.execute("SELECT telegram_group_id, id FROM conversations WHERE telegram_group_id IS NOT NULL")
            for row in cur.fetchall():
                managed_groups[row[0]] = row[1]
            cur.close()
            db.close()
            log.info("Loaded %d managed groups", len(managed_groups))
        except Exception as e:
            log.warning("Failed to load managed groups: %s", e)
    else:
        log.info("No DB — skipping managed groups load")

    yield

    await app_client.stop()  # releases SQLite lock first
    upload_session()  # persist session after clean shutdown
    log.info("Userbot stopped")


api = FastAPI(title="Rooted Userbot", lifespan=lifespan)


class CreateGroupRequest(BaseModel):
    conversation_id: str
    user_a_telegram_id: int
    user_b_telegram_id: int
    user_a_name: str
    user_b_name: str


class SendMessageRequest(BaseModel):
    telegram_group_id: int
    sender_name: str
    content: str
    content_type: str = "text"


@api.post("/create-group")
async def create_group(req: CreateGroupRequest):
    """Create a Telegram supergroup for a match and add both users."""
    try:
        # Create supergroup
        group_title = f"{req.user_a_name} & {req.user_b_name}"
        group = await app_client.create_supergroup(group_title)
        group_id = group.id

        log.info("Created group %d: %s", group_id, group_title)

        # Generate invite link (limited to 2 members)
        invite = await app_client.create_chat_invite_link(group_id, member_limit=2)
        invite_link = invite.invite_link
        log.info("Invite link for group %d: %s", group_id, invite_link)

        # Try direct add (works if userbot has interacted with user before)
        added = []
        for tid, name in [(req.user_a_telegram_id, req.user_a_name),
                          (req.user_b_telegram_id, req.user_b_name)]:
            try:
                await app_client.add_chat_members(group_id, tid)
                added.append(name)
                log.info("Added %s (%d) to group %d", name, tid, group_id)
            except Exception as e:
                log.info("Direct add failed for %s (%d), will use invite link: %s", name, tid, e)

        # Set group description
        try:
            await app_client.set_chat_description(
                group_id,
                f"Rooted match — chat privately. Your phone number is not shared."
            )
        except Exception:
            pass

        # Send welcome message
        await app_client.send_message(
            group_id,
            f"You've matched on Rooted! Start chatting.\n\n"
            f"Your phone number is NOT shared. Only your display name is visible.\n"
            f"Be respectful. Report issues via the Rooted app."
        )

        # Save to DB (skip if unavailable)
        db = get_db()
        if db:
            try:
                cur = db.cursor()
                cur.execute(
                    "UPDATE conversations SET telegram_group_id = %s, telegram_invite_link = %s WHERE id = %s",
                    (group_id, invite_link, req.conversation_id)
                )
                db.commit()
                cur.close()
                db.close()
            except Exception as e:
                log.error("Failed to save group to DB: %s", e)
        else:
            log.info("No DB — skipping group save (group_id=%d, conv=%s)", group_id, req.conversation_id)

        # Track this group
        managed_groups[group_id] = req.conversation_id

        return {
            "group_id": group_id,
            "invite_link": invite_link,
            "added": added,
        }

    except FloodWait as e:
        log.warning("FLOOD_WAIT: must wait %d seconds before creating groups", e.value)
        raise HTTPException(status_code=429, detail=f"rate_limited: retry after {e.value}s")
    except Exception as e:
        log.error("Failed to create group: %s", e)
        raise HTTPException(status_code=500, detail=str(e))


class LeaveGroupRequest(BaseModel):
    telegram_group_id: int


@api.post("/leave-group")
async def leave_group(req: LeaveGroupRequest):
    """Leave a Telegram group to free up a slot (500 group cap)."""
    try:
        await app_client.leave_chat(req.telegram_group_id)
        managed_groups.pop(req.telegram_group_id, None)
        log.info("Left group %d", req.telegram_group_id)
        return {"status": "left"}
    except Exception as e:
        log.warning("Failed to leave group %d: %s", req.telegram_group_id, e)
        return {"status": "already_left"}


@api.post("/send-message")
async def send_message(req: SendMessageRequest):
    """Send a message from Mini App to the Telegram group."""
    try:
        if req.content_type == "text":
            await app_client.send_message(
                req.telegram_group_id,
                f"<b>{req.sender_name}:</b> {req.content}",
                parse_mode="html",
            )
        return {"status": "sent"}
    except Exception as e:
        log.error("Failed to send message to group %d: %s", req.telegram_group_id, e)
        raise HTTPException(status_code=500, detail=str(e))


@api.get("/health")
async def health():
    me = await app_client.get_me()
    return {"status": "ok", "user": me.first_name, "groups_managed": len(managed_groups)}


# ==========================================
# MESSAGE LISTENER (Telegram → DB)
# ==========================================

@app_client.on_message(filters.group & ~filters.me)
async def on_group_message(client: Client, message: Message):
    """Capture messages from managed Telegram groups and save to DB."""
    group_id = message.chat.id

    if group_id not in managed_groups:
        return  # Not our group

    conversation_id = managed_groups[group_id]
    sender_telegram_id = message.from_user.id if message.from_user else None

    if not sender_telegram_id:
        return

    # Determine content
    content_type = "text"
    content = message.text or ""

    if message.photo:
        content_type = "photo"
        content = message.photo.file_id
    elif message.voice:
        content_type = "voice_note"
        content = message.voice.file_id

    if not content:
        return

    # Look up internal user_id from telegram_id and save message
    db = get_db()
    if not db:
        log.info("No DB — skipping message save from %d in conv %s", sender_telegram_id, conversation_id[:8])
        return

    try:
        cur = db.cursor()

        cur.execute("SELECT id FROM users WHERE telegram_id = %s", (sender_telegram_id,))
        row = cur.fetchone()
        if not row:
            cur.close()
            db.close()
            return

        user_id = row[0]

        # Save message
        cur.execute(
            """INSERT INTO messages (conversation_id, sender_id, content_type, content, source, synced_to_telegram)
               VALUES (%s, %s, %s, %s, 'telegram', TRUE)""",
            (conversation_id, user_id, content_type, content)
        )

        # Update conversation metadata
        cur.execute(
            """UPDATE conversations SET last_message_at = NOW(), message_count = message_count + 1
               WHERE id = %s""",
            (conversation_id,)
        )

        db.commit()
        cur.close()
        db.close()

        log.info("Saved message from %d in conv %s", sender_telegram_id, conversation_id[:8])

    except Exception as e:
        log.error("Failed to save group message: %s", e)


# ==========================================
# ENTRY POINT
# ==========================================

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(api, host="0.0.0.0", port=8090)
