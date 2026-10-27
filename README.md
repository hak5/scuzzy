# Scuzzy Discord Bot

Scuzzy is a simple Discord bot written in Golang. It uses [discordgo](https://github.com/bwmarrin/discordgo).

## Bot Configuration
A sample bot configuration is provided in the `cmd` directory.

Commands are Discord slash commands and are registered with the server on startup. Set `message_content_intent` to `true` only after Discord has approved the Message Content intent for the app; spam detection and deleted message logs need it to see message text.

## License
This project is licensed under the BSD-3-Clause.

## Moderation:
- Duplicate and cross-channel spam detection. Scam and raid accounts often post the same message over and over in one channel, or blast it across many channels within seconds. Scuzzy compares each message with the same user's messages from the last 30 seconds. If one message repeats past a threshold, Scuzzy deletes every copy, times the user out, and alerts moderators in a private channel so they can decide whether to ban. To do this, it has to read message text and attachment names. It doesn't keep that text: it stores only a one-way keyed hash (HMAC-SHA256) of each message, for about 30 seconds, in memory only. The hash key is random each time the bot starts and is never saved. Staff messages are exempt.
- Honeypot channel. A trap channel that real members are told not to post in. Accounts that post there anyway, almost always automated spam bots, are banned automatically and reported to moderators.
- Deleted message log. When a message is deleted, Scuzzy posts its author, channel and text to a staff-only log channel. This lets moderators review harassment or scams that were deleted to hide them. Before Scuzzy deletes anything itself (spam, honeypot bans, `/purge`, Move message), it copies the messages to the moderator channel in the same format as Copy message.
- Move or copy a message. A moderator can right-click an off-topic message (Apps > Move message / Copy message) and pick the right channel; Scuzzy reposts it with its original text and author.
- Staff tools. Kick, ban, purge, slow mode, and an ignore list for abusive bot users.

## Community utilities:
- Self-assigned name color roles and opt-in roles, such as product beta groups.
- A welcome direct message for new members, with links to the rules.
- User and server info lookups.
- Quick links to Hak5 documentation, GitHub, payloads, the shop and support.
- A docs search that answers product questions from the Hak5 documentation.
- Small helpers such as temperature and unit conversions.
- A YouTube feed. Scuzzy checks a YouTube channel's public feed every few minutes and posts new videos, Shorts and live streams to a feed channel. Set `youtube_channel_id` (the channel's `UC...` ID) and `youtube_feed_channel`. Live stream detection needs a YouTube Data API key in the `SCUZZY_YOUTUBE_API_KEY` environment variable; without one, streams are posted as videos. Posted video IDs are saved to `youtube_seen.json` next to the config so restarts don't repost them.

## Data handling:
- Scuzzy has no database.
- Apart from the spam hashes above, the only message data it keeps is Discord's standard in-memory cache of recent messages, which it uses for the deleted-message log and which is cleared on restart.
- It doesn't sell, share or analyze user data, or use it for advertising, profiling or AI training.
