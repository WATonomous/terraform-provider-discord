---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "discord_color Data Source - discord"
subcategory: ""
description: |-
  A simple helper to get the integer representation of a hex or RGB color.
---

# discord_color (Data Source)

A simple helper to get the integer representation of a hex or RGB color.

## Example Usage

```terraform
data "discord_color" "blue" {
  hex = "#4287f5"
}

data "discord_color" "green" {
  rgb = "rgb(46, 204, 113)"
}

resource "discord_role" "blue" {
  // ...
  color = data.discord_color.blue.dec
}

resource "discord_role" "green" {
  // ...
  color = data.discord_color.green.dec
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `hex` (String) The hex color code. Either this or `rgb` is required.
- `rgb` (String) The RGB color, in format: `rgb(R, G, B)`. Either this or `hex` is required.

### Read-Only

- `dec` (Number) The integer representation of the passed color.
- `id` (String) The ID of this resource.
