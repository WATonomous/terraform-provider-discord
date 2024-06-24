package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"golang.org/x/net/context"
)

func dataSourceDiscordMembers() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMembersRead,
		Description: "Fetches all members in a server.",

		Schema: map[string]*schema.Schema{
			"server_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The server ID to search for the user in.",
			},
			"members": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The members in the server.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"user_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The user's ID.",
						},
						"username": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The user's username.",
						},
						"discriminator": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The user's discriminator.",
						},
						"joined_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The time at which the user joined.",
						},
						"premium_since": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The time at which the user became premium.",
						},
						"avatar": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The avatar hash of the user.",
						},
						"nick": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The current nickname of the user.",
						},
						"roles": {
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Computed:    true,
							Set:         schema.HashString,
							Description: "IDs of the roles that the user has.",
						},
					},
				},
			},
		},
	}
}

func dataSourceMembersRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var members []*discordgo.Member
	client := m.(*Context).Session
	serverId := d.Get("server_id").(string)

	after := ""
	// Fetch all members, with pagination
	for {
		page, err := client.GuildMembers(serverId, after, 1000, discordgo.WithContext(ctx))
		if err != nil {
			return diag.Errorf("Failed to fetch members for %s: %s", serverId, err.Error())
		}
		members = append(members, page...)
		if len(page) < 1000 {
			break
		}
		after = page[len(page)-1].User.ID
	}

	memberList := make([]interface{}, 0, len(members))
	for _, member := range members {
		roles := make([]string, 0, len(member.Roles))
		for _, r := range member.Roles {
			roles = append(roles, r)
		}

		var discriminator string
		if member.User.Discriminator == "0" {
			// Use an empty string to indicate no discriminator
			discriminator = ""
		} else {
			discriminator = member.User.Discriminator
		}

		memberList = append(memberList, map[string]interface{}{
			"user_id":       member.User.ID,
			"username":      member.User.Username,
			"discriminator": discriminator,
			"joined_at":     member.JoinedAt.String(),
			"avatar":        member.User.Avatar,
			"nick":          member.Nick,
			"roles":         roles,
		})
	}

	d.SetId(serverId)
	d.Set("members", memberList)

	return diags
}
