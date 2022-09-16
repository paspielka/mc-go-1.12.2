package transaction

import (
	. "github.com/edouard127/mc-go-1.12.2/struct"
)

func UpdateVelocity(g *Game, entityID int32, velocity Vector3) {
	e := g.World.Entities[entityID]
	e.SetPosition(e.Position.Add(velocity))
}
