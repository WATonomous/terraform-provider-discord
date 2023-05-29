package discord

import (
	"context"

	"github.com/andersfylling/disgord"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDiscordChannel() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDiscordChannelRead,
		Schema: map[string]*schema.Schema{
			"server_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"channel_id": {
				ExactlyOneOf: []string{"channel_id", "name"},
				Type:         schema.TypeString,
				Optional:     true,
			},
			"name": {
				ExactlyOneOf: []string{"channel_id", "name"},
				Type:         schema.TypeString,
				Optional:     true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"position": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"category": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sync_perms_with_category": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceDiscordChannelRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error
	var channel *disgord.Channel
	client := m.(*Context).Client

	if v, ok := d.GetOk("channel_id"); ok {
		channel, err = client.Channel(getId(v.(string))).Get()
		if err != nil {
			return diag.Errorf("Failed to fetch channel %s: %s", v.(string), err.Error())
		}
	}
	if v, ok := d.GetOk("name"); ok {
		channels, err := client.Guild(getId(d.Get("server_id").(string))).GetChannels()
		if err != nil {
			return diag.Errorf("Failed to fetch channels for server %s: %s", d.Get("server_id").(string), err.Error())
		}
		for _, ch := range channels {
			// Print the channel name and ID for debugging
			if ch.Name == v.(string) {
				channel = ch
				break
			}
		}
		if channel == nil {
			return diag.Errorf("Failed to find channel with name %s", v.(string))
		}
	}

	channelType, ok := getTextChannelType(channel.Type)
	if !ok {
		return diag.Errorf("Invalid channel type: %d", channel.Type)
	}

	d.SetId(channel.ID.String())
	d.Set("server_id", channel.GuildID.String())
	d.Set("channel_id", channel.ID.String())
	d.Set("type", channelType)
	d.Set("name", channel.Name)
	d.Set("position", channel.Position)

	if channelType != "category" {
		if channel.ParentID.IsZero() {
			d.Set("category", nil)
		} else {
			d.Set("category", channel.ParentID.String())
		}

		parent, err := client.Channel(channel.ParentID).Get()
		if err != nil {
			return diag.Errorf("Failed to fetch category of channel %s: %s", channel.ID.String(), err.Error())
		}

		synced := arePermissionsSynced(channel, parent)
		d.Set("sync_perms_with_category", synced)
	}

	return diags
}
