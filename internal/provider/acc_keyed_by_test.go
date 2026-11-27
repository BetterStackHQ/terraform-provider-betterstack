package provider

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// A team member is keyed by e-mail: the read and the delete send ?email=, the role change sends the
// e-mail in its path, and no request ever addresses a member by its u123 id, since the e-mail is
// the one key a member and a pending invitation to them share.

const teamMemberAddr = "betterstack_team_member.test"

// teamMemberMock serves two roles and, like the API, invites by default: a created team member is a
// pending invitation until accepted. A role change sets role_id, and the role name follows from it.
func teamMemberMock(t *testing.T) *MockAPI {
	mock := newMockAPI(t)
	for id, role := range map[string]string{"201": "responder", "202": "member"} {
		if err := mock.SeedRecord("role", map[string]string{}, id, map[string]interface{}{"name": role, "role": role}); err != nil {
			t.Fatal(err)
		}
	}
	mock.InjectOnCreate("team_member", map[string]interface{}{"invited_at": "2026-11-21T10:00:00.000Z"})
	mock.AfterWrite("team_member", func(attrs map[string]interface{}) {
		switch {
		case attrs["role"] == "member" || attrs["role_id"] == "202":
			attrs["role"], attrs["role_id"] = "member", "202"
		default:
			attrs["role"], attrs["role_id"] = "responder", "201"
		}
	})
	return mock
}

func teamMemberConfig(url, role string, protected bool) string {
	return providerConfig(url) + fmt.Sprintf("resource \"betterstack_team_member\" \"test\" {\n email = \"alice@example.com\"\n role = %q\n deletion_protection = %t\n}\n", role, protected)
}

func acceptInvitation(t *testing.T, mock *MockAPI) {
	if mock.SetFields("team_member", "email", "alice@example.com", map[string]interface{}{
		"invited_at": nil, "first_name": "Alice", "last_name": "Smith", "created_at": "2026-11-21T11:00:00.000Z",
	}) != 1 {
		t.Fatal("no invitation to accept")
	}
}

// requestsWith returns the requests so far that start with this method and path.
func requestsWith(mock *MockAPI, prefix string) []string {
	var out []string
	for _, request := range mock.Requests() {
		if strings.HasPrefix(request, prefix) {
			out = append(out, request)
		}
	}
	return out
}

func noRequestByID(mock *MockAPI) error {
	for _, request := range mock.Requests() {
		if regexp.MustCompile(`/api/v3/team-members/u?[0-9]+`).MatchString(request) {
			return fmt.Errorf("a request addressed a team member by id: %s", request)
		}
	}
	return nil
}

// An accepted invitation reads back as the member with no diff, a role change on the member happens
// in place through the role change by e-mail, the import takes the e-mail, and the member is removed
// only once deletion_protection is off.
func TestAccTeamMemberAcceptedMemberChangesRoleInPlace(t *testing.T) {
	mock := teamMemberMock(t)
	removed := providerConfig(mock.URL()) + "removed {\n from = " + teamMemberAddr + "\n lifecycle {\n destroy = false\n }\n}\n"
	imported := providerConfig(mock.URL()) + "import {\n to = " + teamMemberAddr + "\n id = \"alice@example.com\"\n}\n" +
		"resource \"betterstack_team_member\" \"test\" {\n email = \"alice@example.com\"\n role = \"member\"\n}\n"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: teamMemberConfig(mock.URL(), "responder", true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(teamMemberAddr, "id", "alice@example.com"),
					resource.TestCheckResourceAttr(teamMemberAddr, "invited_at", "2026-11-21T10:00:00.000Z"),
					resource.TestCheckResourceAttr(teamMemberAddr, "role_id", "201"),
				),
			},
			{
				PreConfig: func() { acceptInvitation(t, mock) },
				Config:    teamMemberConfig(mock.URL(), "responder", true),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(teamMemberAddr, plancheck.ResourceActionNoop)},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(teamMemberAddr, "first_name", "Alice"),
					resource.TestCheckNoResourceAttr(teamMemberAddr, "invited_at"),
				),
			},
			{
				Config: teamMemberConfig(mock.URL(), "member", true),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(teamMemberAddr, plancheck.ResourceActionUpdate)},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(teamMemberAddr, "id", "alice@example.com"),
					resource.TestCheckResourceAttr(teamMemberAddr, "role_id", "202"),
					func(*terraform.State) error {
						if got := requestsWith(mock, "POST /api/v3/team-members/alice@example.com/change-role/202"); len(got) != 1 {
							return fmt.Errorf("expected one role change by e-mail, got %v", mock.Requests())
						}
						if got := requestsWith(mock, "DELETE "); len(got) != 0 {
							return fmt.Errorf("a role change removed someone: %v", got)
						}
						return nil
					},
				),
			},
			{Config: providerConfig(mock.URL()), ExpectError: regexp.MustCompile(`Deletion protection is on`)},
			{Config: removed},
			{
				Config: imported,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(teamMemberAddr, plancheck.ResourceActionNoop)},
				},
				Check: resource.TestCheckResourceAttr(teamMemberAddr, "deletion_protection", "true"),
			},
			{Config: teamMemberConfig(mock.URL(), "member", false)},
		},
		CheckDestroy: func(*terraform.State) error {
			deletes := requestsWith(mock, "DELETE ")
			if len(deletes) != 1 || deletes[0] != "DELETE /api/v3/team-members?email=alice%40example.com" {
				return fmt.Errorf("expected one delete by e-mail, got %v", deletes)
			}
			return noRequestByID(mock)
		},
	})
}

