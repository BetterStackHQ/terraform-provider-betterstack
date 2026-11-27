package provider

// Waiting for a write to settle. Some APIs accept a create, update or delete at once (202) and do
// the work afterwards: a server is `starting` until it runs. The API description says, per write,
// which read to poll and what success and failure look like (the waits the CLI's --wait follows);
// the binding carries them as JSON in Waits. Nothing here knows a product.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

type waitBinding struct {
	Source string      `json:"source"`
	Path   []string    `json:"path"`
	Value  interface{} `json:"value"`
}

type waitPredicate struct {
	All      []waitPredicate `json:"all"`
	Any      []waitPredicate `json:"any"`
	Not      *waitPredicate  `json:"not"`
	Path     []string        `json:"path"`
	Expected *waitBinding    `json:"expected"`
	In       []interface{}   `json:"in"`
	AtLeast  *waitBinding    `json:"at_least"`
}

type waitPolicy struct {
	Bindings      map[string]waitBinding `json:"bindings"`
	Success       *waitPredicate         `json:"success"`
	Failure       []waitPredicate        `json:"failure"`
	When          *waitPredicate         `json:"when"`
	Messages      []string               `json:"messages"`
	PendingCodes  []string               `json:"pending_codes"`
	Disappearance bool                   `json:"disappearance"`
}

// waitScope is the data a binding may name: the request body, the path variables and the response.
type waitScope struct {
	request map[string]interface{}
	params  map[string]interface{}
	result  map[string]interface{}
	before  map[string]interface{}
}

// waitError is a write the API accepted whose work didn't finish: Object is the last thing read.
type waitError struct {
	Object  map[string]interface{}
	Message string
}

func (e *waitError) Error() string { return e.Message }

// waitDelays space the polls; waitLimit is how long a write may take before the provider stops
// watching. The work itself carries on.
var (
	waitDelays = []time.Duration{time.Second, 2 * time.Second, 3 * time.Second, 5 * time.Second}
	waitLimit  = time.Hour
)

func waitWalk(root map[string]interface{}, path []string) (interface{}, bool) {
	var value interface{} = root
	for _, key := range path {
		object, ok := value.(map[string]interface{})
		if !ok {
			return nil, false
		}
		if value, ok = object[key]; !ok {
			return nil, false
		}
	}
	return value, root != nil
}

func (s waitScope) bind(b *waitBinding) (interface{}, bool) {
	switch b.Source {
	case "literal":
		return b.Value, true
	case "request":
		return waitWalk(s.request, b.Path)
	case "params":
		return waitWalk(s.params, b.Path)
	case "result":
		return waitWalk(s.result, b.Path)
	case "before":
		return waitWalk(s.before, b.Path)
	}
	return nil, false
}

// waitEqual compares as the JSON reads: 1 and "1", true and "true" are the same value.
func waitEqual(a, b interface{}) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return stringifyScalar(a) == stringifyScalar(b)
}

func waitNumber(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case json.Number:
		f, err := n.Float64()
		return f, err == nil
	case float64:
		return n, true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	}
	return 0, false
}

// holds says whether the predicate holds for an observed object. Anything missing is "not yet".
func (p *waitPredicate) holds(root map[string]interface{}, scope waitScope) bool {
	if len(p.All) > 0 {
		for i := range p.All {
			if !p.All[i].holds(root, scope) {
				return false
			}
		}
		return true
	}
	if len(p.Any) > 0 {
		for i := range p.Any {
			if p.Any[i].holds(root, scope) {
				return true
			}
		}
		return false
	}
	if p.Not != nil {
		return !p.Not.holds(root, scope)
	}
	value, ok := waitWalk(root, p.Path)
	if !ok {
		return false
	}
	if p.Expected != nil {
		wanted, ok := scope.bind(p.Expected)
		return ok && waitEqual(value, wanted)
	}
	if len(p.In) > 0 {
		for _, allowed := range p.In {
			if waitEqual(value, allowed) {
				return true
			}
		}
		return false
	}
	if p.AtLeast != nil {
		bound, ok := scope.bind(p.AtLeast)
		a, okA := waitNumber(value)
		b, okB := waitNumber(bound)
		return ok && okA && okB && a >= b
	}
	return false
}

