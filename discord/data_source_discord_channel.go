package discord

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDiscordChannel() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDiscordChannelRead,
		Description: "Fetches a channel's information.",
		Schema: map[string]*schema.Schema{
			"server_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of server this channel is in.",
			},
			"channel_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ID of the channel.",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the channel.",
			},
			"position": {
				Type:        schema.TypeInt,
				Default:     1,
				Optional:    true,
				Description: "Position of the channel, `0`-indexed.",
				ValidateFunc: func(val interface{}, key string) (warns []string, errors []error) {
					v := val.(int)

					if v < 0 {
						errors = append(errors, fmt.Errorf("position must be greater than 0, got: %d", v))
					}

					return
				},
			},
		},
	}
}

func dataSourceDiscordChannelRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Session

	serverID := d.Get("server_id").(string)
	channelID := d.Get("channel_id").(string)
	channelName := d.Get("name").(string)

	var channel *discordgo.Channel

	if channelID != "" {
		channel, _ = client.Channel(channelID)
	} else if channelName != "" {
		channels, _ := client.GuildChannels(serverID)
		for _, c := range channels {
			if c.Name == channelName {
				channel = c
				break
			}
		}
	} else {
		return diag.Errorf("Either channel_id or channel name must be provided")
	}

	if channel == nil {
		return diag.Errorf("Channel with ID %s or name %s not found", channelID, channelName)
	}

	d.SetId(channel.ID)
	d.Set("server_id", channel.GuildID)
	d.Set("channel_id", channel.ID)
	d.Set("name", channel.Name)
	d.Set("position", channel.Position)

	return diags
}
