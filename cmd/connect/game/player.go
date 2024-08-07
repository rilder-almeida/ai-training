package game

import "fmt"

type playerSet struct {
	Red  Player
	Blue Player
}

// Players represents the set of players that can be used.
var Players = playerSet{
	Red:  newPlayer("Red"),
	Blue: newPlayer("Blue"),
}

// =============================================================================

// Set of known players.
var players = make(map[string]Player)

// Player represents a player in the system.
type Player struct {
	name string
}

func newPlayer(player string) Player {
	p := Player{player}
	players[player] = p
	return p
}

// IsZero checks of the player is set to its zero value.
func (p Player) IsZero() bool {
	return p.name == ""
}

// String returns the name of the player.
func (p Player) String() string {
	return p.name
}

// Equal provides support for the go-cmp package and testing.
func (p Player) Equal(p2 Player) bool {
	return p.name == p2.name
}

// =============================================================================

// ParsePlayer parses the string value and returns a player if one exists.
func ParsePlayer(value string) (Player, error) {
	player, exists := players[value]
	if !exists {
		return Player{}, fmt.Errorf("invalid player %q", value)
	}

	return player, nil
}

// MustParsePlayer parses the string value and returns a player if one exists. If
// an error occurs the function panics.
func MustParsePlayer(value string) Player {
	role, err := ParsePlayer(value)
	if err != nil {
		panic(err)
	}

	return role
}
