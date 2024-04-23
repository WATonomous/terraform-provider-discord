package discord

import (
	"fmt"
	"strings"

	"github.com/andersfylling/disgord"
)

func parseTwoIds(id string) (string, string, error) {
	parts := strings.SplitN(id, ":", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected attribute1:attribute2", id)
	}

	return parts[0], parts[1], nil
}

// Helper function for generating a two part ID
func generateTwoPartId(one string, two string) string {
	return fmt.Sprintf("%s:%s", one, two)
}

func parseThreeIds(id string) (string, string, string, error) {
	parts := strings.SplitN(id, ":", 3)

	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", fmt.Errorf("unexpected format of ID (%s), expected attribute1:attribute2:attriburte3", id)
	}

	return parts[0], parts[1], parts[2], nil
}

func generateThreePartId(one string, two string, three string) string {
	return fmt.Sprintf("%s:%s:%s", one, two, three)
}

func getId(v string) disgord.Snowflake {
	return disgord.ParseSnowflakeString(v)
}

func getMinorId(v interface{}) disgord.Snowflake {
	str := v.(string)
	if strings.Contains(str, ":") {
		_, secondId, _ := parseTwoIds(str)

		return getId(secondId)
	}

	return getId(v.(string))
}

func getMajorId(v interface{}) disgord.Snowflake {
	str := v.(string)
	if strings.Contains(str, ":") {
		firstId, _, _ := parseTwoIds(str)

		return getId(firstId)
	}

	return getId(v.(string))
}

func getBothIds(v interface{}) (disgord.Snowflake, disgord.Snowflake, error) {
	firstId, secondId, err := parseTwoIds(v.(string))
	if err != nil {
		return 0, 0, err
	}

	return disgord.ParseSnowflakeString(firstId), disgord.ParseSnowflakeString(secondId), nil
}
