#!/bin/bash
# Rooted Bot Setup Script
# Run this after deploying the API to configure the Telegram bot.

set -e

if [ -z "$TELEGRAM_BOT_TOKEN" ]; then
  echo "Error: TELEGRAM_BOT_TOKEN environment variable is required"
  echo "Get it from @BotFather on Telegram"
  exit 1
fi

if [ -z "$API_URL" ]; then
  echo "Error: API_URL environment variable is required (e.g. https://rooted-api-xxx.run.app)"
  exit 1
fi

if [ -z "$MINIAPP_URL" ]; then
  echo "Error: MINIAPP_URL environment variable is required (e.g. https://rooted-miniapp.pages.dev)"
  exit 1
fi

API="https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN"

echo "=== Setting webhook ==="
curl -s -X POST "$API/setWebhook" \
  -H "Content-Type: application/json" \
  -d "{
    \"url\": \"$API_URL/webhook/telegram\",
    \"allowed_updates\": [\"message\", \"callback_query\", \"pre_checkout_query\"]
  }" | python3 -m json.tool

echo ""
echo "=== Setting bot commands ==="
curl -s -X POST "$API/setMyCommands" \
  -H "Content-Type: application/json" \
  -d '{
    "commands": [
      {"command": "start", "description": "Start Rooted / Open profile"},
      {"command": "profile", "description": "View or edit your profile"},
      {"command": "matches", "description": "View your matches"},
      {"command": "explore", "description": "Browse profiles"},
      {"command": "chat", "description": "Switch conversation"},
      {"command": "settings", "description": "Manage settings"},
      {"command": "help", "description": "Help and FAQ"}
    ]
  }' | python3 -m json.tool

echo ""
echo "=== Setting menu button (Mini App) ==="
curl -s -X POST "$API/setChatMenuButton" \
  -H "Content-Type: application/json" \
  -d "{
    \"menu_button\": {
      \"type\": \"web_app\",
      \"text\": \"Open Rooted\",
      \"web_app\": {\"url\": \"$MINIAPP_URL\"}
    }
  }" | python3 -m json.tool

echo ""
echo "=== Setting bot description ==="
curl -s -X POST "$API/setMyDescription" \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Rooted — Dating, rooted in what matters. A cross-cultural dating app for the African diaspora. Find meaningful connections across continents."
  }' | python3 -m json.tool

echo ""
echo "=== Setting short description ==="
curl -s -X POST "$API/setMyShortDescription" \
  -H "Content-Type: application/json" \
  -d '{
    "short_description": "Cross-cultural dating for the African diaspora"
  }' | python3 -m json.tool

echo ""
echo "=== Webhook info ==="
curl -s "$API/getWebhookInfo" | python3 -m json.tool

echo ""
echo "=== Done! ==="
echo "Bot is configured. Users can now:"
echo "  1. Search for your bot on Telegram"
echo "  2. Tap /start"
echo "  3. Create their profile via the Mini App"
echo "  4. Start matching!"
