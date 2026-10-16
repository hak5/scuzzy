# Scuzzy Bot Privacy Policy

Last updated: October 12, 2026

## About this policy

Scuzzy reads messages in the Hak5 Discord server only to moderate it, and it does not keep a database of what it reads. This policy explains what Scuzzy sees, why, and where that information goes.

Scuzzy is a moderation and utility bot run by Hak5 LLC ("Hak5", "we") for the official Hak5 Discord server. This policy covers only the bot. Hak5's websites, shop and other services are covered by the [Hak5 Privacy Policy](https://shop.hak5.org/pages/privacy), and conduct in the server is governed by the [Hak5 Community Guidelines](https://hak5.org/pages/community). Your use of Discord itself is covered by [Discord's Privacy Policy](https://discord.com/privacy).

Scuzzy works only in the Hak5 Discord server. It is not a public bot and cannot be added to other servers.

## Information Scuzzy processes

Scuzzy only sees what Discord sends it from the Hak5 server. It does not collect anything from outside Discord.

| Information | When Scuzzy sees it |
| --- | --- |
| Message text and attachment file names and sizes | Every message posted in a channel Scuzzy can read |
| Your Discord user ID, username and display name | With every message and event Scuzzy receives |
| Channel and message IDs, and when a message was sent | With every message |
| Your roles and nickname, and when you joined the server | When you or someone else runs a command such as `/userinfo` or `/color`, and to check whether you are staff |
| The text and author of a deleted message | When a recently sent message is deleted |
| The fact that you joined the server | When you join |
| The question you ask | When you use the `/askdocs` command |

Scuzzy does not read your direct messages with other people, your messages in other servers, voice channels, your email address, or payment information. It ignores direct messages sent to it.

## How Scuzzy uses it

Scuzzy uses this information only to run and moderate the Hak5 server:

- **Spam detection.** Scuzzy compares a one-way, keyed fingerprint of each new message with fingerprints of your other recent messages. It does not store the message text for this. If the same message is posted repeatedly in one channel, or across several channels within a short window (30 seconds by default), Scuzzy may delete the copies, time you out, and alert moderators. Staff messages are not checked.
- **Honeypot channel.** Posting in a designated trap channel may result in an automatic ban and removal of your recent messages. These actions are reported to moderators.
- **Commands.** Scuzzy responds to slash commands, such as setting a color role, joining a role, showing user or server info, or unit conversions.
- **Moderation tools.** Staff use Scuzzy to kick, ban, purge, and set slow mode. When staff move or copy a message to another channel, Discord sends Scuzzy that message's text and author so it can be reposted.
- **Deleted message log.** When a message is deleted, Scuzzy posts its author, channel and text to a staff-only log channel so moderators can review abuse.
- **Welcome.** When you join, Scuzzy sends you a welcome direct message and may give you default roles.
- **Docs answers.** The `/askdocs` command sends your question to the Hak5 documentation search, powered by GitBook, and posts the AI-generated answer in the channel.

Scuzzy does not use your information for advertising, profiling, or training AI models.

## Storage and retention

Scuzzy has no database. Most of what it sees is held briefly in memory and then discarded.

| Where | What | How long |
| --- | --- | --- |
| Bot memory | The most recent 1,024 messages, plus member and channel details | Replaced by newer messages, and cleared when the bot restarts |
| Bot memory | Spam-check fingerprints of your recent messages (keyed HMAC-SHA256 hashes, never the text) | About 1 minute after the detection window ends; the key is discarded on restart |
| Bot configuration file | User IDs that staff have told Scuzzy to ignore | Until staff remove them |
| Bot server activity log | Usernames of people who ran commands, user IDs involved in automatic moderation actions, and errors | Kept on the Hak5 bot server for troubleshooting |
| Staff-only Discord channels | Deleted-message logs and moderation reports, which can include message text | Until staff delete them, subject to Discord's own retention |

The bot runs on a server operated by Hak5. Information posted to Discord channels is also stored by Discord under [Discord's Privacy Policy](https://discord.com/privacy).

## Who we share it with

We do not sell or rent information that Scuzzy processes. It goes only to:

- **Hak5 moderators and administrators**, through staff-only channels, to review spam, abuse, and rule violations.
- **Discord**, which delivers every message and action Scuzzy sends or receives.
- **GitBook**, which receives the text of questions you ask with `/askdocs`, so it can search Hak5's documentation. Your Discord identity is not sent with the question.
- **Others where required by law**, or to protect the safety of the community, as described in the [Hak5 Privacy Policy](https://shop.hak5.org/pages/privacy).

## Your choices

- **Don't interact.** Scuzzy only reads messages in the Hak5 server. Leaving the server stops it from seeing anything new from you.
- **Delete your messages.** You can delete your own messages at any time. A copy may remain in the staff-only deleted-message log.
- **Ask us.** To request access to, or deletion of, information held in moderation logs, contact a Hak5 moderator in the server or use the contact details in the [Hak5 Privacy Policy](https://shop.hak5.org/pages/privacy). We may keep records needed to enforce the [Community Guidelines](https://hak5.org/pages/community) or to meet legal obligations.

## Children

Scuzzy is not directed at children. Discord requires users to be at least 13, or older where local law requires.

## Changes to this policy

We may update this policy when Scuzzy's features change. The date at the top shows when it was last updated.

## Contact

Questions about this policy can be sent to Hak5 using the contact details in the [Hak5 Privacy Policy](https://shop.hak5.org/pages/privacy).