// settle follows the wait the binding declares for slot after the write answered result. It
// returns the settled object, or nil when a delete's object is gone.
func (c *Client) settle(ctx context.Context, b *apiBinding, slot string, vars map[string]string, request, result map[string]interface{}) (map[string]interface{}, error) {
	return c.settleOn(ctx, b, b.Waits[slot], vars, request, result, nil)
}

// settleOn follows one wait; before is the object as it was read before the change, when it was.
func (c *Client) settleOn(ctx context.Context, b *apiBinding, raw string, vars map[string]string, request, result, before map[string]interface{}) (map[string]interface{}, error) {
	if raw == "" {
		return result, nil
	}
	var policy waitPolicy
	if err := json.Unmarshal([]byte(raw), &policy); err != nil {
		return result, fmt.Errorf("%s: reading the wait the API describes: %w", b.Resource, err)
	}
	params := map[string]interface{}{}
	for k, v := range vars {
		params[k] = v
	}
	scope := waitScope{request: request, params: params, result: result, before: before}
	if policy.When != nil && !policy.When.holds(request, scope) {
		return result, nil
	}
	observe := map[string]string{}
	for k, v := range vars {
		observe[k] = v
	}
	for name, binding := range policy.Bindings {
		binding := binding
		value, ok := scope.bind(&binding)
		if !ok || value == nil {
			// A delete that answers without a body leaves the object at the address it was sent to.
			if _, known := vars[name]; known {
				continue
			}
			return result, &waitError{Object: result, Message: fmt.Sprintf("%s: the API accepted the change, but its response doesn't say which object to watch (%s).", b.Resource, name)}
		}
		observe[name] = stringifyScalar(value)
	}
	started := time.Now()
	last := result
	for attempt := 0; ; attempt++ {
		row, found, err := c.Get(ctx, b, observe)
		if err != nil {
			var apiErr *APIError
			if !(errors.As(err, &apiErr) && waitPending(apiErr, policy.PendingCodes)) {
				return last, err
			}
		} else if !found {
			if policy.Disappearance {
				return nil, nil
			}
			return last, &waitError{Object: last, Message: fmt.Sprintf("%s was not found while waiting for it to be ready.", b.Resource)}
		} else {
			last = row
			if policy.Disappearance && b.gone(row) {
				return nil, nil
			}
			if policy.Success != nil && policy.Success.holds(row, scope) {
				return row, nil
			}
			if policy.Success == nil && !policy.Disappearance {
				return row, nil
			}
			for i := range policy.Failure {
				if policy.Failure[i].holds(row, scope) {
					return row, &waitError{Object: row, Message: fmt.Sprintf("%s ended up %s instead of ready.%s", b.Resource, waitState(row), waitMessages(row, policy.Messages))}
				}
			}
		}
		if time.Since(started) > waitLimit {
			return last, &waitError{Object: last, Message: fmt.Sprintf("%s wasn't ready after %s: it is still %s. The work carries on, and the next apply picks it up.%s", b.Resource, waitLimit, waitState(last), waitMessages(last, policy.Messages))}
		}
		delay := waitDelays[len(waitDelays)-1]
		if attempt < len(waitDelays) {
			delay = waitDelays[attempt]
		}
		select {
		case <-ctx.Done():
			return last, &waitError{Object: last, Message: fmt.Sprintf("%s: stopped waiting (%s). The work carries on, and the next apply picks it up.", b.Resource, ctx.Err())}
		case <-time.After(delay):
		}
	}
}

func waitState(row map[string]interface{}) string {
	for _, key := range []string{"status", "state"} {
		if v, ok := row[key].(string); ok && v != "" {
			return v
		}
	}
	return "in progress"
}

func waitMessages(row map[string]interface{}, keys []string) string {
	var out []string
	for _, key := range keys {
		if v, ok := row[key].(string); ok && v != "" {
			out = append(out, v)
		}
	}
	if len(out) == 0 {
		return ""
	}
	return " " + strings.Join(out, " ")
}

// waitPending says the read failed with a code the description marks as "not yet".
func waitPending(err *APIError, codes []string) bool {
	if len(codes) == 0 {
		return false
	}
	for _, code := range codes {
		if strings.Contains(err.Body, `"`+code+`"`) {
			return true
		}
	}
	return false
}
