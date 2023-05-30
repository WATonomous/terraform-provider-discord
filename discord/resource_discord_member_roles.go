package discord

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/andersfylling/disgord"
	"github.com/andersfylling/snowflake/v5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"golang.org/x/net/context"
)

type RoleSchema struct {
	RoleId  disgord.Snowflake `json:"role_id"`
	HasRole bool              `json:"has_role"`
}

func convertToRoleSchema(v interface{}) (*RoleSchema, error) {
	var roleSchema *RoleSchema
	j, _ := json.MarshalIndent(v, "", "    ")
	err := json.Unmarshal(j, &roleSchema)

	return roleSchema, err
}

func resourceDiscordMemberRoles() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMemberRolesCreate,
		ReadContext:   resourceMemberRolesRead,
		UpdateContext: resourceMemberRolesUpdate,
		DeleteContext: resourceMemberRolesDelete,
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
			"role": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"role_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"has_role": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
					},
				},
			},
		},
	}
}

func resourceMemberRolesCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := m.(*Context).Client

	serverId := getId(d.Get("server_id").(string))
	userId := getId(d.Get("user_id").(string))

	if _, err := client.Guild(serverId).Member(userId).Get(); err != nil {
		return diag.Errorf("Could not get member %s in %s: %s", userId.String(), serverId.String(), err.Error())
	}

	d.SetId(generateTwoPartId(serverId.String(), userId.String()))

	diags = append(diags, resourceMemberRolesRead(ctx, d, m)...)
	diags = append(diags, resourceMemberRolesUpdate(ctx, d, m)...)

	return diags
}

func resourceMemberRolesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

	items := d.Get("role").(*schema.Set).List()
	roles := make([]interface{}, 0, len(items))

	for _, r := range items {
		v, _ := convertToRoleSchema(r)

		roles = append(roles, map[string]interface{}{
			"role_id":  v.RoleId.String(),
			"has_role": hasRole(member, v.RoleId),
		})
	}

	if err := d.Set("role", roles); err != nil {
		return diag.Errorf("Could not set roles: %s", err.Error())
	}

	return diags
}

func resourceMemberRolesUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Client

	serverId := getId(d.Get("server_id").(string))
	userId := getId(d.Get("user_id").(string))

	member, err := client.Guild(serverId).Member(userId).Get()
	if err != nil {
		return diag.Errorf("Could not get member %s in %s: %s", userId.String(), serverId.String(), err.Error())
	}

	old, new := d.GetChange("role")
	oldItems := old.(*schema.Set).List()
	items := new.(*schema.Set).List()

	roles := member.Roles

	for _, r := range items {
		v, _ := convertToRoleSchema(r)
		hasRole := hasRole(member, v.RoleId)
		// If it's supposed to have the role, and it does, continue
		if hasRole && v.HasRole {
			continue
		}
		// If it's supposed to have the role, and it doesn't, add it
		if v.HasRole && !hasRole {
			roles = append(roles, v.RoleId)
		}
		// If it's not supposed to have the role, and it does, remove it
		if !v.HasRole && hasRole {
			roles = removeRoleById(roles, v.RoleId)
		}
	}

	// If the change removed the role, and the user has it, remove it
	for _, r := range oldItems {
		v, _ := convertToRoleSchema(r)
		if wasRemoved(items, v) && v.HasRole {
			roles = removeRoleById(roles, v.RoleId)
		}
	}

	if _, err := client.Guild(serverId).Member(userId).Update(&disgord.UpdateMember{
		Roles: &roles,
	}); err != nil {
		return diag.Errorf("Failed to edit member %s: %s", userId.String(), err.Error())
	}

	return diags
}

func wasRemoved(items []interface{}, v *RoleSchema) bool {
	for _, i := range items {
		item, _ := convertToRoleSchema(i)
		if item.RoleId == v.RoleId {
			return false
		}
	}

	return true
}

func resourceMemberRolesDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Client
	serverId := getId(d.Get("server_id").(string))
	userId := getId(d.Get("user_id").(string))

	member, err := client.Guild(serverId).Member(userId).Get()
	if err != nil {
		return diag.Errorf("Could not get member %s in %s: %s", userId.String(), serverId.String(), err.Error())
	}

	items := d.Get("role").(*schema.Set).List()
	roles := member.Roles

	for _, r := range items {
		v, _ := convertToRoleSchema(r)
		hasRole := hasRole(member, v.RoleId)
		// if it's supposed to have the role, and it does, remove it
		if hasRole && v.HasRole {
			roles = removeRoleById(roles, v.RoleId)
		}
	}

	client.Guild(serverId).Member(userId).Update(&disgord.UpdateMember{
		Roles: &roles,
	})

	return diags
}
