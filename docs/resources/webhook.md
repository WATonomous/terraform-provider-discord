# Discord Webhook Resource

A resource to create a webhook for a channel

## Example Usage

```hcl-terraform
data discord_local_image avatar {
    file = "avatar.png"
}

resource discord_webhook webhook {
    channel_id      = var.channel_id
    name            = "Bot"
    avatar_data_uri = data.discord_local_image.avatar.data_uri
}

output webhook-url {
    value = nonsensitive(discord_webhook.webhook.url)
}
```

## Argument Reference

* `channel_id` (Required) ID of the channel to create a webhook for
* `name` (Optional) Default name of the webhook
* `avatar_url` (Optional) Remote URL for setting the default avatar of the
  webhook
* `avatar_data_uri` (Optional) Data URI of an image to set as the default avatar
  of the webhook

## Attributes Reference

* `avatar_hash` Hash of the avatar
* `token` The webhook token
* `url` The webhook URL
* `slack_url` The Slack-compatible webhook URL
* `github_url` The GitHub-compatible webhook URL
