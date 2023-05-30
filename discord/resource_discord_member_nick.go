package discord

import (
	"log"
	"strings"

	"github.com/andersfylling/snowflake/v5"
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

	client := m.(*Context).Client

	serverId := getId(d.Get("server_id").(string))
	userId := getId(d.Get("user_id").(string))

	if _, err := client.Guild(serverId).Member(userId).Get(); err != nil {
		return diag.Errorf("Could not get member %s in %s: %s", userId.String(), serverId.String(), err.Error())
	}

	d.SetId(generateTwoPartId(serverId.String(), userId.String()))

	diags = append(diags, resourceMemberNickRead(ctx, d, m)...)
	diags = append(diags, resourceMemberNickUpdate(ctx, d, m)...)

	return diags
}

func resourceMemberNickRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Client

	// parse server ID and userID out of the ID:
	var serverId, userId snowflake.Snowflake
	sId, uId, err := parseTwoIds(d.Id())
	if err != nil {
		log.Default().Printf("Unable to parse IDs out of the resource ID. Falling back on legacy config behavior.")
		serverId = getId(d.Get("server_id").(string))
		userId = getId(d.Get("user_id").(string))
	} else {
		serverId = getId(sId)
		userId = getId(uId)
	}

	member, err := client.Guild(serverId).Member(userId).Get()
	if err != nil {
		// If error string contains "Unknown Member", it's because the member is not in the server.
		// This is not an error, so we just return an empty ID.
		if strings.Contains(err.Error(), "Unknown Member") {
			log.Default().Printf("Member %s not found in server %s. Removing from state.", userId.String(), serverId.String())
			d.SetId("")
			return nil
		}
		return diag.Errorf("Could not get member %s in %s: %s", userId.String(), serverId.String(), err.Error())
	}

	d.Set("nick", member.Nick)

	return diags
}

func resourceMemberNickUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Client

	serverId := getId(d.Get("server_id").(string))
	userId := getId(d.Get("user_id").(string))

	member, err := client.Guild(serverId).Member(userId).Get()
	if err != nil {
		return diag.Errorf("Could not get member %s in %s: %s", userId.String(), serverId.String(), err.Error())
	}

	old, new := d.GetChange("nick")
	if old == new {
		return diags
	}

	if err := member.UpdateNick(ctx, client, new.(string)); err != nil {
		return diag.Errorf("Failed to update member nickname from %s to %s! Member %s. Error: %s", old, new, userId.String(), err.Error())
	}

	return diags
}

func resourceMemberNickDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Client
	serverId := getId(d.Get("server_id").(string))
	userId := getId(d.Get("user_id").(string))

	member, err := client.Guild(serverId).Member(userId).Get()
	if err != nil {
		return diag.Errorf("Could not get member %s in %s: %s", userId.String(), serverId.String(), err.Error())
	}

	if err := member.UpdateNick(ctx, client, ""); err != nil {
		return diag.Errorf("Failed to delete member nickname %s! Member %s. Error: %s", member.Nick, userId.String(), err.Error())
	}

	return diags
}
