package discord

import (
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"golang.org/x/net/context"
)

func resourceDiscordMemberNick() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMemberNickCreate,
		ReadContext:   resourceMemberNickRead,
		UpdateContext: resourceMemberNickUpdate,
		DeleteContext: resourceMemberNickDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Description: "A resource to manage member nicknames for a server.",
		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"server_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"nick": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceMemberNickCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := m.(*Context).Session

	serverId := d.Get("server_id").(string)
	userId := d.Get("user_id").(string)

	if _, err := client.GuildMember(serverId, userId, discordgo.WithContext(ctx)); err != nil {
		return diag.Errorf("Could not get member %s in %s: %s", userId, serverId, err.Error())
	}

	d.SetId(generateTwoPartId(serverId, userId))

	diags = append(diags, resourceMemberNickRead(ctx, d, m)...)
	diags = append(diags, resourceMemberNickUpdate(ctx, d, m)...)

	return diags
}

func resourceMemberNickRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Session

	// parse server ID and userID out of the ID:
	var serverId, userId string
	sId, uId, err := parseTwoIds(d.Id())
	if err != nil {
		log.Default().Printf("Unable to parse IDs out of the resource ID. Falling back on legacy config behavior.")
		serverId = d.Get("server_id").(string)
		userId = d.Get("user_id").(string)
	} else {
		serverId = sId
		userId = uId
	}

	member, err := client.GuildMember(serverId, userId, discordgo.WithContext(ctx))
	if err != nil {
		// If error string contains "Unknown Member", it's because the member is not in the server.
		// This is not an error, so we just return an empty ID.
		if strings.Contains(err.Error(), "Unknown Member") {
			log.Default().Printf("Member %s not found in server %s. Removing from state.", userId, serverId)
			d.SetId("")
			return nil
		}

		return diag.Errorf("Could not get member %s in %s: %s", userId, serverId, err.Error())
	}

	d.Set("nick", member.Nick)

	return diags
}

func resourceMemberNickUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Session

	serverId := d.Get("server_id").(string)
	userId := d.Get("user_id").(string)

	member, err := client.GuildMember(serverId, userId, discordgo.WithContext(ctx))
	if err != nil {
		return diag.Errorf("Could not get member %s in %s: %s", userId, serverId, err.Error())
	}

	nick := member.Nick
	_, newNick := d.GetChange("nick")

	if nick == newNick.(string) {
		return diags
	}

	if _, err := client.GuildMemberEdit(serverId, userId, &discordgo.GuildMemberParams{
		Nick: newNick.(string),
	}, discordgo.WithContext(ctx)); err != nil {
		return diag.Errorf("Failed to edit member %s: %s", userId, err.Error())
	}

	return diags
}

func resourceMemberNickDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Session
	serverId := d.Get("server_id").(string)
	userId := d.Get("user_id").(string)

	_, err := client.GuildMember(serverId, userId, discordgo.WithContext(ctx))
	if err != nil {
		return diag.Errorf("Could not get member %s in %s: %s", userId, serverId, err.Error())
	}

	if _, err := client.GuildMemberEdit(serverId, userId, &discordgo.GuildMemberParams{
		Nick: "",
	}, discordgo.WithContext(ctx)); err != nil {
		return diag.Errorf("Failed to delete member roles %s: %s", userId, err.Error())
	}

	return diags
}
