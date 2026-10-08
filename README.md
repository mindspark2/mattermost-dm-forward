# Forward a direct message to another user

Mattermost **Team Edition** plugin that forwards a direct-message or group-message post, including images and other files, to one or more people. Each person receives a **separate direct message**. The original conversation gets a short thread note listing who received it.

Plugin id: `com.mst.forward-private`.

This uses the standard plugin API. It does not need an Enterprise or Professional license.

## What it does

- Source must be a **direct message** or **group message**. Channel posts are rejected.
- Hover the message and open **Message actions** (the apps icon). Choose **Forward to…**.
- The built-in **Forward** item in the ⋯ menu is unchanged. On Team Edition it only re-posts inside the same conversation.
- Optional comment, then up to 8 recipients (System Console setting **Max recipients per forward**).
- Each recipient gets their own DM with the original text, the comment, and copies of the attachments.
- The original message gets a thread reply such as `@you forwarded this message to @alex and @sam.`
- Fallback: `/forwardprivate` plus a permalink to the message.

## Requirements

- Mattermost Team Edition **11.0** or newer (tested on 11.11).
- Docker, to build the server binary and webapp bundle. Go and Node are not required on the host.

## Build

```bash
chmod +x scripts/docker-build.sh
./scripts/docker-build.sh
```

The tarball is `dist/com.mst.forward-private.tar.gz`.

## Install

System Console → Plugins → Upload plugin, then enable **MST Forward Private**.

Or with mmctl:

```bash
mmctl plugin add --force dist/com.mst.forward-private.tar.gz
mmctl plugin enable com.mst.forward-private
```

Health check:

```text
GET /plugins/com.mst.forward-private/api/v1/health
```

After enabling or upgrading, hard-refresh the browser. Desktop apps need **Clear Cache and Reload** before **Message actions** shows **Forward to…**.

## Settings

| Setting | Default | Purpose |
|---|---|---|
| Enabled | on | Show the forward action |
| Max recipients per forward | 8 | Cap of 1–8 people |
| Show attribution on forwarded copies | on | Footer that the copy came from a private conversation |
| Receipt in source conversation | on | Thread note on the original message |
