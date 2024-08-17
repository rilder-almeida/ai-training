package board

import "fmt"

type playerSet struct {
	Red  Player
	Blue Player
}

// Players represents the set of players that can be used.
var Players = playerSet{
	Red:  newPlayer("RED"),
	Blue: newPlayer("BLUE"),
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
