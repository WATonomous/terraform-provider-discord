package discord

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDiscordServer() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDiscordServerRead,
		Description: "Fetches a server's information.",
		Schema: map[string]*schema.Schema{
			"server_id": {
				ExactlyOneOf: []string{"server_id", "name"},
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The server ID to search for.",
			},
			"name": {
				ExactlyOneOf: []string{"server_id", "name"},
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The server name to search for.",
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the server.",
			},
			"region": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The region of the server.",
			},
			"default_message_notifications": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The default message notification level of the server.",
			},
			"verification_level": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The required verification level of the server.",
			},
			"explicit_content_filter": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The explicit content filter level of the server.",
			},
			"afk_timeout": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The AFK timeout of the server.",
			},
			"icon_hash": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The hash of the server icon.",
			},
			"splash_hash": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The hash of the server splash.",
			},
			"afk_channel_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The AFK channel ID.",
			},
			"owner_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the owner.",
			},
		},
	}
}

func dataSourceDiscordServerRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error
	var server *discordgo.Guild
	client := m.(*Context).Session

	if v, ok := d.GetOk("server_id"); ok {
		server, err = client.Guild(v.(string), discordgo.WithContext(ctx))
		if err != nil {
			return diag.Errorf("Failed to fetch server %s: %s", v.(string), err.Error())
		}
	}
	if v, ok := d.GetOk("name"); ok {
		// Discord API supports max 200 guilds per request
		guilds, err := client.UserGuilds(200, "", "", false, discordgo.WithContext(ctx))
		if err != nil {
			return diag.Errorf("Failed to fetch server list: %s", err.Error())
		}

		for _, s := range guilds {
			if s.Name == v.(string) {
				server, err = client.Guild(s.ID, discordgo.WithContext(ctx))
				if err != nil {
					return diag.Errorf("Failed to fetch server %s: %s", v.(string), err.Error())
				}
				break
			}
		}

		if server == nil {
			return diag.Errorf("Failed to fetch server %s (nil response)", v.(string))
		}
	}

	d.SetId(server.ID)
	d.Set("server_id", server.ID)
	d.Set("name", server.Name)
	d.Set("region", server.Region)
	d.Set("afk_timeout", server.AfkTimeout)
	d.Set("icon_hash", server.Icon)
	d.Set("splash_hash", server.Splash)
	d.Set("default_message_notifications", int(server.DefaultMessageNotifications))
	d.Set("verification_level", int(server.VerificationLevel))
	d.Set("explicit_content_filter", int(server.ExplicitContentFilter))

	if server.AfkChannelID != "" {
		d.Set("afk_channel_id", server.AfkChannelID)
	}
	if server.OwnerID != "" {
		d.Set("owner_id", server.OwnerID)
	}

	return diags
}
