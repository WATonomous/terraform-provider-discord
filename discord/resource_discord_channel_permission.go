package discord

import (
	"log"

	"github.com/andersfylling/disgord"
	"github.com/andersfylling/snowflake/v5"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"golang.org/x/net/context"
)

func resourceDiscordChannelPermission() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceChannelPermissionCreate,
		ReadContext:   resourceChannelPermissionRead,
		UpdateContext: resourceChannelPermissionUpdate,
		DeleteContext: resourceChannelPermissionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"channel_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"type": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
				ValidateDiagFunc: func(val interface{}, path cty.Path) (diags diag.Diagnostics) {
					v := val.(string)

					if v != "role" && v != "user" {
						diags = append(diags, diag.Errorf("%s is not a valid type. Must be \"role\" or \"user\"", v)...)
					}

					return diags
				},
			},
			"overwrite_id": {
				ForceNew: true,
				Required: true,
				Type:     schema.TypeString,
			},
			"allow": {
				AtLeastOneOf: []string{"allow", "deny"},
				Optional:     true,
				Type:         schema.TypeInt,
			},
			"deny": {
				AtLeastOneOf: []string{"allow", "deny"},
				Optional:     true,
				Type:         schema.TypeInt,
			},
		},
	}
}

func resourceChannelPermissionCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Client

	channelId := getId(d.Get("channel_id").(string))
	overwriteId := getId(d.Get("overwrite_id").(string))
	permissionType, _ := getDiscordChannelPermissionType(d.Get("type").(string))

	if err := client.Channel(channelId).UpdatePermissions(overwriteId, &disgord.UpdateChannelPermissions{
		Allow: disgord.PermissionBit(d.Get("allow").(int)),
		Deny:  disgord.PermissionBit(d.Get("deny").(int)),
		Type:  permissionType,
	}); err != nil {
		return diag.Errorf("Failed to update channel permissions %s: %s", channelId.String(), err.Error())
	} else {
		d.SetId(generateThreePartId(channelId.String(), overwriteId.String(), d.Get("type").(string)))

		return diags
	}
}

func resourceChannelPermissionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Client

	var channelId, overwriteId snowflake.Snowflake
	var permissionType uint

	cId, oId, pt, err := parseThreeIds(d.Id())
	if err != nil {
		log.Default().Printf("Unable to parse IDs out of the resource ID. Falling back on legacy config behavior.")
		channelId = getId(d.Get("channel_id").(string))
		overwriteId = getId(d.Get("overwrite_id").(string))
		permissionType, _ = getDiscordChannelPermissionType(d.Get("type").(string))
	} else {
		channelId = getId(cId)
		overwriteId = getId(oId)
		permissionType, _ = getDiscordChannelPermissionType(pt)

		d.Set("channel_id", channelId.String())
		d.Set("overwrite_id", overwriteId.String())
		d.Set("type", pt)
	}

	channel, err := client.Channel(channelId).Get()
	if err != nil {
		return diag.Errorf("Failed to find channel %s: %s", channelId.String(), err.Error())
	}

	for _, x := range channel.PermissionOverwrites {
		if uint(x.Type) == uint(permissionType) && x.ID == overwriteId {
			d.Set("allow", int(x.Allow))
			d.Set("deny", int(x.Deny))
			break
		}
	}

	return diags
}

func resourceChannelPermissionUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Client

	channelId := getId(d.Get("channel_id").(string))
	overwriteId := getId(d.Get("overwrite_id").(string))
	permissionType, _ := getDiscordChannelPermissionType(d.Get("type").(string))

	if err := client.Channel(channelId).UpdatePermissions(overwriteId, &disgord.UpdateChannelPermissions{
		Allow: disgord.PermissionBit(d.Get("allow").(int)),
		Deny:  disgord.PermissionBit(d.Get("deny").(int)),
		Type:  uint(permissionType),
	}); err != nil {
		return diag.Errorf("Failed to update channel permissions %s: %s", channelId.String(), err.Error())
	} else {
		return diags
	}
}

func resourceChannelPermissionDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Client

	channelId := getId(d.Get("channel_id").(string))
	overwriteId := getId(d.Get("overwrite_id").(string))

	if err := client.Channel(channelId).DeletePermission(overwriteId); err != nil {
		return diag.Errorf("Failed to delete channel permissions %s: %s", channelId.String(), err.Error())
	} else {
		return diags
	}
}
