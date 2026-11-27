package provider

// A deleted object can keep its name for a while: an SSH key's slug until every zone confirmed the delete,
// a node pool's name until its workers are gone. A replacement deletes first, so its create would be
// refused after the old object is gone.

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

// reservedLimit is how long a create is sent again for a value this run freed.
var reservedLimit = 15 * time.Minute

// reservedValues are the values of objects this run deleted.
type reservedValues struct {
	mu     sync.Mutex
	values map[string]bool
}

// reservedKey is the value in the place it is unique: its resource, its team and the path it is created under.
func reservedKey(b *apiBinding, vars, attrs map[string]string, value string) string {
	parts := []string{b.Resource, value, "team=" + attrs["team_id"]}
	for name, v := range vars {
		if b.Create != nil && strings.Contains(b.Create.Path, "{"+name+"}") {
			parts = append(parts, name+"="+v)
		}
	}
	sort.Strings(parts[3:])
	return strings.Join(parts, "|")
}

// reservedOf are the keys of the values the attributes hold, by value.
func reservedOf(b *apiBinding, vars map[string]string, attrs map[string]interface{}) map[string]string {
	values := itemStrings(attrs[b.ReservedField])
	if one, ok := attrs[b.ReservedField].(string); ok {
		values = []string{one}
	}
	team := map[string]string{"team_id": stringifyScalar(attrs["team_id"])}
	keys := map[string]string{}
	for _, value := range values {
		keys[value] = reservedKey(b, vars, team, value)
	}
	return keys
}

// Freed records that this run deleted the object that held these attributes.
func (c *Client) Freed(b *apiBinding, vars map[string]string, attrs map[string]interface{}) {
	c.reserved.mu.Lock()
	defer c.reserved.mu.Unlock()
	if c.reserved.values == nil {
		c.reserved.values = map[string]bool{}
	}
	for _, key := range reservedOf(b, vars, attrs) {
		c.reserved.values[key] = true
	}
}

// taken forgets what a create holds again: the next refusal of it is of this run's own object.
func (c *Client) taken(b *apiBinding, vars map[string]string, attrs map[string]interface{}) {
	c.reserved.mu.Lock()
	defer c.reserved.mu.Unlock()
	for _, key := range reservedOf(b, vars, attrs) {
		delete(c.reserved.values, key)
	}
}

// callCreate sends the create, again while the API refuses exactly the value this run freed. A value held
// by anything else fails at once.
func (c *Client) callCreate(ctx context.Context, b *apiBinding, vars map[string]string, attrs map[string]interface{}) (map[string]interface{}, int, error) {
	started := time.Now()
	for attempt := 0; ; attempt++ {
		out, status, err := c.call(ctx, b, b.Create, vars, attrs)
		if err == nil && b.ReservedField != "" {
			c.taken(b, vars, attrs)
		}
		if err == nil || !c.stillReserved(b, vars, attrs, err) || time.Since(started) > reservedLimit {
			return out, status, err
		}
		delay := waitDelays[len(waitDelays)-1]
		if attempt < len(waitDelays) {
			delay = waitDelays[attempt]
		}
		select {
		case <-ctx.Done():
			return out, status, err
		case <-time.After(delay):
		}
	}
}

// stillReserved is true for a 422 that refuses this one field, each time with the declared message for a
// value this run freed.
func (c *Client) stillReserved(b *apiBinding, vars map[string]string, attrs map[string]interface{}, err error) bool {
	var apiErr *APIError
	if b.ReservedField == "" || !errors.As(err, &apiErr) || apiErr.Status != http.StatusUnprocessableEntity {
		return false
	}
	var body interface{}
	if json.Unmarshal([]byte(apiErr.Body), &body) != nil {
		return false
	}
	details, ok := digPath(body, b.Errors.DetailPaths["details"]).(map[string]interface{})
	messages, _ := details[b.ReservedField].([]interface{})
	if !ok || len(details) != 1 || len(messages) == 0 {
		return false
	}
	c.reserved.mu.Lock()
	defer c.reserved.mu.Unlock()
	freed := map[string]bool{}
	for value, key := range reservedOf(b, vars, attrs) {
		if c.reserved.values[key] {
			freed[strings.ReplaceAll(b.ReservedMessage, "{value}", value)] = true
		}
	}
	for _, message := range messages {
		if !freed[stringifyScalar(message)] {
			return false
		}
	}
	return true
}