// No API call changes an invitation's role, so the apply cancels the invitation and sends a new one
// with the new role.
func TestAccTeamMemberInvitationRoleChangeInvitesAgain(t *testing.T) {
	mock := teamMemberMock(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories,
		Steps: []resource.TestStep{
			{Config: teamMemberConfig(mock.URL(), "responder", true)},
			{
				Config: teamMemberConfig(mock.URL(), "member", true),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(teamMemberAddr, plancheck.ResourceActionUpdate)},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(teamMemberAddr, "id", "alice@example.com"),
					resource.TestCheckResourceAttr(teamMemberAddr, "role_id", "202"),
					func(*terraform.State) error {
						if got := requestsWith(mock, "POST /api/v3/team-members"); len(got) != 2 {
							return fmt.Errorf("expected the first and the new invitation, got %v", got)
						}
						if got := requestsWith(mock, "DELETE /api/v3/team-members?email=alice%40example.com"); len(got) != 1 {
							return fmt.Errorf("expected the invitation to be cancelled by e-mail, got %v", mock.Requests())
						}
						if got := requestsWith(mock, "POST /api/v3/team-members/alice@example.com/change-role/"); len(got) != 0 {
							return fmt.Errorf("an invitation got a role change: %v", got)
						}
						return nil
					},
				),
			},
			{Config: teamMemberConfig(mock.URL(), "member", false)},
		},
		CheckDestroy: func(*terraform.State) error { return noRequestByID(mock) },
	})
}

// The API answers an invitation to someone already on the team with that member and changes
// nothing; the create refuses to take them into state and points to the import.
func TestAccTeamMemberRefusesToAdoptAnExistingMember(t *testing.T) {
	mock := teamMemberMock(t)
	mock.FailWrites("team_member", 200, `{"message":"User is already a member of the team","data":{"id":"7","type":"team_member","attributes":{"email":"alice@example.com","role":"admin"}}}`)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      teamMemberConfig(mock.URL(), "responder", true),
				ExpectError: regexp.MustCompile(`(?s)alice@example.com already exists, so nothing was created.*terraform import`),
			},
		},
	})
}

// Someone already in the organization joins the team at once: the create answers with no object,
// and the provider reads the member it made by e-mail.
func TestAccTeamMemberReadsAMemberTheCreateAddedAtOnce(t *testing.T) {
	mock := teamMemberMock(t)
	if err := mock.SeedRecord("team_member", map[string]string{}, "7", map[string]interface{}{"email": "alice@example.com", "role": "responder", "role_id": "201", "first_name": "Alice"}); err != nil {
		t.Fatal(err)
	}
	mock.FailWrites("team_member", 201, `{"message":"Invitation sent","data":null}`)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: teamMemberConfig(mock.URL(), "responder", false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(teamMemberAddr, "id", "alice@example.com"),
					resource.TestCheckResourceAttr(teamMemberAddr, "first_name", "Alice"),
				),
			},
			{Config: providerConfig(mock.URL()) + "removed {\n from = " + teamMemberAddr + "\n lifecycle {\n destroy = false\n }\n}\n"},
		},
	})
}

// The API matches an e-mail without letter case: a change of case only updates the spelling in
// state and never replaces (removes) the member.
func TestAccTeamMemberEmailCaseChangeIsNoReplacement(t *testing.T) {
	mock := teamMemberMock(t)
	spelled := func(email string) string {
		return providerConfig(mock.URL()) + fmt.Sprintf("resource \"betterstack_team_member\" \"test\" {\n email = %q\n role = \"responder\"\n}\n", email)
	}
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories,
		Steps: []resource.TestStep{
			{Config: spelled("alice@example.com")},
			{
				Config: spelled("Alice@Example.com"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(teamMemberAddr, plancheck.ResourceActionUpdate)},
				},
			},
			{Config: spelled("Alice@Example.com"), PlanOnly: true},
			{Config: providerConfig(mock.URL()) + "removed {\n from = " + teamMemberAddr + "\n lifecycle {\n destroy = false\n }\n}\n"},
		},
		CheckDestroy: func(*terraform.State) error {
			if got := requestsWith(mock, "DELETE "); len(got) != 0 {
				return fmt.Errorf("a case change removed someone: %v", got)
			}
			return nil
		},
	})
}

// role and role_id that name different roles fail the update instead of leaving state that
// disagrees with the server.
func TestAccTeamMemberRefusesRoleAndRoleIDThatDisagree(t *testing.T) {
	mock := teamMemberMock(t)
	both := func(role, roleID string) string {
		return providerConfig(mock.URL()) + fmt.Sprintf("resource \"betterstack_team_member\" \"test\" {\n email = \"alice@example.com\"\n role = %q\n role_id = %q\n deletion_protection = false\n}\n", role, roleID)
	}
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {},
				Config:    teamMemberConfig(mock.URL(), "responder", false),
			},
			{
				PreConfig:   func() { acceptInvitation(t, mock) },
				Config:      both("member", "201"),
				ExpectError: regexp.MustCompile(`role "member" and role_id 201 name different things`),
			},
		},
	})
}
