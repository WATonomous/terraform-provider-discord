package discord

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var permissions map[string]int64

// Reference: https://discord.com/developers/docs/topics/permissions
func dataSourceDiscordPermission() *schema.Resource {
	permissions = map[string]int64{
		"create_instant_invite":               0x0000000000000001,
		"kick_members":                        0x0000000000000002,
		"ban_members":                         0x0000000000000004,
		"administrator":                       0x0000000000000008,
		"manage_channels":                     0x0000000000000010,
		"manage_guild":                        0x0000000000000020,
		"add_reactions":                       0x0000000000000040,
		"view_audit_log":                      0x0000000000000080,
		"priority_speaker":                    0x0000000000000100,
		"stream":                              0x0000000000000200,
		"view_channel":                        0x0000000000000400,
		"send_messages":                       0x0000000000000800,
		"send_tts_messages":                   0x0000000000001000,
		"manage_messages":                     0x0000000000002000,
		"embed_links":                         0x0000000000004000,
		"attach_files":                        0x0000000000008000,
		"read_message_history":                0x0000000000010000,
		"mention_everyone":                    0x0000000000020000,
		"use_external_emojis":                 0x0000000000040000,
		"view_guild_insights":                 0x0000000000080000,
		"connect":                             0x0000000000100000,
		"speak":                               0x0000000000200000,
		"mute_members":                        0x0000000000400000,
		"deafen_members":                      0x0000000000800000,
		"move_members":                        0x0000000001000000,
		"use_vad":                             0x0000000002000000,
		"change_nickname":                     0x0000000004000000,
		"manage_nicknames":                    0x0000000008000000,
		"manage_roles":                        0x0000000010000000,
		"manage_webhooks":                     0x0000000020000000,
		"manage_guild_expressions":            0x0000000040000000,
		"use_application_commands":            0x0000000080000000,
		"request_to_speak":                    0x0000000100000000,
		"manage_events":                       0x0000000200000000,
		"manage_threads":                      0x0000000400000000,
		"create_public_threads":               0x0000000800000000,
		"create_private_threads":              0x0000001000000000,
		"use_external_stickers":               0x0000002000000000,
		"send_messages_in_threads":            0x0000004000000000,
		"use_embedded_activities":             0x0000008000000000,
		"moderate_members":                    0x0000010000000000,
		"view_creator_monetization_analytics": 0x0000020000000000,
		"use_soundboard":                      0x0000040000000000,
		"create_guild_expressions":            0x0000080000000000,
		"create_events":                       0x0000100000000000,
		"use_external_sounds":                 0x0000200000000000,
		"send_voice_messages":                 0x0000400000000000,
		"send_polls":                          0x0002000000000000,
		"use_external_apps":                   0x0004000000000000,
	}

	schemaMap := make(map[string]*schema.Schema)
	schemaMap["allow_extends"] = &schema.Schema{
		Type:        schema.TypeInt,
		Optional:    true,
		Description: "The base permission bits for allow to extend.",
	}
	schemaMap["deny_extends"] = &schema.Schema{
		Type:        schema.TypeInt,
		Optional:    true,
		Description: "The base permission bits for deny to extend.",
	}
	schemaMap["allow_bits"] = &schema.Schema{
		Type:        schema.TypeInt,
		Computed:    true,
		Description: "The allow permission bits.",
	}
	schemaMap["deny_bits"] = &schema.Schema{
		Type:        schema.TypeInt,
		Computed:    true,
		Description: "The deny permission bits.",
	}
	for k := range permissions {
		schemaMap[k] = &schema.Schema{
			Optional:     true,
			Type:         schema.TypeString,
			Default:      "unset",
			Description:  fmt.Sprintf("The value to set for the `%s` permission bit. Must be `allow`, `unset`, or `deny`. (default `unset`)", k),
			ValidateFunc: validation.StringInSlice([]string{"allow", "unset", "deny"}, false),
		}
	}

	return &schema.Resource{
		ReadContext: dataSourceDiscordPermissionRead,
		Description: "A simple helper to get computed bit total of a list of permissions.",
		Schema:      schemaMap,
	}
}

func dataSourceDiscordPermissionRead(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	var allowBits int64
	var denyBits int64
	for perm, bit := range permissions {
		switch d.Get(perm).(string) {
		case "allow":
			allowBits |= bit
		case "deny":
			denyBits |= bit
		}
	}

	d.SetId(strconv.Itoa(Hashcode(fmt.Sprintf("%d:%d", allowBits, denyBits))))
	d.Set("allow_bits", allowBits|(int64(d.Get("allow_extends").(int))))
	d.Set("deny_bits", denyBits|(int64(d.Get("deny_extends").(int))))

	return diags
}
