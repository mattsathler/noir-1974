package identity

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"noir-1974/apps/api/internal/world"
)

type characterInsertRow struct{}

func (characterInsertRow) Scan(destinations ...any) error {
	*destinations[0].(*string) = "Ada"
	*destinations[1].(*int) = 1
	return nil
}

type characterStoreFixture struct {
	visualJSON string
}

func (fixture *characterStoreFixture) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.NewCommandTag("INSERT 1"), nil
}
func (fixture *characterStoreFixture) QueryRow(_ context.Context, query string, args ...any) pgx.Row {
	if len(args) > 2 {
		fixture.visualJSON, _ = args[2].(string)
	}
	return characterInsertRow{}
}

func TestCharacterVisualIdentityIsPersistedAndReturned(t *testing.T) {
	fixture := &characterStoreFixture{}
	store := NewPostgresStore(fixture)
	want := world.VisualIdentity{
		HairID: "hair_long_wavy_ginger", SkinID: "skin_olive", EyesID: "eyes_green",
		BeardID: "beard_moustache_ginger", FaceID: "face_scar_short", AccessoryIDs: []string{"accessories_glasses_round_brass"},
	}
	created, err := store.createCharacter(context.Background(), "account-id", "Ada", want)
	if err != nil {
		t.Fatalf("createCharacter() error = %v", err)
	}
	if created.DisplayName != "Ada" || created.Visual.HairID != want.HairID {
		t.Fatalf("created character = %+v", created)
	}
	var persisted world.VisualIdentity
	if err := json.Unmarshal([]byte(fixture.visualJSON), &persisted); err != nil {
		t.Fatalf("decode persisted visual JSON: %v", err)
	}
	if persisted.HairID != want.HairID || persisted.SkinID != want.SkinID || persisted.EyesID != want.EyesID || persisted.BeardID != want.BeardID || persisted.FaceID != want.FaceID || len(persisted.AccessoryIDs) != 1 || persisted.AccessoryIDs[0] != want.AccessoryIDs[0] {
		t.Fatalf("persisted visual = %+v, want %+v", persisted, want)
	}
	if err := world.ValidateVisualIdentity(persisted); err != nil {
		t.Fatalf("persisted appearance is not accepted by the catalog: %v", err)
	}
}
