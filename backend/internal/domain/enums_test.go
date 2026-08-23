package domain

import "testing"

func TestRoleGates(t *testing.T) {
	if !RoleOwner.CanManageMembers() || RoleCaregiver.CanManageMembers() || RoleViewer.CanWrite() {
		t.Fatal("rbac gates")
	}
	if !RoleCaregiver.CanWrite() || !RoleOwner.CanDeletePet() || RoleCaregiver.CanDeletePet() {
		t.Fatal("write/delete gates")
	}
}

func TestEventMapsToReminder(t *testing.T) {
	k, ok := EventVaccine.ReminderKind()
	if !ok || k != ReminderVaccine {
		t.Fatal("vaccine map")
	}
	if _, ok := EventSymptom.ReminderKind(); ok {
		t.Fatal("symptom should not recalc vaccine cycle")
	}
}

func TestFrozenEnums(t *testing.T) {
	for _, v := range []interface{ Valid() bool }{
		RoleOwner, CheckinFeed, SlotNight, EventDeworm, SeveritySevere,
		ReminderDeworm, ChannelWecom, ExpenseToy, MediaPhoto, SpeciesCat, GenderFemale,
	} {
		if !v.Valid() {
			t.Fatalf("frozen enum invalid: %v", v)
		}
	}
}
