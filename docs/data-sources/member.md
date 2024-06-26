---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "discord_member Data Source - discord"
subcategory: ""
description: |-
  Fetches a member's information from a server.
---

# discord_member (Data Source)

Fetches a member's information from a server.

## Example Usage

```terraform
data "discord_member" "jake" {
  server_id = "81384788765712384"
  user_id   = "103559217914318848"
}

output "jakes_username_and_discrim" {
  value = "${data.discord_member.jake.username}#${data.discord_member.jake.discriminator}"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `server_id` (String) The server ID to search for the user in.

### Optional

- `discriminator` (String, Deprecated) The discriminator to search for. `username` is required when using this.
- `user_id` (String) The user ID to search for. Required if not searching by `username` / `discriminator`.
- `username` (String) The username to search for.

### Read-Only

- `avatar` (String) The avatar hash of the user.
- `id` (String) The user's ID.
- `in_server` (Boolean) Whether the user is in the server.
- `joined_at` (String) The time at which the user joined.
- `nick` (String) The current nickname of the user.
- `premium_since` (String) The time at which the user became premium.
- `roles` (Set of String) IDs of the roles that the user has.
